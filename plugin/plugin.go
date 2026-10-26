// Package plugin registers the analyzer with the golangci-lint module plugin system.
package plugin

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/bmcszk/whenthen"
)

type Plugin struct{}

func init() {
	register.Plugin("whenthen", New)
}

func New(any) (register.LinterPlugin, error) {
	return Plugin{}, nil
}

func (Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{whenthen.Analyzer}, nil
}

func (Plugin) GetLoadMode() string {
	return register.LoadModeSyntax
}
