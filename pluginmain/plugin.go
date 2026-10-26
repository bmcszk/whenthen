// Package main is the Go plugin entry point for golangci-lint.
package main

import (
	"golang.org/x/tools/go/analysis"

	"github.com/bmcszk/whenthen"
)

func New(any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{whenthen.Analyzer}, nil
}

// main exists so the package also builds without -buildmode=plugin.
func main() {}
