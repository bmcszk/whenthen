set shell := ["bash", "-uc"]

# List available recipes.
default:
    just --list

# Compile every package.
build:
    go build ./...

# Run the test suite with the race detector.
test:
    go test -race ./...

# Run tests with a coverage summary.
cover:
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out | tail -1

# Run the linters.
lint:
    golangci-lint run ./...

# Check formatting without writing.
fmt-check:
    golangci-lint fmt --diff

# Apply the formatters.
fmt:
    golangci-lint fmt

# Dogfood: vet this repository and the prose fixture with its own analyzer.
selfcheck:
    go build -o bin/whenthen ./cmd/whenthen
    go vet -vettool=$(pwd)/bin/whenthen ./...
    go vet -vettool=$(pwd)/bin/whenthen ./testdata/prose/...

# Run the fuzz targets briefly.
fuzz:
    go test -run=NONE -fuzz=FuzzHasWord -fuzztime=15s .
    go test -run=NONE -fuzz=FuzzEmptyPhase -fuzztime=15s .

# Install the analyzer command.
install:
    go install ./cmd/whenthen

# Build the golangci-lint Go plugin.
plugin:
    ./scripts/build-plugin.sh

# Run every gate.
verify: fmt-check lint selfcheck test build prose-check

# Probe file that must stay clean: prose containing phase words must not be flagged.
prose-check:
	go build -o bin/whenthen ./cmd/whenthen
	go vet -vettool=$(pwd)/bin/whenthen ./testdata/prose/...
