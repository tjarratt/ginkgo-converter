package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

/*
 * RewritePackage takes a name (eg: my-package/tools), finds its test files using
 * Go's build package, and then rewrites them. A ginkgo test suite file will
 * also be added for this package, and all of its child packages.
 */
func RewritePackage(packageName string) {
	pkg, err := build.Default.Import(packageName, ".", build.ImportMode(0))
	if err != nil {
		panic(fmt.Sprintf("unexpected error reading package: '%s'\n%s\n", os.Args[1], err.Error()))
	}

	for _, filename := range findTestsInPackage(pkg) {
		rewriteTestsInFile(filename)
	}
	return
}

/*
 * Given a package, findTestsInPackage reads the test files in the directory,
 * and then recurses on each child package, returning a slice of all test files
 * found in this process.
 */
func findTestsInPackage(pkg *build.Package) (testfiles []string) {
	for _, file := range append(pkg.TestGoFiles, pkg.XTestGoFiles...) {
		testfiles = append(testfiles, filepath.Join(pkg.Dir, file))
	}

	dirFiles, err := ioutil.ReadDir(pkg.Dir)
	if err != nil {
		panic(fmt.Sprintf("unexpected error reading dir: '%s'\n%s\n", pkg.Dir, err.Error()))
	}

	for _, file := range dirFiles {
		if !file.IsDir() {
			continue
		}

		packageName := filepath.Join(pkg.ImportPath, file.Name())
		subPackage, err := build.Default.Import(packageName, ".", build.ImportMode(0))
		if err != nil {
			panic(fmt.Sprintf("unexpected error reading package: '%s'\n%s\n", packageName, err.Error()))
		}

		testfiles = append(testfiles, findTestsInPackage(subPackage)...)
	}

	addGinkgoSuiteForPackage(pkg)
	return
}

/*
 * Shells out to `ginkgo bootstrap` to create a test suite file
 */
func addGinkgoSuiteForPackage(pkg *build.Package) {
	originalDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	suite_test_file := filepath.Join(pkg.Dir, pkg.Name+"_suite_test.go")
	_, err = os.Stat(suite_test_file)
	if err == nil {
		return // test file already exists, this should be a no-op
	}

	err = os.Chdir(pkg.Dir)
	if err != nil {
		panic(err)
	}

	output, err := exec.Command("ginkgo", "bootstrap").Output()

	if err != nil {
		panic(fmt.Sprintf("error running 'ginkgo bootstrap'.\n stdout: %s\n%s\n", output, err.Error()))
	}

	err = os.Chdir(originalDir)
	if err != nil {
		panic(err)
	}
}
