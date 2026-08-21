//go:build !windows

package input

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func tempFIFOPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "test.fifo")
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "explicit path creates FIFO",
			path:    filepath.Join(t.TempDir(), "explicit.fifo"),
			wantErr: false,
		},
		{
			name:    "invalid path returns error",
			path:    "/nonexistent/dir/test.fifo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := New(tt.path, func(string) {})
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer os.Remove(tt.path)

			// Verify FIFO was created.
			info, err := os.Stat(tt.path)
			if err != nil {
				t.Fatalf("stat FIFO: %v", err)
			}
			if info.Mode()&os.ModeNamedPipe == 0 {
				t.Errorf("expected named pipe, got mode %v", info.Mode())
			}

			if reader.Path() != tt.path {
				t.Errorf("Path() = %q, want %q", reader.Path(), tt.path)
			}
		})
	}
}

func TestNewDefaultPath(t *testing.T) {
	reader, err := New("", func(string) {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer reader.Stop()

	expected := fmt.Sprintf("/tmp/ais-%d.fifo", os.Getpid())
	if reader.Path() != expected {
		t.Errorf("Path() = %q, want %q", reader.Path(), expected)
	}
}

func TestReadFromPipe(t *testing.T) {
	path := tempFIFOPath(t)

	var mu sync.Mutex
	var received []string

	reader, err := New(path, func(text string) {
		mu.Lock()
		received = append(received, text)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("create pipe reader: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go reader.Start(ctx)

	// Give the goroutine time to open the pipe.
	time.Sleep(50 * time.Millisecond)

	// Write to the pipe.
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open pipe for writing: %v", err)
	}
	fmt.Fprintln(f, "hello world")
	fmt.Fprintln(f, "second line")
	f.Close()

	// Allow time for data to be read.
	time.Sleep(100 * time.Millisecond)

	cancel()
	// Wait for Stop to complete its cleanup.
	reader.Stop()

	mu.Lock()
	defer mu.Unlock()

	if len(received) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d: %v", len(received), received)
	}

	if !strings.Contains(received[0], "hello world") {
		t.Errorf("first chunk = %q, want it to contain 'hello world'", received[0])
	}
	if !strings.Contains(received[1], "second line") {
		t.Errorf("second chunk = %q, want it to contain 'second line'", received[1])
	}
}

func TestMultipleWrites(t *testing.T) {
	path := tempFIFOPath(t)

	var mu sync.Mutex
	var received []string

	reader, err := New(path, func(text string) {
		mu.Lock()
		received = append(received, text)
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("create pipe reader: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go reader.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// Write multiple separate messages.
	for i := range 5 {
		f, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			t.Fatalf("open pipe for writing (iteration %d): %v", i, err)
		}
		fmt.Fprintf(f, "message %d\n", i)
		f.Close()
		time.Sleep(20 * time.Millisecond)
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	reader.Stop()

	mu.Lock()
	defer mu.Unlock()

	if len(received) < 5 {
		t.Errorf("expected at least 5 messages, got %d: %v", len(received), received)
	}
}

func TestStopCleansUp(t *testing.T) {
	path := tempFIFOPath(t)

	reader, err := New(path, func(string) {})
	if err != nil {
		t.Fatalf("create pipe reader: %v", err)
	}

	// Verify the FIFO exists before Stop.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("FIFO should exist before Stop: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go reader.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	cancel()
	reader.Stop()

	// Verify the FIFO is removed after Stop.
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("FIFO should be removed after Stop")
	}
}

func TestStopIdempotent(t *testing.T) {
	path := tempFIFOPath(t)

	reader, err := New(path, func(string) {})
	if err != nil {
		t.Fatalf("create pipe reader: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go reader.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	cancel()

	// Calling Stop multiple times should not panic.
	reader.Stop()
	reader.Stop()
	reader.Stop()
}

func TestContextCancellation(t *testing.T) {
	path := tempFIFOPath(t)

	reader, err := New(path, func(string) {})
	if err != nil {
		t.Fatalf("create pipe reader: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- reader.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Start returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Start did not return after context cancellation")
	}

	reader.Stop()
}
