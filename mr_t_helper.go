package main

import(
	"go/ast"
)

func typeFromMrTPackage(name string) *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X: &ast.Ident{Name: "mr"},
		Sel: &ast.Ident{Name: name},
	}
}

func newMrTFromIdent(ident *ast.Ident) *ast.CallExpr {
	return &ast.CallExpr{
		Lparen: ident.NamePos + 1,
		Rparen: ident.NamePos + 2,
    Fun: typeFromMrTPackage("T"),
	}
}

func newMrTestingT() *ast.SelectorExpr {
	return typeFromMrTPackage("TestingT")
}
