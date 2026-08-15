version := `git describe --tags --always --dirty 2>/dev/null || echo dev`

install:
    #!/usr/bin/env sh
    dir="$(go env GOBIN)"; [ -n "$dir" ] || dir="$(go env GOPATH)/bin"
    go build -ldflags="-X 'github.com/B4Dmonkey/bit-pro/cmd.version={{version}}'" -o "$dir/bp" .

run *ARGS:
    go run . {{ARGS}}

test:
    go test ./...

fmt:
    go fmt ./...

lint:
    golangci-lint run ./...
