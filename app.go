package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hclareth7/ais/internal/config"
	"github.com/hclareth7/ais/internal/input"
	"github.com/hclareth7/ais/internal/llm"
	"github.com/hclareth7/ais/internal/platform"
	"github.com/hclareth7/ais/internal/scanner"
	"github.com/hclareth7/ais/internal/types"
	"github.com/hclareth7/ais/internal/watcher"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx          context.Context
	rootPath     string
	cfgMgr       *config.Manager
	watcher      *watcher.Watcher
	pipeReader   *input.PipeReader
	streamCancel context.CancelFunc // nil when no stream active
	streamMu     sync.Mutex         // guards streamCancel
}

func NewApp(rootPath string) *App {
	return &App{
		rootPath: rootPath,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.cfgMgr = config.NewManager()
	if err := a.cfgMgr.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load config: %v\n", err)
	}

	if a.rootPath != "" {
		a.cfgMgr.Update(func(c *config.Config) {
			c.LastOpenedPath = a.rootPath
		})
	}

	w, err := watcher.New(a.rootPath, func(event watcher.WatchEvent) {
		if event.IsCreate {
			wailsRuntime.EventsEmit(a.ctx, "file:created", event.Path)
		} else {
			wailsRuntime.EventsEmit(a.ctx, "file:changed", event.Path)
		}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to start watcher: %v\n", err)
	} else {
		a.watcher = w
		go a.watcher.Start()
	}
}

func (a *App) shutdown(ctx context.Context) {
	// Cancel any active stream to prevent goroutine leaks.
	a.streamMu.Lock()
	if a.streamCancel != nil {
		a.streamCancel()
		a.streamCancel = nil
	}
	a.streamMu.Unlock()

	if a.pipeReader != nil {
		a.pipeReader.Stop()
		a.pipeReader = nil
	}
	if a.watcher != nil {
		a.watcher.Stop()
	}
	if a.cfgMgr != nil {
		a.cfgMgr.Save()
	}
}

func (a *App) GetFileTree() (*types.FileNode, error) {
	return scanner.ScanDirectory(a.rootPath)
}

func (a *App) ReadFile(relativePath string) (string, error) {
	absPath, err := filepath.Abs(filepath.Join(a.rootPath, relativePath))
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	if !strings.HasPrefix(absPath, a.rootPath+string(os.PathSeparator)) && absPath != a.rootPath {
		return "", fmt.Errorf("path outside root: %s", relativePath)
	}
	return scanner.ReadFileContent(absPath)
}

func (a *App) GetRootPath() string {
	return a.rootPath
}

func (a *App) OpenFolder() (string, error) {
	path, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Open Folder",
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	return path, a.SetRootPath(path)
}

func (a *App) SetRootPath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path not found: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	if a.watcher != nil {
		a.watcher.Stop()
	}

	a.rootPath = absPath

	a.cfgMgr.Update(func(c *config.Config) {
		c.LastOpenedPath = absPath
		recent := []string{absPath}
		for _, p := range c.RecentPaths {
			if p != absPath && len(recent) < 10 {
				recent = append(recent, p)
			}
		}
		c.RecentPaths = recent
	})

	w, err := watcher.New(absPath, func(event watcher.WatchEvent) {
		if event.IsCreate {
			wailsRuntime.EventsEmit(a.ctx, "file:created", event.Path)
		} else {
			wailsRuntime.EventsEmit(a.ctx, "file:changed", event.Path)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to start watcher: %w", err)
	}
	a.watcher = w
	go a.watcher.Start()

	return nil
}

func (a *App) GetConfig() config.Config {
	return a.cfgMgr.Get()
}

func (a *App) UpdateConfig(cfg config.Config) error {
	a.cfgMgr.Update(func(c *config.Config) {
		*c = cfg
	})
	return a.cfgMgr.Save()
}

func (a *App) GetTheme() string {
	return a.cfgMgr.Get().Theme
}

func (a *App) SetTheme(theme string) error {
	a.cfgMgr.Update(func(c *config.Config) {
		c.Theme = theme
	})
	return a.cfgMgr.Save()
}

func (a *App) SetCornerRadius(radius float64) {
	platform.UpdateCornerRadius(radius)
}

// StartStream begins streaming a Claude API response. It opens an HTTP connection
// to the Anthropic API and emits llm:chunk, llm:done, and llm:error events as
// content arrives. Only one stream may be active at a time.
func (a *App) StartStream(prompt string) error {
	a.streamMu.Lock()
	if a.streamCancel != nil {
		a.streamMu.Unlock()
		return fmt.Errorf("a stream is already active")
	}

	apiKey, err := llm.GetAPIKey()
	if err != nil {
		a.streamMu.Unlock()
		return fmt.Errorf("no API key configured: %w", err)
	}

	// Resolve model from config; fall back to Sonnet (quality-first default).
	model := a.cfgMgr.Get().SelectedModel
	if model == "" {
		model = llm.ModelSonnet
	}

	client := llm.NewClient(apiKey, model)
	ctx, cancel := context.WithCancel(a.ctx)
	a.streamCancel = cancel
	a.streamMu.Unlock()

	go func() {
		defer func() {
			a.streamMu.Lock()
			a.streamCancel = nil
			a.streamMu.Unlock()
		}()

		req := llm.StreamRequest{Prompt: prompt, Model: model}
		err := client.Stream(ctx, req, func(chunk llm.StreamChunk) {
			if chunk.Done {
				wailsRuntime.EventsEmit(a.ctx, "llm:done", chunk)
			} else {
				wailsRuntime.EventsEmit(a.ctx, "llm:chunk", chunk)
			}
		})
		if err != nil {
			streamErr, ok := err.(*llm.StreamError)
			if ok {
				wailsRuntime.EventsEmit(a.ctx, "llm:error", map[string]string{
					"code":    streamErr.Code,
					"message": streamErr.Message,
				})
			} else {
				// Log the full error for debugging but emit a generic message
				// to avoid leaking SDK internals to the frontend.
				fmt.Fprintf(os.Stderr, "llm: unexpected error: %v\n", err)
				wailsRuntime.EventsEmit(a.ctx, "llm:error", map[string]string{
					"code":    llm.ErrCodeAPI,
					"message": "unexpected error",
				})
			}
		}
	}()

	return nil
}

// CancelStream cancels the currently active stream by cancelling the context.
// Safe to call when no stream is active.
func (a *App) CancelStream() error {
	a.streamMu.Lock()
	defer a.streamMu.Unlock()
	if a.streamCancel == nil {
		return nil // no-op when no stream active
	}
	a.streamCancel()
	a.streamCancel = nil
	return nil
}

// SetAPIKey stores the Anthropic API key in secure storage (OS keychain with
// fallback to credentials file). The key is never stored in config.json.
func (a *App) SetAPIKey(key string) error {
	return llm.SetAPIKey(key)
}

// HasAPIKey returns whether an API key is configured in any storage location.
// It never exposes the key value — only a boolean status.
func (a *App) HasAPIKey() bool {
	return llm.HasAPIKey()
}

// DeleteAPIKey removes the API key from all storage locations.
func (a *App) DeleteAPIKey() error {
	return llm.DeleteAPIKey()
}

// GetAvailableModels returns the list of supported Claude models, ordered
// by cost (cheapest first).
func (a *App) GetAvailableModels() []string {
	return llm.AvailableModels
}

// StartPipe creates a named pipe (FIFO) and begins reading from it. Each line
// received on the pipe is emitted as a "pipe:data" event to the frontend.
// Only one pipe may be active at a time. Returns the pipe file path.
func (a *App) StartPipe(path string) (string, error) {
	if a.pipeReader != nil {
		return "", fmt.Errorf("pipe already active")
	}

	reader, err := input.New(path, func(text string) {
		wailsRuntime.EventsEmit(a.ctx, "pipe:data", text)
	})
	if err != nil {
		return "", fmt.Errorf("start pipe: %w", err)
	}

	a.pipeReader = reader
	go reader.Start(a.ctx)

	return reader.Path(), nil
}

// StopPipe stops the active pipe reader and removes the FIFO file.
// Safe to call when no pipe is active.
func (a *App) StopPipe() error {
	if a.pipeReader != nil {
		a.pipeReader.Stop()
		a.pipeReader = nil
	}
	return nil
}
