package fixtures

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSomethingImportant", func() {

			whatever := &UselessStruct{}
			t.Fail(whatever.ImportantField != "SECRET_PASSWORD")
		})
	})
}
