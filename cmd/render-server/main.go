package main

import (
	"log"

	"autosyncstudio/internal/renderserverapp"
)

func main() {
	app := renderserverapp.NewApp()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
