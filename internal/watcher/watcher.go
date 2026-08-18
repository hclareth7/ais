package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	".svn":         true,
	"__pycache__":  true,
	"vendor":       true,
	".venv":        true,
}

// WatchEvent represents a file system event for a markdown file.
type WatchEvent struct {
	Path     string // relative path from root
	IsCreate bool   // true for newly created files
}

type Watcher struct {
	fsWatcher *fsnotify.Watcher
	root      string
	callback  func(WatchEvent)
	done      chan struct{}
	once      sync.Once
	stopped   bool
	stoppedMu sync.RWMutex
}

func New(root string, callback func(WatchEvent)) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		fsWatcher: fsw,
		root:      root,
		callback:  callback,
		done:      make(chan struct{}),
	}

	if err := w.addRecursive(root); err != nil {
		fsw.Close()
		return nil, err
	}

	return w, nil
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if skipDirs[name] || (len(name) > 0 && name[0] == '.' && path != root) {
				return filepath.SkipDir
			}
			return w.fsWatcher.Add(path)
		}
		return nil
	})
}

func isMarkdown(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown")
}

func (w *Watcher) Start() error {
	debounce := make(map[string]*time.Timer)
	var mu sync.Mutex

	for {
		select {
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return nil
			}

			if !isMarkdown(filepath.Base(event.Name)) {
				if event.Has(fsnotify.Create) {
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						w.addRecursive(event.Name)
					}
				}
				continue
			}

			isCreate := event.Has(fsnotify.Create)
			if event.Has(fsnotify.Write) || isCreate || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				mu.Lock()
				if timer, exists := debounce[event.Name]; exists {
					timer.Stop()
				}
				name := event.Name
				created := isCreate
				debounce[event.Name] = time.AfterFunc(100*time.Millisecond, func() {
					w.stoppedMu.RLock()
					isStopped := w.stopped
					w.stoppedMu.RUnlock()
					if isStopped {
						return
					}
					rel, err := filepath.Rel(w.root, name)
					if err != nil {
						rel = name
					}
					w.callback(WatchEvent{Path: rel, IsCreate: created})
					mu.Lock()
					delete(debounce, name)
					mu.Unlock()
				})
				mu.Unlock()
			}

		case _, ok := <-w.fsWatcher.Errors:
			if !ok {
				return nil
			}

		case <-w.done:
			return nil
		}
	}
}

func (w *Watcher) Stop() {
	w.once.Do(func() {
		w.stoppedMu.Lock()
		w.stopped = true
		w.stoppedMu.Unlock()
		close(w.done)
		w.fsWatcher.Close()
	})
}
