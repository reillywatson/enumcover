package enumcover

import (
	"go/ast"
	"strconv"
	"strings"

	"honnef.co/go/tools/lint"
	"honnef.co/go/tools/ssa"
)

func NewChecker() lint.Checker {
	return &checker{}
}

type checker struct{}

func (*checker) Init(*lint.Program) {}
func (*checker) Name() string       { return "enumcover" }
func (*checker) Prefix() string     { return "enumcover" }
func (*checker) Funcs() map[string]lint.Func {
	return map[string]lint.Func{
		"enumcover001": enumcoverCheck,
	}
}

func enumcoverCheck(j *lint.Job) {
	for _, file := range j.Program.Files {
		commentMap := ast.NewCommentMap(j.Program.SSA.Fset, file, file.Comments)
		ast.Inspect(file, func(n ast.Node) bool {
			if n == nil {
				return true
			}
			for _, comments := range commentMap[n] {
				for _, comment := range comments.List {
					if strings.HasPrefix(strings.TrimSpace(comment.Text), "//handleall:") {
						parts := strings.Split(comment.Text, ":")
						if len(parts) == 2 {
							typeName := fullTypeName(j, file, n, strings.TrimSpace(parts[1]))
							checkConsts(j, n, typeName)
						}
					}
				}
			}
			return true
		})
	}
}

func fullTypeName(j *lint.Job, file *ast.File, n ast.Node, typeName string) string {
	selectorParts := strings.Split(typeName, ".")
	if len(selectorParts) == 2 {
		for _, fimport := range file.Imports {
			var pkgName string
			if fimport.Name != nil {
				if fimport.Name.Name == "." {
					// TODO: handle dot imports
					j.Errorf(n, "Dot imports are unhandled!")
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
		pkg := j.NodePackage(n)
		typeName = pkg.Pkg.Path() + "." + typeName
	}
	return typeName
}

func checkConsts(j *lint.Job, n ast.Node, typeName string) {
	namesForType := map[string]bool{}
	for _, c := range allConstsWithType(j, typeName) {
		namesForType[c] = false
	}
	ast.Inspect(n, func(n ast.Node) bool {
		if expr, ok := n.(ast.Expr); ok {
			t := j.Program.Info.TypeOf(expr)
			if t != nil && t.String() == typeName {
				switch n := n.(type) {
				case *ast.BasicLit:
					namesForType[unquote(n.Value)] = true
				case *ast.Ident:
					if n.Obj != nil && n.Obj.Kind == ast.Con {
						if decl, ok := n.Obj.Decl.(*ast.ValueSpec); ok {
							for _, value := range decl.Values {
								if lit, ok := value.(*ast.BasicLit); ok {
									namesForType[unquote(lit.Value)] = true
								}
							}
						}
					}
				}
			}
		}
		return true
	})
	for k, v := range namesForType {
		if !v {
			j.Errorf(n, "Unhandled const: %s", k)
		}
	}
}

func unquote(str string) string {
	if unquoted, err := strconv.Unquote(str); err == nil {
		return unquoted
	}
	return str
}

func allConstsWithType(j *lint.Job, targetType string) []string {
	consts := []string{}
	for _, pkg := range j.Program.SSA.AllPackages() {
		for _, member := range pkg.Members {
			if namedConst, ok := member.(*ssa.NamedConst); ok {
				val := unquote(namedConst.Value.Value.ExactString())
				typeName := namedConst.Type().String()
				if typeName == targetType {
					consts = append(consts, val)
				}
			}
		}
	}
	return consts
}
