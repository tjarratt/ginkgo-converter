package main

import (
	. "convert"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		println(fmt.Sprintf("usage: %s /path/to/some/file_test.go", os.Args[0]))
		os.Exit(1)
	}

	testFiles, err := RewritePackage(os.Args[1])
	if err != nil {
		fmt.Printf("unexpected error reading package: '%s'\n%s\n", os.Args[1], err.Error())
		os.Exit(1)
	}

	for _, filename := range testFiles {
		err := RewriteTestsInFile(filename)

		if err != nil {
			fmt.Printf("unexpected error rewriting tests in file: %s\n%s", filename, err.Error())
			os.Exit(1)
		}
	}
}
