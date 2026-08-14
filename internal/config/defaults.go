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
		LastOpenedPath: "",
		FontSize:       16,
		SidebarWidth:   260,
		RecentPaths:    []string{},
	}
}
