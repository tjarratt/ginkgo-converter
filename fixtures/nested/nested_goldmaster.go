package nested

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSomethingLessImportant", func() {

			whatever := &UselessStruct{}
			t.Fail(whatever.ImportantField != "SECRET_PASSWORD")
		})
	})
}
