package main

import (
	"github.com/reillywatson/enumcover"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(enumcover.Analyzer) }
