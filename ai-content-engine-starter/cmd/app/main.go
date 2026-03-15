package main

import (
	"log"

	"ai-content-engine-starter/internal/app"
)

func main() {
	application := app.New()
	if err := application.Run(); err != nil {
		log.Fatalf("application failed: %v", err)
	}
}
