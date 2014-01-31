package main

import(
	"os"
  "os/exec"
  "go/ast"
)

func addGinkgoSuiteFile(pathToPackage string) {
	originalDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = os.Chdir(pathToPackage)
	if err != nil {
		panic(err)
	}

	// FIXME : don't do this if we already have a test suite file
	cmd := exec.Command("ginkgo", "bootstrap")
	err = cmd.Run()

	if err != nil {
		panic(err.Error())
	}

	err = os.Chdir(originalDir)
	if err != nil {
		panic(err)
	}
}

func createInitBlock() *ast.FuncDecl {
	blockStatement := &ast.BlockStmt{List: []ast.Stmt{}}
	fieldList := &ast.FieldList{}
	funcType := &ast.FuncType{Params: fieldList}
	ident := &ast.Ident{Name: "init"}

	return &ast.FuncDecl{Name: ident, Type: funcType, Body: blockStatement}
}

func createDescribeBlock() *ast.ExprStmt {
	blockStatement := &ast.BlockStmt{List: []ast.Stmt{}}

	fieldList := &ast.FieldList{}
	funcType := &ast.FuncType{Params: fieldList}
	funcLit := &ast.FuncLit{Type: funcType, Body: blockStatement}
	basicLit := &ast.BasicLit{Kind: 9, Value: "\"Testing with ginkgo\""}
	describeIdent := &ast.Ident{Name: "Describe"}
	callExpr := &ast.CallExpr{Fun: describeIdent, Args: []ast.Expr{basicLit, funcLit}}

	return &ast.ExprStmt{X: callExpr}
}
