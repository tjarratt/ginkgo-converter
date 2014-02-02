package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		println(fmt.Sprintf("usage: %s /path/to/some/file_test.go", os.Args[0]))
		os.Exit(1)
	}

	defer func() {
		err := recover()
		if err != nil {
			switch err := err.(type) {
			case error:
				println(err.Error())
			case string:
				println(err)
			default:
				println(fmt.Sprintf("unexpected error: %#v", err))
			}
			os.Exit(1)
		}
	}()

	RewritePackage(os.Args[1])
}
