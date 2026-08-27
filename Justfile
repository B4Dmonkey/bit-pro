version := `git describe --tags --match 'bit--v*' --always --dirty 2>/dev/null | sed 's/^bit--v//' || echo dev`

MIGRATIONS_DIR := justfile_directory() / "db" / "migrations"

export DATABASE_URL := "sqlite:" + justfile_directory() / "db" / "bit.db"

db-gen-queries:
    sqlc generate

# create a new migration file
db-migrate name:
    dbmate --migrations-dir "{{MIGRATIONS_DIR}}" new {{name}}

# run all pending migrations against the throwaway db/bit.db
db-up:
    dbmate --no-dump-schema --migrations-dir "{{MIGRATIONS_DIR}}" up

# roll back the last migration against the throwaway db/bit.db
db-down:
    dbmate --no-dump-schema --migrations-dir "{{MIGRATIONS_DIR}}" down

# show migration status
db-status:
    dbmate --migrations-dir "{{MIGRATIONS_DIR}}" status

install: db-gen-queries
    "{{justfile_directory()}}/scripts/install.sh"

# cut a release: compute the next version for a bump level, guard, write plugin.json, commit and tag
release level:
    "{{justfile_directory()}}/scripts/release.sh" {{level}}

# publish the current version's tag to origin (tag only — push the branch yourself)
release-push:
    "{{justfile_directory()}}/scripts/release-push.sh"

run *ARGS: db-gen-queries
    go run . {{ARGS}}

test: db-gen-queries
    go test ./...

fmt:
    go fmt ./...

lint: db-gen-queries
    golangci-lint run ./...
