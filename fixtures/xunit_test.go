package fixtures

import (
	"testing"
)

type UselessStruct struct {
	ImportantField string
}

func TestSomethingImportant(t *testing.T) {
	whatever := &UselessStruct{}
	t.Fail(whatever.ImportantField != "SECRET_PASSWORD")
}
