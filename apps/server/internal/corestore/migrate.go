package corestore

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type migrationFile struct {
	Name    string
	Version string
	SQL     string
}

type AppliedMigration struct {
	Version   string
	Name      string
	AppliedAt time.Time
}

const (
	schemaMigrationLockKey int64 = 741001
	bootstrapLockKey       int64 = 741002
)

func (s *PostgresStore) Migrate(ctx context.Context) error {
	if s == nil || s.db == nil {
		return nil
	}

	files, err := loadMigrations()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS public.schema_migration (
			version VARCHAR(64) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, schemaMigrationLockKey); err != nil {
		return err
	}

	applied, err := appliedMigrations(ctx, tx)
	if err != nil {
		return err
	}

	for _, file := range files {
		if _, ok := applied[file.Version]; ok {
			continue
		}
		if _, err := tx.ExecContext(ctx, file.SQL); err != nil {
			return fmt.Errorf("apply migration %s: %w", file.Name, err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO public.schema_migration (version, name, applied_at)
			VALUES ($1, $2, NOW())
		`, file.Version, file.Name); err != nil {
			return err
		}
	}

	// Keep the embedded extension schema in sync for existing dev databases.
	if err := ensureExtendedSchema(ctx, tx); err != nil {
		return fmt.Errorf("apply extended schema: %w", err)
	}

	return tx.Commit()
}

func lockBootstrap(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, bootstrapLockKey)
	return err
}

func loadMigrations() ([]migrationFile, error) {
	return loadMigrationsFromFS(migrationFS, "migrations")
}

func loadMigrationsFromFS(fsys fs.FS, dir string) ([]migrationFile, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}

	files := make([]migrationFile, 0, len(entries))
	seenVersions := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		version, ok := migrationVersionFromName(entry.Name())
		if !ok {
			return nil, fmt.Errorf("invalid migration filename %q", entry.Name())
		}
		if previous, exists := seenVersions[version]; exists {
			return nil, fmt.Errorf("duplicate migration version %s: %s and %s", version, previous, entry.Name())
		}
		seenVersions[version] = entry.Name()
		raw, err := fs.ReadFile(fsys, filepath.ToSlash(filepath.Join(dir, entry.Name())))
		if err != nil {
			return nil, err
		}
		files = append(files, migrationFile{
			Name:    entry.Name(),
			Version: version,
			SQL:     string(raw),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].Version == files[j].Version {
			return files[i].Name < files[j].Name
		}
		return files[i].Version < files[j].Version
	})
	return files, nil
}

func migrationVersionFromName(name string) (string, bool) {
	version, _, ok := strings.Cut(name, "_")
	if !ok || strings.TrimSpace(version) == "" {
		return "", false
	}
	return version, true
}

func appliedMigrations(ctx context.Context, tx *sql.Tx) (map[string]struct{}, error) {
	rows, err := tx.QueryContext(ctx, `SELECT version FROM public.schema_migration`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions[version] = struct{}{}
	}
	return versions, rows.Err()
}

func (s *PostgresStore) MigrationStatus(ctx context.Context) ([]AppliedMigration, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, `
		SELECT version, name, applied_at
		FROM public.schema_migration
		ORDER BY version
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AppliedMigration, 0)
	for rows.Next() {
		var item AppliedMigration
		if err := rows.Scan(&item.Version, &item.Name, &item.AppliedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
