package main

import (
	"os"

	"github.com/reillywatson/enumcover"
	"honnef.co/go/tools/lint"
	"honnef.co/go/tools/lint/lintutil"
)

func main() {
	fs := lintutil.FlagSet("enumcover")
	fs.Parse(os.Args[1:])

	checkers := []lint.Checker{
		enumcover.NewChecker(),
	}
	lintutil.ProcessFlagSet(checkers, fs)
}
