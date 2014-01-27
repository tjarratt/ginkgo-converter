package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		println("Error: Not enough args. Expected path to test file")
		println(fmt.Sprintf("usage: %s /path/to/some/file_test.go", os.Args[0]))
		os.Exit(1)
	}

	testFiles, err := findTestsForPackage(os.Args[1])
		if err != nil {
		fmt.Printf("unexpected error reading package: %s\n%s\n", os.Args[1], err.Error())
		os.Exit(1)
	}

	for _, filename := range testFiles {
		err := findTestsInFile(filename)

		if err != nil {
			panic(err.Error())
		}
	}
}

func findTestsForPackage(packageName string) (tests []string, err error) {
	pkg, err := build.Default.Import(packageName, ".", build.ImportMode(0))
	if err != nil {
		return
	}

	for _, file := range pkg.TestGoFiles {
		tests = append(tests, filepath.Join(pkg.Dir, file))
	}

	dirFiles, err := ioutil.ReadDir(pkg.Dir)
	if err != nil {
		return
	}

	for _, file := range dirFiles {
		if !file.IsDir() {
			continue
		}

		moreTests, err := findTestsForPackage(filepath.Join(pkg.ImportPath, file.Name()))
		tests = append(tests, moreTests...)

		if err != nil {
			return tests, err
		}
	}

	return
}

func findTestsInFile(pathToFile string) (err error) {
	fileSet := token.NewFileSet()
	rootNode, err := parser.ParseFile(fileSet, pathToFile, nil, 0)
	if err != nil {
		return
	}

	addGinkgoImports(rootNode)

	testsToRewrite := findTestFuncs(rootNode)
	topLevelInitFunc := createInitBlock()

	describeBlock := createDescribeBlock()
	topLevelInitFunc.Body.List = append(topLevelInitFunc.Body.List, describeBlock)

	for _, testFunc := range testsToRewrite {
		rewriteTestInGinkgo(testFunc, rootNode, describeBlock)
	}

	rootNode.Decls = append(rootNode.Decls, topLevelInitFunc)

	var buffer bytes.Buffer
	if err = format.Node(&buffer, fileSet, rootNode); err != nil {
		println(err.Error())
		return
	}

	// TODO: take a flag to overwrite in place
	newFileName := strings.Replace(pathToFile, "_test.go", "_ginkgo_test.go", 1)
	ioutil.WriteFile(newFileName, buffer.Bytes(), 0666)
	return
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
	callExpr := &ast.CallExpr{Fun: itBlockIdent, Args: []ast.Expr{basicLit, funcLit}}
	expressionStatement := &ast.ExprStmt{X: callExpr}

	var block *ast.BlockStmt = blockStatementFromDescribe(describe)
	block.List = append(block.List, expressionStatement)
	rootNode.Decls = append(rootNode.Decls[:funcIndex], rootNode.Decls[funcIndex+1:]...)

	return
}

func blockStatementFromDescribe(desc *ast.ExprStmt) *ast.BlockStmt {
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

func addGinkgoImports(rootNode *ast.File) {
	var importDecl *ast.GenDecl
	for _, declaration := range rootNode.Decls {
		decl, ok := declaration.(*ast.GenDecl)
		if !ok || len(decl.Specs) == 0 {
			continue
		}

		_, ok = decl.Specs[0].(*ast.ImportSpec)
		if ok {
			importDecl = decl
			break
		}
	}

	if len(importDecl.Specs) == 0 {
		// TODO: might need to create a import decl here
		panic("unimplemented : expected to find an imports block")
	}

	needsGinkgo, needsGomega := true, true
	for _, importSpec := range importDecl.Specs {
		importSpec, ok := importSpec.(*ast.ImportSpec)
		if !ok {
			continue
		}

		if importSpec.Path.Value == "\"github.com/onsi/ginkgo\"" {
			needsGinkgo = false
		} else if importSpec.Path.Value == "\"github.com/onsi/gomega\"" {
			needsGomega = false
		}
	}

	if needsGinkgo {
		importGinkgo := createImport("\"github.com/onsi/ginkgo\"")
		importDecl.Specs = append(importDecl.Specs, importGinkgo)
	}

	if needsGomega {
		importGomega := createImport("\"github.com/onsi/gomega\"")
		importDecl.Specs = append(importDecl.Specs, importGomega)
	}
}

func createImport(path string) *ast.ImportSpec {
	return &ast.ImportSpec{
		Name: &ast.Ident{Name: "."},
		Path: &ast.BasicLit{Kind: 9, Value: path},
	}
}
