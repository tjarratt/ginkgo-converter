package amazing

import (
	. "github.com/onsi/gingko"
	. "github.com/onsi/gomega"
	"testing"
)

func init() {
	Describe("Using ginkgo", func() {
		It("TestSomethingImportant", func() {
			whatever := &UselessStruct{}
			t.Fail(whatever.ImportantField != "SECRET_PASSWORD")
		})
	})
}
