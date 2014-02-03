What's that you say? Ginkgo???
-----------------------

That's right! [Ginkgo](https://github.com/onsi/ginkgo) is a great BDD framework for Go. However, many projects already have a relatively large test suite built up around the XUnit style test framework that Go ships with. You can use ginkgo-converter to quickly convert your existing test suite to use Ginkgo without having to rewrite all of your tests by hand.

Getting Started
---------------
* `go get github.com/ginkgo/ginkgo`            # install the ginkgo cli tool
* `git submodule add github.com/onsi/ginkgo`   # the actual testrunner and library
* `git submodule add github.com/onsi/gomega`   # default matchers for ginkgo
* `git submodule add github.com/tjarratt/mr_t` # provides a *testing.T compatible interface
* `go install github.com/tjarratt/ginkgo-converter`
* `bin/ginkgo-converter your/package/name`
* now run your tests with `ginkgo`

How does it work?
-----------------

ginkgo-converter rewrites your crummy old XUnit style tests into amazing Ginkgo tests:

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

Ginkgo-Convert's secret sauce is the ast/parser format package that Go ships with. We read your tests, looking for matching `func TestXXXX(t *testing.T) { }` declarations and rewrite them as `It("TestXXXX", func() { })` blocks.

This has been tested against several codebases with 18K+ lines of code. So far the only hard problem to solve has been code that uses a *testing.T. If you're using a library that accepts an interface that *testing.T conforms to, then `ginkgo-convert` is smooth sailing.

What is left to do?
-----------------------
- issue PR to github.com/onsi/ginkgo
- discover more edgecases in the AST rewriter (on-going)
