package nested

import (
	. "github.com/onsi/ginkgo"
	. "github.com/tjarratt/mr_t"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSomethingLessImportant", func() {

			whatever := &UselessStruct{}
			T().Fail(whatever.ImportantField != "SECRET_PASSWORD")
		})
	})
}
