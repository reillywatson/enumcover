package enumcover_test

import (
	"testing"

	"github.com/reillywatson/enumcover"
	"golang.org/x/tools/go/analysis/analysistest"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, enumcover.Analyzer, "constdecl", "imported", "renamedimport", "stringenum")
}
