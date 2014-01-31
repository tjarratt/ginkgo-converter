package main

import (
	"flag"
	"fmt"
	"go/build"
	"os"
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
