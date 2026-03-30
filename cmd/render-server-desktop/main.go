package main

import (
	"embed"
	"log"
	"time"

	"autosyncstudio/internal/renderserverapp"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed index.html
var assets embed.FS

func main() {
	go func() {
		app := renderserverapp.NewApp()
		if err := app.Run(); err != nil {
			log.Println("render-server http:", err)
		}
	}()

	time.Sleep(1200 * time.Millisecond)

	err := wails.Run(&options.App{
		Title:         "AutoSync Render Server",
		Width:         1380,
		Height:        940,
		MinWidth:      1200,
		MinHeight:     840,
		DisableResize: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 32, A: 255},
		Windows: &windows.Options{
			DisableWindowIcon: false,
			WindowClassName:   "AutoSyncRenderServerDesktop",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
