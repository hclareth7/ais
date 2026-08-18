package input

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"
)

// PipeReader reads from a named pipe (FIFO) and invokes a callback for each
// chunk of data received. Only supported on Unix systems (Linux, macOS).
type PipeReader struct {
	path        string
	cancel      context.CancelFunc
	done        chan struct{}
	once        sync.Once
	mu          sync.Mutex // guards currentFile
	currentFile *os.File   // the currently open FIFO file descriptor
	callback    func(string)
}

// New creates a named pipe at the given path and returns a PipeReader.
// If path is empty, a default path of /tmp/ais-{pid}.fifo is used.
// The callback is invoked with each line of data read from the pipe.
func New(path string, callback func(string)) (*PipeReader, error) {
	if path == "" {
		path = fmt.Sprintf("/tmp/ais-%d.fifo", os.Getpid())
	}

	// Remove existing pipe if any, ignore errors for non-existent files.
	os.Remove(path)

	if err := syscall.Mkfifo(path, 0666); err != nil {
		return nil, fmt.Errorf("create pipe: %w", err)
	}

	return &PipeReader{
		path:     path,
		done:     make(chan struct{}),
		callback: callback,
	}, nil
}

// setFile stores the current file handle under the mutex.
func (p *PipeReader) setFile(f *os.File) {
	p.mu.Lock()
	p.currentFile = f
	p.mu.Unlock()
}

// closeCurrentFile closes the active file descriptor (if any) to unblock
// a pending Read syscall. Safe to call concurrently.
func (p *PipeReader) closeCurrentFile() {
	p.mu.Lock()
	if p.currentFile != nil {
		p.currentFile.Close()
		p.currentFile = nil
	}
	p.mu.Unlock()
}

// Start begins reading from the pipe in a blocking loop.
// It reads until Stop() is called or the context is cancelled.
// Call this in a goroutine.
func (p *PipeReader) Start(ctx context.Context) error {
	ctx, p.cancel = context.WithCancel(ctx)
	defer close(p.done)

	// Spawn a watchdog that closes the file descriptor when the context is
	// cancelled, which unblocks any pending Read syscall on the FIFO.
	go func() {
		<-ctx.Done()
		p.closeCurrentFile()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// Open with O_RDWR so the read side does not receive EOF when no
		// writer is connected. This is a standard pattern for Go FIFO reading.
		f, err := os.OpenFile(p.path, os.O_RDWR, os.ModeNamedPipe)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return fmt.Errorf("open pipe: %w", err)
			}
		}

		p.setFile(f)

		// Check again after open — context may have been cancelled while
		// we were blocked on OpenFile.
		select {
		case <-ctx.Done():
			f.Close()
			return nil
		default:
		}

		scanner := bufio.NewScanner(f)
		// 10MB max buffer, matching the file size limit in internal/scanner.
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				f.Close()
				return nil
			default:
			}
			p.callback(scanner.Text() + "\n")
		}

		p.setFile(nil)
		f.Close()
	}
}

// Path returns the FIFO file path.
func (p *PipeReader) Path() string {
	return p.path
}

// Stop stops the reader and removes the pipe file. Safe to call multiple times.
// Also safe to call when Start() was never invoked.
func (p *PipeReader) Stop() {
	p.once.Do(func() {
		if p.cancel != nil {
			p.cancel()
			// The watchdog goroutine in Start() closes the file descriptor
			// when the context is cancelled, which unblocks the Read.
			<-p.done
		}
		os.Remove(p.path)
	})
}
