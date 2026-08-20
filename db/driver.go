package db

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	"github.com/amacneil/dbmate/v2/pkg/dbutil"

	_ "modernc.org/sqlite"
)

func init() {
	dbmate.RegisterDriver(newDriver, "sqlite")
}

type driver struct {
	migrationsTableName string
	databaseURL         *url.URL
	log                 io.Writer
}

func newDriver(config dbmate.DriverConfig) dbmate.Driver {
	return &driver{
		migrationsTableName: config.MigrationsTableName,
		databaseURL:         config.DatabaseURL,
		log:                 config.Log,
	}
}

func (drv *driver) path() string {
	return drv.databaseURL.Path
}

func (drv *driver) Open() (*sql.DB, error) {
	return sql.Open("sqlite", drv.path())
}

func (drv *driver) CreateDatabase() error {
	fmt.Fprintf(drv.log, "Creating: %s\n", drv.path())

	db, err := drv.Open()
	if err != nil {
		return err
	}
	defer dbutil.MustClose(db)

	return db.Ping()
}

func (drv *driver) DropDatabase() error {
	fmt.Fprintf(drv.log, "Dropping: %s\n", drv.path())

	exists, err := drv.DatabaseExists()
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	return os.Remove(drv.path())
}

func (drv *driver) DumpSchema(*sql.DB, ...string) ([]byte, error) {
	return nil, errors.New("dumping the schema requires the sqlite3 binary, which bit-pro does not depend on")
}

func (drv *driver) DatabaseExists() (bool, error) {
	_, err := os.Stat(drv.path())
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (drv *driver) MigrationsTableExists(db *sql.DB) (bool, error) {
	exists := false

	err := db.QueryRow("SELECT 1 FROM sqlite_master WHERE type='table' AND name=?",
		drv.migrationsTableName).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	return exists, err
}

func (drv *driver) CreateMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(fmt.Sprintf(
		"create table if not exists %s (version varchar(128) primary key)",
		drv.quotedMigrationsTableName()))

	return err
}

func (drv *driver) SelectMigrations(db *sql.DB, limit int) (map[string]bool, error) {
	limitClause := ""
	if limit >= 0 {
		limitClause = " limit " + strconv.Itoa(limit)
	}

	rows, err := db.Query(fmt.Sprintf("select version from %s order by version desc%s",
		drv.quotedMigrationsTableName(), limitClause))
	if err != nil {
		return nil, err
	}
	defer dbutil.MustClose(rows)

	migrations := map[string]bool{}

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}

		migrations[version] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return migrations, nil
}

func (drv *driver) InsertMigration(db dbutil.Transaction, version string) error {
	_, err := db.Exec(
		fmt.Sprintf("insert into %s (version) values (?)", drv.quotedMigrationsTableName()),
		version)

	return err
}

func (drv *driver) DeleteMigration(db dbutil.Transaction, version string) error {
	_, err := db.Exec(
		fmt.Sprintf("delete from %s where version = ?", drv.quotedMigrationsTableName()),
		version)

	return err
}

func (drv *driver) Ping() error {
	db, err := drv.Open()
	if err != nil {
		return err
	}
	defer dbutil.MustClose(db)

	return db.Ping()
}

func (drv *driver) QueryError(query string, err error) error {
	return &dbmate.QueryError{Err: err, Query: query}
}

func (drv *driver) quotedMigrationsTableName() string {
	return quoteIdentifier(drv.migrationsTableName)
}

func quoteIdentifier(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
