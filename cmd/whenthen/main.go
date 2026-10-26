package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/bmcszk/whenthen"
)

func main() {
	singlechecker.Main(whenthen.Analyzer)
}
