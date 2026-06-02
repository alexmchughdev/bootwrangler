package main

import (
	"embed"
	"log"

	"github.com/alexmchughdev/bootwrangler/internal/app"
	_ "github.com/alexmchughdev/bootwrangler/internal/renderers/all"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	service := app.NewService()

	err := wails.Run(&options.App{
		Title:  "BootWrangler",
		Width:  1440,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 10, G: 18, B: 32, A: 1},
		OnStartup:        service.Startup,
		Bind: []interface{}{
			service,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
