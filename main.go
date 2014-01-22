package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}

	findTestsInFile(os.Args[1])
}

func usage() {
	println("Error: Not enough args. Expected path to test file")
	println(fmt.Sprintf("usage: %s /path/to/some/file_test.go", os.Args[0]))
}

func findTestsInFile(pathToFile string) {
	if _, err := os.Stat(pathToFile); err != nil {
		dir, _ := os.Getwd()
		fmt.Printf("Couldn't find file from dir %s\n", dir)
		fmt.Printf("Error: given file '%s' does not exist\n", pathToFile)
		return
	}

	fileSet := token.NewFileSet()
	rootNode, err := parser.ParseFile(fileSet, pathToFile, nil, 0)
	if err != nil {
		fmt.Printf("Error parsing '%s':\n%s\n", pathToFile, err)
		return
	}

	testsToRewrite := findTestFuncs(rootNode)
	topLevelInitFunc := createInitBlock()

	describeBlock := createDescribeBlock()
	topLevelInitFunc.Body.List = append(topLevelInitFunc.Body.List, describeBlock)

	for _, testFunc := range testsToRewrite {
		rewriteTestInGinkgo(testFunc, rootNode, describeBlock)
	}

	rootNode.Decls = append(rootNode.Decls, topLevelInitFunc)

	var buffer bytes.Buffer
	if err := format.Node(&buffer, fileSet, rootNode); err != nil {
		println(err.Error())
		return
	}

	// TODO: take a flag to overwrite in place
	newFileName := strings.Replace(pathToFile, "_test.go", "_ginkgo_test.go", 1)
	ioutil.WriteFile(newFileName, buffer.Bytes(), 0666)
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

func findTestFuncs(rootNode *ast.File) (testsToRewrite []*ast.FuncDecl) {
	testNameRegexp := regexp.MustCompile("^Test[A-Z].+")

	ast.Inspect(rootNode, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		switch node := node.(type) {
		case *ast.FuncDecl:
			funcName := node.Name.Name
			// FIXME: also look at the params for this func
			matches := testNameRegexp.MatchString(funcName)
			if matches {
				testsToRewrite = append(testsToRewrite, node)
			}
		}

		return true
	})

	return
}

func rewriteTestInGinkgo(testFunc *ast.FuncDecl, rootNode *ast.File, describe *ast.ExprStmt) {
	var funcIndex int = -1
	for index, child := range rootNode.Decls {
		if child == testFunc {
			funcIndex = index
			break
		}
	}

	if funcIndex < 0 {
		println("Assert Error: Error finding index for test node %s\n", testFunc.Name.Name)
		os.Exit(1)
	}

	// create a new node
	blockStatement := &ast.BlockStmt{List: testFunc.Body.List}
	fieldList := &ast.FieldList{}
	funcType := &ast.FuncType{Params: fieldList}
	funcLit := &ast.FuncLit{Type: funcType, Body: blockStatement}
	basicLit := &ast.BasicLit{Kind: 9, Value: fmt.Sprintf("\"%s\"", testFunc.Name.Name)}
	itBlockIdent := &ast.Ident{Name: "It"}
	callExpr := &ast.CallExpr{Fun: itBlockIdent, Args: []ast.Expr{basicLit, funcLit} }
	expressionStatement := &ast.ExprStmt{X: callExpr}

	// attach the test expressions to the describe's list of statments
	var block *ast.BlockStmt = blockStatementFromDescribe(describe)
	block.List = append(block.List, expressionStatement)

	// append this to the declarations on the root node
	rootNode.Decls = append(rootNode.Decls[:funcIndex], rootNode.Decls[funcIndex+1:]...)

	return
}

func blockStatementFromDescribe(desc *ast.ExprStmt) (*ast.BlockStmt) {
	var funcLit *ast.FuncLit

	for _, node := range desc.X.(*ast.CallExpr).Args {
		switch node := node.(type) {
		case *ast.FuncLit:
    	funcLit = node
	    break
  	}
	}

	return funcLit.Body
}
