package main

import (
	"log"
	"os"

	"github.com/jakshi/rbg/internal/app"
	"github.com/jakshi/rbg/internal/commands"
)

const configFilePath = ".config/rbg/config.json"

func main() {

	app, err := app.NewApp(configFilePath)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	if err := commands.Run(app, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
