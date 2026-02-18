package enumcover

import (
	"fmt"
	"go/ast"
	"go/types"
	"regexp"
	"strconv"
	"strings"

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

func enumcoverCheck(pass *analysis.Pass) (any, error) {
	var index *constIndex

	for _, file := range pass.Files {
		if !fileHasEnumcoverDirective(file) {
			continue
		}

		commentMap := ast.NewCommentMap(pass.Fset, file, file.Comments)
		for n, groups := range commentMap {
			for _, group := range groups {
				for _, comment := range group.List {
					match := commentRegex.FindStringSubmatch(comment.Text)
					if len(match) == 2 {
						if index == nil {
							builtIndex := newConstIndex(pass)
							index = &builtIndex
						}
						typeName := fullTypeName(pass, file, n, strings.TrimSpace(match[1]))
						checkConsts(pass, n, typeName, *index)
					} else if strings.Contains(comment.Text, "enumcover:") {
						reportNodef(pass, comment, "Malformed enumcover comment (should be of the form \"enumcover:sometypename\"): %v", comment.Text)
					}
				}
			}
		}
	}

	return nil, nil
}

func fileHasEnumcoverDirective(file *ast.File) bool {
	for _, group := range file.Comments {
		for _, comment := range group.List {
			if strings.Contains(comment.Text, "enumcover:") {
				return true
			}
		}
	}
	return false
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

func checkConsts(pass *analysis.Pass, n ast.Node, typeName string, index constIndex) {
	allConsts := index.constsForType(typeName)
	namesForType := map[string]bool{}

	ast.Inspect(n, func(n ast.Node) bool {
		expr, ok := n.(ast.Expr)
		if !ok {
			return true
		}
		t := pass.TypesInfo.TypeOf(expr)
		if t == nil || t.String() != typeName {
			return true
		}

		switch n := n.(type) {
		case *ast.BasicLit:
			namesForType[unquote(n.Value)] = true
		case *ast.Ident:
			if namedConst, ok := allConsts[n.Name]; ok {
				namesForType[namedConst.val] = true
			}
		case *ast.SelectorExpr:
			if n.Sel != nil {
				if namedConst, ok := allConsts[n.Sel.Name]; ok {
					namesForType[namedConst.val] = true
				}
			}
		}
		return true
	})

	if len(allConsts) == 0 {
		reportNodef(pass, n, "No consts found for type %v", typeName)
	}
	for _, want := range allConsts {
		if !namesForType[want.val] {
			reportNodef(pass, n, "Unhandled const: %v", want)
		}
	}
}

func reportNodef(pass *analysis.Pass, node ast.Node, format string, args ...any) {
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

type constIndex struct {
	byType map[string]map[string]constVal
}

func newConstIndex(pass *analysis.Pass) constIndex {
	index := constIndex{byType: map[string]map[string]constVal{}}
	seenPkgs := map[*types.Package]struct{}{}

	var visit func(pkg *types.Package)
	visit = func(pkg *types.Package) {
		if pkg == nil {
			return
		}
		if _, ok := seenPkgs[pkg]; ok {
			return
		}
		seenPkgs[pkg] = struct{}{}

		scope := pkg.Scope()
		if scope != nil {
			for _, name := range scope.Names() {
				namedConst, ok := scope.Lookup(name).(*types.Const)
				if !ok {
					continue
				}
				typeName := namedConst.Type().String()
				constsForType := index.byType[typeName]
				if constsForType == nil {
					constsForType = map[string]constVal{}
					index.byType[typeName] = constsForType
				}
				constsForType[namedConst.Name()] = constVal{
					name: namedConst.Name(),
					val:  unquote(namedConst.Val().ExactString()),
				}
			}
		}

		for _, imp := range pkg.Imports() {
			visit(imp)
		}
	}

	visit(pass.Pkg)
	return index
}

func (i constIndex) constsForType(typeName string) map[string]constVal {
	consts := i.byType[typeName]
	if consts == nil {
		return map[string]constVal{}
	}
	return consts
}
