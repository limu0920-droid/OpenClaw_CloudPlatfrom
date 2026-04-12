package main

import (
	"context"
	"log"
	"os"

	"openclaw/platformapi/internal/corestore"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	store, err := corestore.Open(databaseURL)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("close store: %v", err)
		}
	}()

	if err := store.Migrate(context.Background()); err != nil {
		log.Fatalf("migrate store: %v", err)
	}

	log.Print("migration completed")
}
