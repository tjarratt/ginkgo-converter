package fixtures

import (
	. "github.com/onsi/ginkgo"
	mr "github.com/tjarratt/mr_t"
)

type UselessStruct struct {
	ImportantField string
	T              *testing.T
}

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSomethingImportant", func() {

			whatever := &UselessStruct{
				T:              mr.T(),
				ImportantField: "twisty maze of passages",
			}
			mr.T().Fail(whatever.ImportantField != "SECRET_PASSWORD")
			assert.Equal(mr.T(), whatever.ImportantField, "SECRET_PASSWORD")
			var foo = func(t TestingT) {}
			foo()
		})
	})
}
