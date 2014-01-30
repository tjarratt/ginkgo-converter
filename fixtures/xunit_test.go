package fixtures

import (
	"testing"
)

type UselessStruct struct {
	ImportantField string
	T              *testing.T
}

func TestSomethingImportant(t *testing.T) {
	whatever := &UselessStruct{
		T:            t,
		ImportantField: "twisty maze of passages",
	}
	t.Fail(whatever.ImportantField != "SECRET_PASSWORD")
	assert.Equal(t, whatever.ImportantField, "SECRET_PASSWORD")
	var foo = func(t *testing.T) {}
	foo()
}
