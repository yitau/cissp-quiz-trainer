package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "CISSP Quiz Trainer",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:   &options.RGBA{R: 245, G: 247, B: 250, A: 1},
		OnStartup:          app.startup,
		OnShutdown:         app.shutdown,
		OnBeforeClose:      app.beforeClose,
		SingleInstanceLock: &options.SingleInstanceLock{UniqueId: "a8054168-8919-45a5-a89f-189342394914"},
		MinWidth:           760,
		MinHeight:          600,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
