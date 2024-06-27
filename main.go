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
	log.Printf("processing %s ...\n", filepath)

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

		// Check if the function is slices.SortFunc
		ident, ok := selExpr.X.(*ast.Ident)
		if !ok {
			return true
		}
		if !(ident.Name == "slices" && isSortFuncName(selExpr.Sel.Name)) {
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

		// change the return type from bool to int
		if ident, ok := funcLit.Type.Results.List[0].Type.(*ast.Ident); ok {
			if ident.Name != "bool" {
				return true
			}
			ident.Name = "int"
		}

		// Update the function body
		newBody := []ast.Stmt{}
		for _, stmt := range funcLit.Body.List {
			if retstmt, ok := stmt.(*ast.ReturnStmt); ok {
				if len(retstmt.Results) != 1 {
					newBody = append(newBody, stmt)
					continue
				}
				expr, ok := retstmt.Results[0].(*ast.BinaryExpr)
				if !ok || hasLogicalOperators(expr) {
					newBody = append(newBody, stmt)
					continue
				}
				// aim to simple expression: a < b, a > b, a == b, a != b
				// if not, just keep the original statement
				//
				// TODO: handle more complex expressions
				if expr.Op != token.LSS && expr.Op != token.GTR && expr.Op != token.EQL && expr.Op != token.NEQ {
					newBody = append(newBody, stmt)
					continue
				}

				firstParam, secondParam := expr.X, expr.Y
				switch expr.Op {
				case token.LSS, token.EQL:
					firstParam, secondParam = expr.X, expr.Y
				case token.GTR, token.NEQ:
					firstParam, secondParam = expr.Y, expr.X
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
				isApply = true
				newBody = append(newBody, newStmt)
			} else {
				newBody = append(newBody, stmt)
			}
		}

		funcLit.Body.List = newBody
		return true
	}, nil)

	if !isApply {
		return nil
	}

	astutil.AddImport(fset, node, "cmp")

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return fmt.Errorf("Failed to format modified file: %w", err)
	}
	if err := os.WriteFile(filepath, buf.Bytes(), info.Mode().Perm()); err != nil {
		return fmt.Errorf("Failed to write modified file: %w", err)
	}
	return nil
}

func isSortFuncName(name string) bool {
	names := []string{
		"SortFunc",
		"SortStableFunc",
	}
	for _, n := range names {
		if strings.HasPrefix(name, n) {
			return true
		}
	}
	return false
}

func hasLogicalOperators(expr *ast.BinaryExpr) bool {
	if expr.Op == token.LAND || expr.Op == token.LOR {
		return true
	}
	if left, ok := expr.X.(*ast.BinaryExpr); ok {
		if hasLogicalOperators(left) {
			return true
		}
	}
	if right, ok := expr.Y.(*ast.BinaryExpr); ok {
		if hasLogicalOperators(right) {
			return true
		}
	}
	return false
}
