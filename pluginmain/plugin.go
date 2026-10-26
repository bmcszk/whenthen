// Package main is the Go plugin entry point for golangci-lint.
//
// Build with `go build -buildmode=plugin`. `main` exists so the package also
// compiles without that flag, which lets the linter run on this source.
package main

import (
	"golang.org/x/tools/go/analysis"

	"github.com/bmcszk/whenthen"
)

func New(any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{whenthen.Analyzer}, nil
}

func main() {}
