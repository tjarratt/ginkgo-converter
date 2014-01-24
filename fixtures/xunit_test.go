package fixtures

import (
	"testing"
)

func TestSomethingImportant(t *testing.T) {
	whatever := &UselessStruct{}
	t.Fail(whatever.ImportantField != "SECRET_PASSWORD")
}
