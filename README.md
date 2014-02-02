What's that you say? Ginkgo???
-----------------------

That's right! [Ginkgo](https://github.com/onsi/ginkgo) is a great BDD framework for Go. A lot of projects already have a relatively large test suite built up around the XUnit style test framework that Go ships with. With ginkgo-converter you can quickly convert your existing test suite to use Ginkgo without having to rewrite all of your tests by hand.

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
  . "github.com/onsi/gingko"
  . "github.com/onsi/gomega"
  mr "github.com/tjarratt/mr_t"
)

func init() {
  Describe("Using ginkgo", func() {
    It("TestSomethingImportant", func() {
      whatever := &UselessStruct{}
      mr.T().Fail(whatever.ImportantField != "SECRET_PASSWORD")
    })
  })
}
```

Okay, but how does it really work?
----------------------------------

Ginkgo-Convert's secret sauce is the ast/parser format package that Go ships with. We read your tests, looking for matching `func TestXXXX(t *testing.T) { }` declarations and rewrite them as `It("TestXXXX", func() { })` blocks. This is currently only a proof-of-concept, but it should already be useful for converting tests.

What is left to do?
-----------------------
- check for the presence of ginkgo && gomega?, ask the user to install it?
- better instructions for users switching from the unit testing framework
- consider converting non-test code files that use *testing.T
- get tests passing on travis-ci.org
- code cleanup
- issue PR to github.com/onsi/ginkgo
- discover more edgecases in the AST rewriter (on-going)
