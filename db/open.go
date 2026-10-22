package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io"
	"net/url"
	"path/filepath"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"

	"github.com/B4Dmonkey/bit-pro/store"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Open() (*sql.DB, error) {
	dir, err := store.Dir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, "main.db")

	mate := dbmate.New(&url.URL{Scheme: "sqlite", Path: path})
	mate.FS = migrationsFS
	mate.MigrationsDir = []string{"migrations"}
	mate.AutoDumpSchema = false
	mate.Log = io.Discard

	if err := mate.CreateAndMigrate(); err != nil {
		return nil, fmt.Errorf("migrating %s: %w", path, err)
	}

	query := url.Values{"_pragma": {"busy_timeout(5000)", "journal_mode(WAL)", "foreign_keys(1)"}}
	dsn := (&url.URL{Scheme: "file", Opaque: filepath.ToSlash(path), RawQuery: query.Encode()}).String()

	return sql.Open("sqlite", dsn)
}
