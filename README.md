What's that you say? Ginkgo???
-----------------------

That's right! [Ginkgo](https://github.com/onsi/ginkgo) is a great TDD framework for Go. A lot of projects already have a relatively large test suite built up around the XUnit style test framework that Go ships with. With ginkgo-converter you can quickly convert your existing test suite to use Ginkgo without having to rewrite all of your tests by hand.

How does it work?
-----------------

ginkgo-gonverter rewrites your crummy old XUnit style tests into amazing Ginkgo tests:

old_boring_test.go
```go
package boring

import (
  "testing"
)

func TestSomethingImportant(t *testing.T) {
  whatever := &UselessStruct{}
  t.Fail(whatever.ImportantField != "SECRET_PASSWORD")
}
```

amazing_ginkgo_test.go
```go
package amazing

import (
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
```

Okay, but how does it really work?
----------------------------------

Ginkgo-Convert's secret sauce is the ast/parser format package that Go ships with. We read your tests, looking for matching `func TestXXXX(t *testing.T) { }` declarations and rewrite them as `It("TestXXXX", func() { })` blocks. This is currently only a proof-of-concept, but it should already be useful for converting tests.

What is left to do?
-----------------------
* "roundtrip" tests
* automatically import ginkgo and gomega
* take a package name to rewrite and recursively walk the directories in it
* create a test suite file for your package (shell out to ginkgo?)
* check for the presence of ginkgo && gomega?, ask the user to install it?
* handle t.Fail(), t.Error, t.Log, t.Skip, etc in the AST rewriter
* discover more edgecases in the AST rewriter (on-going)
