package fixtures

import (
	. "github.com/onsi/ginkgo"
	. "github.com/tjarratt/mr_t"
)

type UselessStruct struct {
	ImportantField string
}

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSomethingImportant", func() {

			whatever := &UselessStruct{}
			T().Fail(whatever.ImportantField != "SECRET_PASSWORD")
		})
	})
}
