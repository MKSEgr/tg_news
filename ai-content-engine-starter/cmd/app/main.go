package main

import (
	"log"

	"ai-content-engine-starter/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("create app: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("application failed: %v", err)
	}
}
