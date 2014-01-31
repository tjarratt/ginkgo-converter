package nested

import (
	. "github.com/onsi/ginkgo"
	mr "github.com/tjarratt/mr_t"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSomethingLessImportant", func() {

			whatever := &UselessStruct{}
			mr.T().Fail(whatever.ImportantField != "SECRET_PASSWORD")
		})
	})
}
