package main

import (
	"embed"
	"fmt"
	"net"
	"log"
	"strings"
	"testing/fstest"

	"autosyncstudio/internal/appmeta"
	"autosyncstudio/internal/studioapp"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed index.html
var assets embed.FS

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	backendURL := fmt.Sprintf("http://%s", ln.Addr().String())

	go func() {
		app := studioapp.NewAppWithAddr(ln.Addr().String())
		if err := app.RunListener(ln); err != nil {
			log.Println("studio http:", err)
		}
	}()

	indexHTML, err := assets.ReadFile("index.html")
	if err != nil {
		log.Fatal(err)
	}
	renderedIndex := strings.ReplaceAll(string(indexHTML), "__BACKEND_URL__", backendURL)
	assetFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(renderedIndex)},
	}

	err = wails.Run(&options.App{
		Title:         appmeta.DisplayName,
		Width:         1080,
		Height:        980,
		MinWidth:      980,
		MinHeight:     860,
		DisableResize: false,
		AssetServer: &assetserver.Options{
			Assets: assetFS,
		},
		BackgroundColour: &options.RGBA{R: 9, G: 17, B: 29, A: 255},
		Windows: &windows.Options{
			DisableWindowIcon: false,
			WindowClassName:   "AutoSyncStudioDesktop",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
