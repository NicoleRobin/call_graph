package parser

import (
	"context"
	"fmt"
	"github.com/nicolerobin/log"
	"go.uber.org/zap"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
)

// ParseDir analyse specific directory
func ParseDir(ctx context.Context, dir string) error {
	// Create the AST by parsing src.
	fileSet := token.NewFileSet() // positions are relative to fset
	pkgs, err := parser.ParseDir(fileSet, dir, nil, 0)
	if err != nil {
		panic(err)
	}

	for pkgName, pkg := range pkgs {
		fmt.Printf("pkgName:%s\n", pkgName)
		for fileName, file := range pkg.Files {
			fmt.Printf("fileName:%s\n", fileName)
			// printFile(file)
			inspectFile(fileSet, file)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println(err)
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			err := ParseDir(ctx, path.Join(dir, entry.Name()))
			if err != nil {
				log.Error("parseDir() failed", zap.Error(err))
				continue
			}
		}
	}
	return nil
}

func inspectFile(fset *token.FileSet, file *ast.File) {
	ast.Inspect(file, func(node ast.Node) bool {
		var s string
		switch x := node.(type) {
		case *ast.BasicLit:
			s = "BasicLit:" + x.Value
		case *ast.Ident:
			s = "Ident:" + x.Name
		case *ast.CallExpr:
			switch fun := x.Fun.(type) {
			case *ast.Ident:
				s = "CallExpr-Ident:" + fun.Name
			case *ast.SelectorExpr:
				s = "CallExpr-SelectorExpr:" + fun.Sel.Name
			}
		}
		if s != "" {
			fmt.Printf("%s:\t%s\n", fset.Position(node.Pos()), s)
		}
		return true
	})
}

func printFile(file *ast.File) {
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			// 如果是函数声明
			funcName := funcDecl.Name.Name
			funcParams := funcDecl.Type.Params
			funcResults := funcDecl.Type.Results

			fmt.Printf("Function:%s\n", funcName)
			fmt.Println("Parameters:")
			for _, param := range funcParams.List {
				for _, name := range param.Names {
					fmt.Printf("%s: %s\n", name.Name, param.Type)
				}
			}
			if funcResults != nil {
				fmt.Printf("Result:")
				for _, result := range funcResults.List {
					fmt.Println(result.Type)
				}
			}
		}
	}
}
