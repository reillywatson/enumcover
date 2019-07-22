package enumcover

import (
	"fmt"
	"go/ast"
	"go/types"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

const Doc = `check that code blocks cover all consts of a given type`

var Analyzer = &analysis.Analyzer{
	Doc:      Doc,
	Name:     "enumcover",
	Run:      enumcoverCheck,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}
var commentRegex = regexp.MustCompile(`enumcover:([\w\.]+)`)

func enumcoverCheck(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		commentMap := ast.NewCommentMap(pass.Fset, file, file.Comments)
		ast.Inspect(file, func(n ast.Node) bool {
			if n == nil {
				return true
			}
			for _, comments := range commentMap[n] {
				for _, comment := range comments.List {
					matches := commentRegex.FindAllStringSubmatch(comment.Text, 1)
					if len(matches) == 1 && len(matches[0]) == 2 {
						typeName := fullTypeName(pass, file, n, strings.TrimSpace(matches[0][1]))
						checkConsts(pass, n, typeName)
					}
				}
			}
			return true
		})
	}
	return nil, nil
}

func fullTypeName(pass *analysis.Pass, file *ast.File, n ast.Node, typeName string) string {
	selectorParts := strings.Split(typeName, ".")
	if len(selectorParts) == 2 {
		for _, fimport := range file.Imports {
			var pkgName string
			if fimport.Name != nil {
				if fimport.Name.Name == "." {
					// TODO: handle dot imports
					reportNodef(pass, n, "Dot imports are unhandled!")
				}
				pkgName = fimport.Name.Name
			} else {
				components := strings.Split(unquote(fimport.Path.Value), "/")
				pkgName = components[len(components)-1]
			}
			if selectorParts[0] == pkgName {
				typeName = unquote(fimport.Path.Value) + "." + selectorParts[1]
			}
		}
	} else {
		typeName = pass.Pkg.Path() + "." + typeName
	}
	return typeName
}

func checkConsts(pass *analysis.Pass, n ast.Node, typeName string) {
	namesForType := map[string]bool{}
	ast.Inspect(n, func(n ast.Node) bool {
		if expr, ok := n.(ast.Expr); ok {
			t := pass.TypesInfo.TypeOf(expr)
			if t != nil && t.String() == typeName {
				switch n := n.(type) {
				case *ast.BasicLit:
					namesForType[unquote(n.Value)] = true
				case *ast.Ident:
					if n.Obj != nil {
						if n.Obj.Kind == ast.Con {
							if decl, ok := n.Obj.Decl.(*ast.ValueSpec); ok {
								for _, value := range decl.Values {
									if lit, ok := value.(*ast.BasicLit); ok {
										namesForType[unquote(lit.Value)] = true
									}
								}
							}
						}
					}
					namesForType[n.Name] = true
				}
			}
		}
		return true
	})
	allConsts := allConstsWithType(pass, typeName)
	if len(allConsts) == 0 {
		reportNodef(pass, n, "No consts found for type %v", typeName)
	}
	for _, want := range allConsts {
		if !namesForType[want.name] && !namesForType[want.val] {
			reportNodef(pass, n, "Unhandled const: %v", want)
		}
	}
}

func reportNodef(pass *analysis.Pass, node ast.Node, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	pass.Report(analysis.Diagnostic{Pos: node.Pos(), End: node.End(), Message: msg})
}

func unquote(str string) string {
	if unquoted, err := strconv.Unquote(str); err == nil {
		return unquoted
	}
	return str
}

type constVal struct {
	name string
	val  string
}

func (c constVal) String() string {
	return fmt.Sprintf("%s (%s)", c.name, c.val)
}

var allPkgs sync.Map

// TODO: do this by storing analysis.Facts about all the consts in each package?
func allConstsWithType(pass *analysis.Pass, targetType string) []constVal {
	var visit func(pkg *types.Package)
	visit = func(pkg *types.Package) {
		if _, ok := allPkgs.Load(pkg); ok {
			return
		}
		allPkgs.Store(pkg, struct{}{})
		for _, imp := range pkg.Imports() {
			visit(imp)
		}
	}
	visit(pass.Pkg)
	consts := []constVal{}
	allPkgs.Range(func(pkgKey, _ interface{}) bool {
		pkg := pkgKey.(*types.Package)
		for _, name := range pkg.Scope().Names() {
			if namedConst, ok := pkg.Scope().Lookup(name).(*types.Const); ok {
				val := unquote(namedConst.Val().ExactString())
				typeName := namedConst.Type().String()
				if typeName == targetType {
					consts = append(consts, constVal{name: namedConst.Name(), val: val})
				}
			}
		}
		return true
	})
	return consts
}
