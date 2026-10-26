// Command whenthen is the `go vet -vettool` binary for the given/when/then
// analyzer. It is also the singlechecker used by `go vet`.
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/bmcszk/whenthen"
)

func main() {
	singlechecker.Main(whenthen.Analyzer)
}
