package main

import (
	"errors"
	"fmt"
	"go/ast"
)

func importsForRootNode(rootNode *ast.File) (imports *ast.GenDecl, err error) {
	for _, declaration := range rootNode.Decls {
		decl, ok := declaration.(*ast.GenDecl)
		if !ok || len(decl.Specs) == 0 {
			continue
		}

		_, ok = decl.Specs[0].(*ast.ImportSpec)
		if ok {
			imports = decl
			return
		}
	}

	err = errors.New(fmt.Sprintf("Could not find imports for root node:\n\t%#v\n", rootNode))
	return
}

func removeTestingImport(rootNode *ast.File) {
	importDecl, err := importsForRootNode(rootNode)
	if err != nil {
		panic(err.Error())
	}

	var index int
	for i, importSpec := range importDecl.Specs {
		importSpec := importSpec.(*ast.ImportSpec)
		if importSpec.Path.Value == "\"testing\"" {
			index = i
			break
		}
	}

	importDecl.Specs = append(importDecl.Specs[:index], importDecl.Specs[index+1:]...)
}

func addGinkgoImports(rootNode *ast.File) {
	importDecl, err := importsForRootNode(rootNode)
	if err != nil {
		panic(err.Error())
	}

	if len(importDecl.Specs) == 0 {
		// TODO: might need to create a import decl here
		panic("unimplemented : expected to find an imports block")
	}

	needsGinkgo, needsMrT := true, true
	for _, importSpec := range importDecl.Specs {
		importSpec, ok := importSpec.(*ast.ImportSpec)
		if !ok {
			continue
		}

		if importSpec.Path.Value == "\"github.com/onsi/ginkgo\"" {
			needsGinkgo = false
		} else if importSpec.Path.Value == "\"github.com/tjarratt/mr_t\"" {
			needsMrT = false
		}
	}

	if needsGinkgo {
		importDecl.Specs = append(importDecl.Specs, createImport(".", "\"github.com/onsi/ginkgo\""))
	}

	if needsMrT {
		importDecl.Specs = append(importDecl.Specs, createImport("mr", "\"github.com/tjarratt/mr_t\""))
	}
}

func createImport(name, path string) *ast.ImportSpec {
	return &ast.ImportSpec{
		Name: &ast.Ident{Name: name},
		Path: &ast.BasicLit{Kind: 9, Value: path},
	}
}
