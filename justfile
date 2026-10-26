set shell := ["bash", "-uc"]

# List available recipes.
default:
    just --list

# Compile every package.
build:
    go build ./...

# Run the test suite.
test:
    go test ./...

# Run tests with coverage.
cover:
    go test -cover ./...

# Run the linters.
lint:
    golangci-lint run ./...

# Apply the formatters.
fmt:
    golangci-lint fmt

# Install the analyzer command.
install:
    go install ./cmd/whenthen

# Build the golangci-lint plugin binary.
plugin:
    ./scripts/build-plugin.sh

# Run lint, tests, and build.
verify: lint test build
