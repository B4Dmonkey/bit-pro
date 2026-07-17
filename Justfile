build:
    go build -o bin/bit .

run *ARGS:
    go run . {{ARGS}}

test:
    go test ./...

lint:
    golangci-lint run ./...
