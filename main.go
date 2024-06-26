package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <file.go>", os.Args[0])
	}

	filepath := os.Args[1]
	if err := run(filepath); err != nil {
		log.Fatalf("Failed to run: %v", err)
	}
}

func run(filepath string) error {
	fset := token.NewFileSet()
	info, err := os.Stat(filepath)
	if err != nil {
		return fmt.Errorf("Failed to stat file: %w", err)
	}
	if info.IsDir() {
		return nil
	}
	if !strings.HasSuffix(info.Name(), ".go") {
		return fmt.Errorf("Not a go file: %s", info.Name())
	}

	node, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("Failed to parse file: %w", err)
	}

	isApply := false
	astutil.Apply(node, func(cr *astutil.Cursor) bool {
		callExpr, ok := cr.Node().(*ast.CallExpr)
		if !ok {
			return true
		}

		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if !(selExpr.X.(*ast.Ident).Name == "slices" && strings.HasPrefix(selExpr.Sel.Name, "SortFunc")) {
			return true
		}

		if len(callExpr.Args) != 2 {
			return true
		}

		funcLit, ok := callExpr.Args[1].(*ast.FuncLit)
		if !ok {
			return true
		}
		if funcLit.Type.Results == nil || len(funcLit.Type.Results.List) != 1 {
			return true
		}

		ident, ok := funcLit.Type.Results.List[0].Type.(*ast.Ident)
		if !ok || ident.Name != "bool" {
			return true
		}

		// Change the return type to int
		ident.Name = "int"

		// Update the function body
		newBody := []ast.Stmt{}
		for _, stmt := range funcLit.Body.List {
			// 無名関数で1行で書かれているような単純な式のみを対象とする
			if stmt, ok := stmt.(*ast.ReturnStmt); ok {
				for _, expr := range stmt.Results {
					if binaryExpr, ok := expr.(*ast.BinaryExpr); ok {
						// compare関数の引数順序を決定する
						var (
							firstParam  ast.Expr
							secondParam ast.Expr
						)
						switch binaryExpr.Op {
						case token.LSS:
							firstParam = binaryExpr.X
							secondParam = binaryExpr.Y
						case token.GTR:
							firstParam = binaryExpr.Y
							secondParam = binaryExpr.X
						}

						newStmt := &ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   ast.NewIdent("cmp"),
										Sel: ast.NewIdent("Compare"),
									},
									Args: []ast.Expr{
										firstParam,
										secondParam,
									},
								},
							},
						}
						newBody = append(newBody, newStmt)
					}
				}
			}
		}
		funcLit.Body.List = newBody
		return true
	}, nil)

	if !isApply {
		return nil
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return fmt.Errorf("Failed to format modified file: %w", err)
	}
	if err := os.WriteFile(filepath, buf.Bytes(), info.Mode().Perm()); err != nil {
		return fmt.Errorf("Failed to write modified file: %w", err)
	}
	log.Printf("processed %s\n", filepath)
	return nil
}
