#!/usr/bin/env sh
set -eu

cd "$(dirname "$0")/.."

version="$(git describe --tags --match 'v*' --always --dirty 2>/dev/null | sed 's/^v//' || echo dev)"

dir="$(go env GOBIN)"
[ -n "$dir" ] || dir="$(go env GOPATH)/bin"

echo "Generating the query layer..."
sqlc generate

echo "Building bp $version into $dir..."
go build -ldflags="-X 'github.com/B4Dmonkey/bit-pro/cmd.version=$version'" -o "$dir/bp" .

echo "Registering the bit-pro marketplace..."
claude plugin marketplace add B4Dmonkey/bit-pro

echo
echo "Installed $dir/bp"
echo "Run 'bp init' in a project to set it up."
