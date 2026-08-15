package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

var version = "dev"

func main() {
	args := os.Args[1:]
	rootPath := "."

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--version", "-v":
			fmt.Printf("ais %s\n", version)
			os.Exit(0)
		case "--mcp":
			fmt.Println("MCP mode not yet implemented")
			os.Exit(0)
		default:
			if len(args[i]) > 0 && args[i][0] != '-' {
				rootPath = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "unknown flag: %s\n", args[i])
				os.Exit(1)
			}
		}
	}

	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Path not found: %s\n", absPath)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Path is not a directory: %s\n", absPath)
		os.Exit(1)
	}

	app := NewApp(absPath)

	err = wails.Run(&options.App{
		Title:     "ais",
		Width:     1200,
		Height:    800,
		MinWidth:  600,
		MinHeight: 400,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		Linux: &linux.Options{
			Icon:                appIcon,
			WindowIsTranslucent: true,
			ProgramName:         "ais",
		},
		OnStartup: app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
