version := `git describe --tags --always --dirty 2>/dev/null || echo dev`

db-gen-queries:
    sqlc generate

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
