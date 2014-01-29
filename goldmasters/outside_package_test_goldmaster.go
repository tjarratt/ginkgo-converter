package fixtures_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/tjarratt/merf"
)

type UselessStruct struct {
	ImportantField string
}

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSomethingImportant", func() {

			whatever := &UselessStruct{}
			t.Fail(whatever.ImportantField != "SECRET_PASSWORD")
		})
	})
}