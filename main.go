package main

import (
	"os"
	"fmt"
	"bytes"
	"regexp"
	"go/ast"
	"go/token"
	"go/parser"
	"go/format"
	"io/ioutil"
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
	parsedFile, err := parser.ParseFile(fileSet, pathToFile, nil, 0)
	if err != nil {
		fmt.Printf("Error parsing '%s':\n%s\n", pathToFile, err)
		return
	}

	testsToRewrite, rootNode := findTestFuncsAndRootNode(parsedFile)
	topLevelDescribe := createDescribeBlock()

	for _, testFunc := range testsToRewrite {
		rewriteTestInGinkgo(testFunc, rootNode, topLevelDescribe)
	}

	rootNode.Decls = append(rootNode.Decls, topLevelDescribe)

	ast.Inspect(rootNode, func(node ast.Node) bool {
		if node == nil {
			return true
		}

		fmt.Printf("%p => %#v\n", node, node)
		return true
	})

	src, err := gofmtFile(parsedFile, fileSet)
	if err != nil {
		println(err.Error())
		return
	}

	// TODO: overwrite in place
	newFileName := fmt.Sprintf("%s.rewrite", pathToFile)
	ioutil.WriteFile(newFileName, src, 0666)
}

func createDescribeBlock() (decl *ast.GenDecl) {
	blockStatement := &ast.BlockStmt{List: []ast.Stmt{}}

	fieldList := &ast.FieldList{}
	funcType := &ast.FuncType{Params: fieldList}
	funcLit := &ast.FuncLit{Type: funcType, Body: blockStatement}
	basicLit := &ast.BasicLit{Kind: 9, Value :"\"Testing with ginkgo\""}
	describeIdent := &ast.Ident{Name: "Describe"}
	callExpr := &ast.CallExpr{Fun: describeIdent, Args: []ast.Expr{basicLit, funcLit} }
	ignoredDesc := &ast.Ident{Name: "_"}
	valueSpec := &ast.ValueSpec{Values: []ast.Expr{callExpr}, Names: []*ast.Ident{ignoredDesc} }
	decl = &ast.GenDecl{Specs: []ast.Spec{valueSpec} }

	return
}

func findTestFuncsAndRootNode(parsedFile *ast.File) (testsToRewrite []*ast.FuncDecl, rootNode *ast.File) {
	testNameRegexp := regexp.MustCompile("^Test[A-Z].+")

	ast.Inspect(parsedFile, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		switch node := node.(type) {
		case *ast.File:
			rootNode = node
		case *ast.FuncDecl:
			funcName := node.Name.Name
			// TODO: also look at the params for this func
			matches := testNameRegexp.MatchString(funcName)
			if matches {
				testsToRewrite = append(testsToRewrite, node)
			}
		}

		return true
	})

	return
}

func rewriteTestInGinkgo(testFunc *ast.FuncDecl, rootNode *ast.File, decl *ast.GenDecl) {
	// find index of testFunc in rootNode.Decls slice
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

	fmt.Printf("found index of test func %s :: %d\n", testFunc.Name.Name, funcIndex)

	// create a new node
	blockStatement := &ast.BlockStmt{List: testFunc.Body.List}
	fieldList := &ast.FieldList{}
	funcType := &ast.FuncType{Params: fieldList}
	funcLit := &ast.FuncLit{Type: funcType, Body: blockStatement}
	basicLit := &ast.BasicLit{Kind: 9, Value: fmt.Sprintf("\"%s\"", testFunc.Name.Name)}
	itBlockIdent := &ast.Ident{Name: "It"}
	callExpr := &ast.CallExpr{Fun: itBlockIdent, Args: []ast.Expr{basicLit, funcLit} }
	expressionStatement := &ast.ExprStmt{X: callExpr}

	var block *ast.BlockStmt = blockStatementFromDecl(decl)
	block.List = append(block.List, expressionStatement)
	rootNode.Decls = append(rootNode.Decls[:funcIndex], rootNode.Decls[funcIndex+1:]...)

	return
}

func blockStatementFromDecl(decl *ast.GenDecl) (*ast.BlockStmt) {
	var funcLit *ast.FuncLit
	valueSpec := decl.Specs[0].(*ast.ValueSpec)
	args := valueSpec.Values[0].(*ast.CallExpr).Args
	for _, node := range args {
		switch node := node.(type) {
		case *ast.FuncLit:
    	funcLit = node
	    break
  	}
	}

	return funcLit.Body
}

func gofmtFile(f *ast.File, fset *token.FileSet) ([]byte, error) {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
