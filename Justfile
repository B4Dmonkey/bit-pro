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
    #!/usr/bin/env sh
    dir="$(go env GOBIN)"; [ -n "$dir" ] || dir="$(go env GOPATH)/bin"
    go build -ldflags="-X 'github.com/B4Dmonkey/bit-pro/cmd.version={{version}}'" -o "$dir/bp" .

run *ARGS: db-gen-queries
    go run . {{ARGS}}

test: db-gen-queries
    go test ./...

fmt:
    go fmt ./...

lint: db-gen-queries
    golangci-lint run ./...
