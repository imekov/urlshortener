// Анализатор, запрещающий использовать прямой вызов os.Exit в функции main пакета main.
package main

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// MainOsExitAnalyzer содержит переменную типа &analysis.Analyzer и хранит имя, описание и функцию для запуска.
var MainOsExitAnalyzer = &analysis.Analyzer{
	Name: "mainosexit",
	Doc:  "checking the use of os.Exit in main function",
	Run:  run,
}

// run обходит файлы в поисках функции main, после чего производится поиск os.Exit.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				if x.Name.Name != "main" {
					return false
				}
			case *ast.SelectorExpr:
				if x.Sel.Name == "Exit" {
					pass.Reportf(x.Pos(), "error dont use os.exit")
				}
			}
			return true
		})
	}
	return nil, nil
}
