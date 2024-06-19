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
	// skip analysis if the package is not "main"
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			// check if the node is a function declaration named "main" without a receiver
			if fn, ok := n.(*ast.FuncDecl); ok {
				if fn.Name.Name == "main" && fn.Recv == nil {
					// inspect the body of the "main" function for calls to os.Exit
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

	return nil, nil
}
