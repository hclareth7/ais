package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	Theme          string   `json:"theme"`
	SSHKeyPaths    []string `json:"sshKeyPaths"`
	IgnoreDirs     []string `json:"ignoreDirs"`
	LastOpenedPath string   `json:"lastOpenedPath"`
	FontSize       int      `json:"fontSize"`
	SidebarWidth   int      `json:"sidebarWidth"`
	RecentPaths    []string `json:"recentPaths"`
	SelectedModel  string   `json:"selectedModel"`
	Provider       string   `json:"provider"`
	VertexProject  string   `json:"vertexProject"`
	VertexRegion   string   `json:"vertexRegion"`
	ZoomLevel      int      `json:"zoomLevel"`
	Opacity        int      `json:"opacity"`
	ReadingWidth   int      `json:"readingWidth"`
	ReaderRadius   int      `json:"readerRadius"`
	BackgroundMode string   `json:"backgroundMode"`
}

type Manager struct {
	mu       sync.RWMutex
	cfg      Config
	filePath string
}

func NewManager() *Manager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}
	return &Manager{
		cfg:      DefaultConfig(),
		filePath: filepath.Join(homeDir, ".config", "ais", "config.json"),
	}
}

func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, &m.cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	defaults := DefaultConfig()
	if m.cfg.VertexRegion == "" {
		m.cfg.VertexRegion = defaults.VertexRegion
	}
	if m.cfg.Provider == "" {
		m.cfg.Provider = defaults.Provider
	}
	if m.cfg.ZoomLevel == 0 {
		m.cfg.ZoomLevel = defaults.ZoomLevel
	}
	if m.cfg.Opacity == 0 {
		m.cfg.Opacity = defaults.Opacity
	}
	if m.cfg.ReadingWidth == 0 {
		m.cfg.ReadingWidth = defaults.ReadingWidth
	}
	if m.cfg.ReaderRadius == 0 {
		m.cfg.ReaderRadius = defaults.ReaderRadius
	}
	if m.cfg.BackgroundMode == "" {
		m.cfg.BackgroundMode = defaults.BackgroundMode
	}
	return nil
}

func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(m.cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(m.filePath, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

func (m *Manager) Update(fn func(*Config)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	fn(&m.cfg)
	return nil
}
