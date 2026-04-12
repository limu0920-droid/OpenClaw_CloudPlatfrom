package main

import (
	"context"
	"encoding/json"
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

	status, err := store.MigrationStatus(context.Background())
	if err != nil {
		log.Fatalf("migration status: %v", err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(status); err != nil {
		log.Fatalf("encode status: %v", err)
	}
}
