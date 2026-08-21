package main

import (
	"context"
	"fmt"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hclareth7/ais/internal/config"
	"github.com/hclareth7/ais/internal/highlights"
	"github.com/hclareth7/ais/internal/input"
	"github.com/hclareth7/ais/internal/llm"
	"github.com/hclareth7/ais/internal/platform"
	"github.com/hclareth7/ais/internal/scanner"
	"github.com/hclareth7/ais/internal/types"
	"github.com/hclareth7/ais/internal/watcher"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// UISettings holds UI-related configuration that can be saved in bulk.
type UISettings struct {
	ZoomLevel      int    `json:"zoomLevel"`
	Opacity        int    `json:"opacity"`
	ReadingWidth   int    `json:"readingWidth"`
	ReaderRadius   int    `json:"readerRadius"`
	BackgroundMode string `json:"backgroundMode"`
}

type App struct {
	ctx            context.Context
	rootPath       string
	rootMu         sync.RWMutex       // guards rootPath
	initialFile    string
	cfgMgr         *config.Manager
	watcher        *watcher.Watcher
	pipeReader     *input.PipeReader
	highlightStore *highlights.Store
	streamCancel   context.CancelFunc // nil when no stream active
	streamMu       sync.Mutex         // guards streamCancel
}

func NewApp(rootPath, initialFile string) *App {
	return &App{
		rootPath:    rootPath,
		initialFile: initialFile,
	}
}

// getRootPath returns the current root path in a thread-safe manner.
func (a *App) getRootPath() string {
	a.rootMu.RLock()
	defer a.rootMu.RUnlock()
	return a.rootPath
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.cfgMgr = config.NewManager()
	if err := a.cfgMgr.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load config: %v\n", err)
	}

	rootPath := a.getRootPath()

	if rootPath != "" {
		a.cfgMgr.Update(func(c *config.Config) {
			c.LastOpenedPath = rootPath
		})
	}

	a.highlightStore = highlights.NewStore(filepath.Join(rootPath, ".ais", "highlights"))

	w, err := watcher.New(rootPath, func(event watcher.WatchEvent) {
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
	return scanner.ScanDirectory(a.getRootPath())
}

func (a *App) ReadFile(relativePath string) (string, error) {
	root := a.getRootPath()
	absPath, err := filepath.Abs(filepath.Join(root, relativePath))
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	if !strings.HasPrefix(absPath, root+string(os.PathSeparator)) && absPath != root {
		return "", fmt.Errorf("path outside root: %s", relativePath)
	}
	return scanner.ReadFileContent(absPath)
}

// GetRootPath returns the current root directory path. Thread-safe.
func (a *App) GetRootPath() string {
	return a.getRootPath()
}

// GetInitialFile returns the file path passed as a CLI argument, if any.
func (a *App) GetInitialFile() string {
	return a.initialFile
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
	absPath, err = filepath.EvalSymlinks(absPath)
	if err != nil {
		return fmt.Errorf("resolve symlinks: %w", err)
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

	a.rootMu.Lock()
	a.rootPath = absPath
	a.rootMu.Unlock()

	a.highlightStore.SetRoot(filepath.Join(absPath, ".ais", "highlights"))

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
// to the Anthropic API (or Vertex AI) and emits llm:chunk, llm:done, and llm:error events as
// content arrives. Only one stream may be active at a time.
func (a *App) StartStream(prompt string) error {
	a.streamMu.Lock()
	if a.streamCancel != nil {
		a.streamMu.Unlock()
		return fmt.Errorf("a stream is already active")
	}

	cfg := a.cfgMgr.Get()
	model := cfg.SelectedModel
	if model == "" {
		model = llm.ModelSonnet
	}

	provider := cfg.Provider
	if provider == "" {
		provider = llm.ProviderAnthropic
	}

	var client *llm.Client
	if provider == llm.ProviderVertex {
		var err error
		client, err = llm.NewVertexClient(a.ctx, cfg.VertexRegion, cfg.VertexProject, model)
		if err != nil {
			a.streamMu.Unlock()
			return fmt.Errorf("vertex configuration error: %w", err)
		}
	} else {
		apiKey, err := llm.GetAPIKey()
		if err != nil {
			a.streamMu.Unlock()
			return fmt.Errorf("no API key configured: %w", err)
		}
		client = llm.NewClient(apiKey, model)
	}

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

// HasAPIKey returns whether the AI backend is ready. For Anthropic provider,
// checks if an API key is configured. For Vertex AI, checks if project and
// region are set (ADC is handled by the SDK at request time).
func (a *App) HasAPIKey() bool {
	cfg := a.cfgMgr.Get()
	provider := cfg.Provider
	if provider == "" {
		provider = llm.ProviderAnthropic
	}
	if provider == llm.ProviderVertex {
		return cfg.VertexProject != "" && cfg.VertexRegion != ""
	}
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

// GetVertexRegions returns the list of Vertex AI regions where Claude is available.
func (a *App) GetVertexRegions() []string {
	return llm.VertexRegions
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

// validateExternalURL checks that a URL is valid and uses http or https.
func validateExternalURL(urlStr string) error {
	parsed, err := neturl.Parse(urlStr)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("invalid URL")
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %s", scheme)
	}
	return nil
}

// OpenExternalURL opens a URL in the user's default browser. Only http
// and https schemes are allowed; all other schemes are rejected.
func (a *App) OpenExternalURL(url string) error {
	if err := validateExternalURL(url); err != nil {
		return err
	}
	wailsRuntime.BrowserOpenURL(a.ctx, url)
	return nil
}

// SaveUISettings persists UI-related settings (zoom, opacity, reading width,
// reader radius, background mode) to the config file.
func (a *App) SaveUISettings(s UISettings) error {
	a.cfgMgr.Update(func(c *config.Config) {
		c.ZoomLevel = s.ZoomLevel
		c.Opacity = s.Opacity
		c.ReadingWidth = s.ReadingWidth
		c.ReaderRadius = s.ReaderRadius
		c.BackgroundMode = s.BackgroundMode
	})
	return a.cfgMgr.Save()
}

// GetHighlights returns all highlights for a given file path. The filePath
// is validated against the root path to prevent traversal attacks.
func (a *App) GetHighlights(filePath string) ([]highlights.Highlight, error) {
	root := a.getRootPath()
	absPath, err := filepath.Abs(filepath.Join(root, filePath))
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	if !strings.HasPrefix(absPath, root+string(os.PathSeparator)) && absPath != root {
		return nil, fmt.Errorf("path outside root: %s", filePath)
	}
	return a.highlightStore.Load(filePath)
}

// AddHighlight adds a highlight to the store. The highlight's FilePath
// is validated against the root path.
func (a *App) AddHighlight(h highlights.Highlight) error {
	root := a.getRootPath()
	absPath, err := filepath.Abs(filepath.Join(root, h.FilePath))
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	if !strings.HasPrefix(absPath, root+string(os.PathSeparator)) && absPath != root {
		return fmt.Errorf("path outside root: %s", h.FilePath)
	}
	return a.highlightStore.Add(h)
}

// RemoveHighlight removes a highlight by ID from the given file.
func (a *App) RemoveHighlight(filePath, highlightID string) error {
	return a.highlightStore.Remove(filePath, highlightID)
}

// ClearHighlights removes all highlights for a given file.
func (a *App) ClearHighlights(filePath string) error {
	return a.highlightStore.Clear(filePath)
}
