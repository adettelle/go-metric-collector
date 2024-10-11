package exitcheckanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Для создания своего анализатора нужно определить переменную типа analysis.Analyzer
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for os exit",
	Run:  run,
}

// анализатор, запрещающий использовать прямой вызов os.Exit в функции main пакета main.
func run(pass *analysis.Pass) (interface{}, error) {
	isTargetPkg := func(x *ast.File) bool {
		return x.Name.Name == "main" && !ast.IsGenerated(x)
	}

	isTargetFunc := func(x *ast.FuncDecl) bool {
		return x.Name.Name == "main"
	}

	for _, file := range pass.Files {
		// запускаем инспектор, который рекурсивно обходит ветви AST
		// передаём инспектирующую функцию анонимно

		ast.Inspect(file, func(node ast.Node) bool {
			// проверяем, какой конкретный тип лежит в узле
			switch x := node.(type) {
			case *ast.File:
				if !isTargetPkg(x) { // если не пакет main, глубже мы не идем
					return false
				}
			case *ast.FuncDecl:
				if !isTargetFunc(x) { // если функция не main
					return false
				}
			case *ast.SelectorExpr:
				indent, ok := x.X.(*ast.Ident)
				if !ok {
					return false
				}
				if indent.Name == "os" && x.Sel.Name == "Exit" {
					pass.Reportf(x.Pos(), "expression os.Exit() in main func in main package")
				}
			}
			return true
		})
	}
	return nil, nil
}
