package tmp_test

import (
	. "github.com/onsi/ginkgo"
	mr "github.com/tjarratt/mr_t"
)

type UselessStruct struct {
	ImportantField string
}

func init() {
	Describe("Testing with ginkgo", func() {
		It("something important", func() {

			whatever := &UselessStruct{}
			mr.T().Fail(whatever.ImportantField != "SECRET_PASSWORD")
		})
	})
}
