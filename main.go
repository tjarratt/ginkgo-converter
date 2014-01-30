package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var shouldOverwriteTests = flag.Bool("destructive", false, "rewrite tests in-place")
var shouldCreateTestSuite = flag.Bool("create-suite", true, "creates a ginkgo suite file")

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		println(fmt.Sprintf("usage: %s /path/to/some/file_test.go --destructive=(true|false)", os.Args[0]))
		println("\n--destructive indicates that you want to update your tests in-place, and can possibly lead to data loss if your tests are not committed to version control (or otherwise backed up)")
		os.Exit(1)
	}

	testFiles, err := findTestsForPackage(flag.Args()[0])
	if err != nil {
		fmt.Printf("unexpected error reading package: '%s'\n%s\n", flag.Args()[0], err.Error())
		os.Exit(1)
	}

	for _, filename := range testFiles {
		err := findTestsInFile(filename)

		if err != nil {
			panic(err.Error())
		}
	}

	pkg, err := build.Default.Import(flag.Args()[0], ".", build.ImportMode(0))
	if err != nil {
		panic(err)
	}

	if *shouldCreateTestSuite {
		addGinkgoSuiteFile(pkg.Dir)
	}
}

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

	var buffer bytes.Buffer
	if err = format.Node(&buffer, fileSet, rootNode); err != nil {
		println(err.Error())
		return
	}

	var fileToWrite string
	var mode os.FileMode
	if *shouldOverwriteTests {
		fileToWrite = pathToFile

		var fileInfo os.FileInfo
		fileInfo, err = os.Stat(pathToFile)
		if err != nil {
			return
		}

		mode = fileInfo.Mode()
	} else {
		fileToWrite = strings.Replace(pathToFile, "_test.go", "_ginkgo_test.go", 1)
		mode = 0644
	}

	ioutil.WriteFile(fileToWrite, buffer.Bytes(), mode)
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

func newMrTFromIdent(ident *ast.Ident) *ast.CallExpr {
	return &ast.CallExpr{
		Lparen: ident.NamePos + 1,
		Rparen: ident.NamePos + 2,
		Fun: &ast.Ident{Name: "T"},
	}
}

func replaceTestingTsInKeyValueExpression(kve *ast.KeyValueExpr, testingT string) {
	ident, ok := kve.Value.(*ast.Ident)
	if !ok {
		return
	}

	kve.Value = newMrTFromIdent(ident)
}

func replaceTestingTsInFuncLiteral(functionLiteral *ast.FuncLit, testingT string) {
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
			arg.Type = &ast.Ident{Name: "TestingT"}
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
			replaceTestingTsInKeyValueExpression(keyValueExpr, testingT)
			return true
		}

		funcLiteral, ok := node.(*ast.FuncLit)
		if ok {
			replaceTestingTsInFuncLiteral(funcLiteral, testingT)
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

			param.Type = &ast.Ident{Name: "TestingT"}
		}
	}
}
