package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var OSExit = &analysis.Analyzer{
	Name: "osexitanalyzer",
	Doc:  "checks for direct os.Exit calls in main function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() == "main" {
			ast.Inspect(file, func(n ast.Node) bool {
				if fn, ok := n.(*ast.FuncDecl); ok {
					if fn.Name.Name == "main" && fn.Recv == nil {
						ast.Inspect(fn.Body, func(n ast.Node) bool {
							if call, ok := n.(*ast.CallExpr); ok {
								if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
									if pkgIdent, ok := sel.X.(*ast.Ident); ok {
										if pkgIdent.Name == "os" && sel.Sel.Name == "Exit" {
											pass.Reportf(call.Pos(), "direct call to os.Exit in main function of main package is not allowed")
										}
									}
								}
							}
							return true
						})
					}
				}
				return true
			})
		}
	}
	return nil, nil
}
