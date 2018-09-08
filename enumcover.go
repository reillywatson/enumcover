package enumcover

import (
	"fmt"
	"go/ast"
	"regexp"
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
func (c *checker) Checks() []lint.Check {
	return []lint.Check{
		{ID: "enumcover001", FilterGenerated: false, Fn: enumcoverCheck},
	}
}

var commentRegex = regexp.MustCompile(`enumcover:([\w\.]+)`)

func enumcoverCheck(j *lint.Job) {
	for _, file := range j.Program.Files {
		commentMap := ast.NewCommentMap(j.Program.SSA.Fset, file, file.Comments)
		ast.Inspect(file, func(n ast.Node) bool {
			if n == nil {
				return true
			}
			for _, comments := range commentMap[n] {
				for _, comment := range comments.List {
					matches := commentRegex.FindAllStringSubmatch(comment.Text, 1)
					if len(matches) == 1 && len(matches[0]) == 2 {
						typeName := fullTypeName(j, file, n, strings.TrimSpace(matches[0][1]))
						checkConsts(j, n, typeName)
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
		typeName = pkg.SSA.Pkg.Path() + "." + typeName
	}
	return typeName
}

func checkConsts(j *lint.Job, n ast.Node, typeName string) {
	namesForType := map[string]bool{}
	ast.Inspect(n, func(n ast.Node) bool {
		if expr, ok := n.(ast.Expr); ok {
			t := j.NodePackage(expr).TypesInfo.TypeOf(expr)
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
	for _, want := range allConstsWithType(j, typeName) {
		if !namesForType[want.name] && !namesForType[want.val] {
			j.Errorf(n, "Unhandled const: %v", want)
		}
	}
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

func allConstsWithType(j *lint.Job, targetType string) []constVal {
	consts := []constVal{}
	for _, pkg := range j.Program.SSA.AllPackages() {
		for _, member := range pkg.Members {
			if namedConst, ok := member.(*ssa.NamedConst); ok {
				val := unquote(namedConst.Value.Value.ExactString())
				typeName := namedConst.Type().String()
				if typeName == targetType {
					consts = append(consts, constVal{name: namedConst.Name(), val: val})
				}
			}
		}
	}
	return consts
}
