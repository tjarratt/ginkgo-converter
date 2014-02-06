package nested

import (
	. "github.com/onsi/ginkgo"
	mr "github.com/tjarratt/mr_t"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("something less important", func() {

			whatever := &UselessStruct{}
			mr.T().Fail(whatever.ImportantField != "SECRET_PASSWORD")
		})
	})
}
