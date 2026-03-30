package main

import (
	"log"

	"autosyncstudio/internal/studioapp"
)

func main() {
	app := studioapp.NewApp()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
