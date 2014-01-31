package fixtures

import (
	. "github.com/onsi/ginkgo"
	mr "github.com/tjarratt/mr_t"
)

func somethingImportant(t mr.TestingT, message *string) {
	t.Log("Something important happened in a test: " + *message)
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSomethingLessImportant", func() {
			somethingImportant(mr.T(), &"hello!")
		})
	})
}
