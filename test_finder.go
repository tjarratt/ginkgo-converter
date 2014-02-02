package main

import (
	"bytes"
	"fmt"
	"go/build"
	"go/format"
	"go/token"
	"go/parser"
	"io/ioutil"
	"os"
	"path/filepath"
	"go/ast"
	"regexp"
)

func findTestsForPackage(packageName string) (tests []string, err error) {
	pkg, err := build.Default.Import(packageName, ".", build.ImportMode(0))
	if err != nil {
		return
	}

	for _, file := range append(pkg.TestGoFiles, pkg.XTestGoFiles...) {
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
	removeTestingImport(rootNode)

	testsToRewrite := findTestFuncs(rootNode)
	topLevelInitFunc := createInitBlock()

	describeBlock := createDescribeBlock()
	topLevelInitFunc.Body.List = append(topLevelInitFunc.Body.List, describeBlock)

	for _, testFunc := range testsToRewrite {
		rewriteTestInGinkgo(testFunc, rootNode, describeBlock)
	}

	rootNode.Decls = append(rootNode.Decls, topLevelInitFunc)
	rewriteOtherFuncsToUseMrT(rootNode.Decls)
	walkNodesInRootNodeReplacingTestingT(rootNode)

	var buffer bytes.Buffer
	if err = format.Node(&buffer, fileSet, rootNode); err != nil {
		println(err.Error())
		return
	}

	fileInfo, err := os.Stat(pathToFile)
	if err != nil {
		return
	}

	mode := fileInfo.Mode()

	ioutil.WriteFile(pathToFile, buffer.Bytes(), mode)
	return
}

func findTestFuncs(rootNode *ast.File) (testsToRewrite []*ast.FuncDecl) {
	testNameRegexp := regexp.MustCompile("^Test[A-Z].+")

	ast.Inspect(rootNode, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		switch node := node.(type) {
		case *ast.FuncDecl:
			matches := testNameRegexp.MatchString(node.Name.Name)

			if matches && receivesTestingT(node) {
				testsToRewrite = append(testsToRewrite, node)
			}
		}

		return true
	})

	return
}

// name is on field.Name.Name
// type is on field.Type.X.X.Name + "." + field.Type.X.Sel.Name
func receivesTestingT(node *ast.FuncDecl) bool {
	if len(node.Type.Params.List) != 1 {
		return false
	}

	// (node.Type.Params.List[0].Names[0].Name == "t") => the name of the testingT arg
	base, ok := node.Type.Params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}

	intermediate := base.X.(*ast.SelectorExpr)
	isTestingPackage := intermediate.X.(*ast.Ident).Name == "testing"
	isTestingT := intermediate.Sel.Name == "T"

	return isTestingPackage && isTestingT
}

func namedTestingTArg(node *ast.FuncDecl) string {
	return node.Type.Params.List[0].Names[0].Name // *exhale*
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
	replaceTestingTsWithMrT(block, namedTestingTArg(testFunc))

	// remove the old test func from the root node
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

func walkNodesInRootNodeReplacingTestingT(rootNode *ast.File) {
	ast.Inspect(rootNode, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		switch node := node.(type) {
		case *ast.StructType:
			replaceTestingTsInStructType(node)
		case *ast.FuncLit:
			replaceTypeDeclTestingTsInFuncLiteral(node)
		}

		return true
	})
}

func rewriteOtherFuncsToUseMrT(declarations []ast.Decl) {
	for _, decl := range declarations {
		decl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		for _, param := range decl.Type.Params.List {
			starExpr, ok := param.Type.(*ast.StarExpr)
			if !ok {
				continue
			}

			selectorExpr, ok := starExpr.X.(*ast.SelectorExpr)
			if !ok {
				continue
			}

			xIdent, ok := selectorExpr.X.(*ast.Ident)
			if !ok || xIdent.Name != "testing" {
				continue
			}

			if selectorExpr.Sel.Name != "T" {
				continue
			}

			param.Type = newMrTestingT()
		}
	}
}

func replaceTestingTsInArgsLists(callExpr *ast.CallExpr, testingT string) {
	for index, arg := range callExpr.Args {
		ident, ok := arg.(*ast.Ident)
		if !ok {
			continue
		}

		if ident.Name == testingT {
			callExpr.Args[index] = newMrTFromIdent(ident)
		}
	}
}

func replaceNamedTestingTsInKeyValueExpression(kve *ast.KeyValueExpr, testingT string) {
	ident, ok := kve.Value.(*ast.Ident)
	if !ok {
		return
	}

	if ident.Name == testingT {
		kve.Value = newMrTFromIdent(ident)
	}
}


func replaceTypeDeclTestingTsInFuncLiteral(functionLiteral *ast.FuncLit) {
	for _, arg := range functionLiteral.Type.Params.List {
		starExpr, ok := arg.Type.(*ast.StarExpr)
		if !ok {
			continue
		}

		selectorExpr, ok := starExpr.X.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		target, ok := selectorExpr.X.(*ast.Ident)
		if !ok {
			continue
		}

		if target.Name == "testing" && selectorExpr.Sel.Name == "T" {
			arg.Type = newMrTestingT()
		}
	}
}

func replaceTestingTsInStructType(structType *ast.StructType) {
	for _, field := range structType.Fields.List {
		starExpr, ok := field.Type.(*ast.StarExpr)
		if !ok {
			continue
		}

		selectorExpr, ok := starExpr.X.(*ast.SelectorExpr)
		if !ok {
			continue
		}

		xIdent, ok := selectorExpr.X.(*ast.Ident)
		if !ok {
			continue
		}

		if xIdent.Name == "testing" && selectorExpr.Sel.Name == "T" {
			field.Type = newMrTestingT()
		}
	}
}

func replaceTestingTsWithMrT(statementsBlock *ast.BlockStmt, testingT string) {
	ast.Inspect(statementsBlock, func(node ast.Node) bool {
		if node == nil {
			return false
		}

		keyValueExpr, ok := node.(*ast.KeyValueExpr)
		if ok {
			replaceNamedTestingTsInKeyValueExpression(keyValueExpr, testingT)
			return true
		}

		funcLiteral, ok := node.(*ast.FuncLit)
		if ok {
			replaceTypeDeclTestingTsInFuncLiteral(funcLiteral)
			return true
		}


		callExpr, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		funCall, ok := callExpr.Fun.(*ast.SelectorExpr)
		if ok {
			replaceTestingTsMethodCalls(funCall, testingT)
			replaceTestingTsInArgsLists(callExpr, testingT)
			return true
		}

		for index, arg := range callExpr.Args {
			ident, ok := arg.(*ast.Ident)
			if !ok {
				return true
			}

			if ident.Name == testingT {
				callExpr.Args[index] = newMrTFromIdent(ident)
			}
		}

		return true
	})
}

func replaceTestingTsMethodCalls(selectorExpr *ast.SelectorExpr, testingT string) {
	ident, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		return
	}

	// fix t.Fail() or any other *testing.T method calls
	// by replacing with T().Fail()
	if ident.Name == testingT {
		selectorExpr.X = newMrTFromIdent(ident)
	}
}
