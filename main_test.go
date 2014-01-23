package main_test

import (
  "os/exec"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
)

func init() {
  Describe("using ginkgo-converter", func() {
    It("rewrites xunit tests as ginkgo tests", func() {
      cmd := exec.Command("ginkgo-converter", "test/fixtures/xunit_test.go")
      err := cmd.Run()

      Expect(err).NotTo(HaveOccured())

      bytes, err := cmd.Output()
      Expect(err).NotTo(HaveOccured())
      println(string(bytes))
    })
  })
}
