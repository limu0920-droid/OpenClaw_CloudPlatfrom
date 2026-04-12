package corestore

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadMigrationsEmbeddedVersionsUniqueAndSorted(t *testing.T) {
	files, err := loadMigrations()
	if err != nil {
		t.Fatalf("load embedded migrations: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected embedded migrations")
	}

	seen := make(map[string]string, len(files))
	for index, file := range files {
		if previous, exists := seen[file.Version]; exists {
			t.Fatalf("duplicate embedded migration version %s in %s and %s", file.Version, previous, file.Name)
		}
		seen[file.Version] = file.Name
		if index == 0 {
			continue
		}
		previous := files[index-1]
		if previous.Version > file.Version {
			t.Fatalf("expected sorted migrations, got %s before %s", previous.Name, file.Name)
		}
		if previous.Version == file.Version && previous.Name > file.Name {
			t.Fatalf("expected stable name ordering for version %s, got %s before %s", file.Version, previous.Name, file.Name)
		}
	}
}

func TestLoadMigrationsRejectsDuplicateVersions(t *testing.T) {
	_, err := loadMigrationsFromFS(fstest.MapFS{
		"migrations/0001_base.sql":  {Data: []byte("select 1;")},
		"migrations/0001_other.sql": {Data: []byte("select 2;")},
		"migrations/0002_next.sql":  {Data: []byte("select 3;")},
	}, "migrations")
	if err == nil {
		t.Fatal("expected duplicate version error")
	}
	if !strings.Contains(err.Error(), "duplicate migration version 0001") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadMigrationsRejectsInvalidFilename(t *testing.T) {
	_, err := loadMigrationsFromFS(fstest.MapFS{
		"migrations/not-a-migration.sql": {Data: []byte("select 1;")},
	}, "migrations")
	if err == nil {
		t.Fatal("expected invalid filename error")
	}
	if !strings.Contains(err.Error(), "invalid migration filename") {
		t.Fatalf("unexpected error: %v", err)
	}
}
