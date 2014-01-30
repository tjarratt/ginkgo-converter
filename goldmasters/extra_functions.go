package fixtures

import (
	. "github.com/onsi/ginkgo"
	. "github.com/tjarratt/mr_t"
)

func somethingImportant(t TestingT, message *string) {
	t.Log("Something important happened in a test: " + *message)
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSomethingLessImportant", func() {
			somethingImportant(T(), &"hello!")
		})
	})
}
