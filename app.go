package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hclareth7/ais/internal/config"
	"github.com/hclareth7/ais/internal/scanner"
	"github.com/hclareth7/ais/internal/types"
	"github.com/hclareth7/ais/internal/watcher"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx      context.Context
	rootPath string
	cfgMgr   *config.Manager
	watcher  *watcher.Watcher
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

	w, err := watcher.New(a.rootPath, func(path string) {
		wailsRuntime.EventsEmit(a.ctx, "file:changed", path)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to start watcher: %v\n", err)
	} else {
		a.watcher = w
		go a.watcher.Start()
	}
}

func (a *App) shutdown(ctx context.Context) {
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

	w, err := watcher.New(absPath, func(changedPath string) {
		wailsRuntime.EventsEmit(a.ctx, "file:changed", changedPath)
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
