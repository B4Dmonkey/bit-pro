version := `git describe --tags --always --dirty 2>/dev/null || echo dev`

build:
    go build -ldflags="-X 'github.com/B4Dmonkey/bit-pro/cmd.version={{version}}'" -o bin/bit .

run *ARGS:
    go run . {{ARGS}}

test:
    go test ./...

lint:
    golangci-lint run ./...
