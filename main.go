package main

import (
	"flag"
	"fmt"
	"go/build"
	"os"
)

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		println(fmt.Sprintf("usage: %s /path/to/some/file_test.go", os.Args[0]))
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

	addGinkgoSuiteFile(pkg.Dir)
}
