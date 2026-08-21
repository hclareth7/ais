package config

func DefaultConfig() Config {
	return Config{
		Theme:       "system",
		SSHKeyPaths: []string{},
		IgnoreDirs: []string{
			".git",
			"node_modules",
			".svn",
			"__pycache__",
			"vendor",
			".venv",
		},
		LastOpenedPath:  "",
		FontSize:        16,
		SidebarWidth:    260,
		RecentPaths:     []string{},
		Provider:        "anthropic",
		VertexRegion:    "us-east5",
		ZoomLevel:       100,
		Opacity:         100,
		ReadingWidth:    1000,
		ReaderRadius:    20,
		BackgroundMode:  "gradient",
	}
}
