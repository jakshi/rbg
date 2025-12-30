package main

import (
	"fmt"
	"log"

	"github.com/jakshi/rbg/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	cfg.CurrentUserName = "kos"
	err = config.SetUser(cfg)
	if err != nil {
		log.Fatalf("Failed to set user: %v", err)
	}

	cfg, err = config.Read()
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	fmt.Println(*cfg)
}
