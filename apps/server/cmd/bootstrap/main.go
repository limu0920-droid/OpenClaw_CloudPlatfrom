package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"openclaw/platformapi/internal/corestore"
	"openclaw/platformapi/internal/models"
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

	bootstrapDataPath := strings.TrimSpace(os.Getenv("BOOTSTRAP_DATA_PATH"))
	if bootstrapDataPath == "" {
		log.Fatal("BOOTSTRAP_DATA_PATH is required; cmd/bootstrap no longer ships built-in mock seed data")
	}

	raw, err := os.ReadFile(bootstrapDataPath)
	if err != nil {
		log.Fatalf("read bootstrap data: %v", err)
	}

	var seed models.Data
	if err := json.Unmarshal(raw, &seed); err != nil {
		log.Fatalf("decode bootstrap data: %v", err)
	}

	data, err := store.Bootstrap(seed)
	if err != nil {
		log.Fatalf("bootstrap store: %v", err)
	}

	log.Printf("bootstrap completed: tenants=%d users=%d instances=%d channels=%d", len(data.Tenants), len(data.Users), len(data.Instances), len(data.Channels))
}
