package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	windowsoptions "github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "Pi Switch - Provider Configurator for piAgent",
		Width:            1360,
		Height:           900,
		MinWidth:         1100,
		MinHeight:        760,
		Frameless:        true,
		WindowStartState: options.Normal,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		OnStartup:        app.startup,
		Windows: &windowsoptions.Options{
			DisableFramelessWindowDecorations: true,
			Theme:                             windowsoptions.Light,
			WebviewIsTransparent:             true,
			WindowIsTranslucent:              true,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
