package main

import (
	"os"

	"github.com/reillywatson/enumcover"
	"honnef.co/go/tools/lint/lintutil"
)

func main() {
	checkers := []lintutil.CheckerConfig{
		{Checker: enumcover.NewChecker(), ExitNonZero: true},
	}
	lintutil.ProcessArgs("enumcover", checkers, os.Args[1:])
}
