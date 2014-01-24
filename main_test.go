package main_test

import (
	"io/ioutil"
	"os"
  "os/exec"
	"path/filepath"
  . "github.com/onsi/ginkgo"
  . "github.com/onsi/gomega"
)

func init() {
  Describe("using ginkgo-converter", func() {
		BeforeEach(func() {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			cmd := exec.Command("rm", filepath.Join(cwd, "fixtures/xunit_ginkgo_test.go"))
			cmd.Run()
		})

    It("rewrites xunit tests as ginkgo tests", func() {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			pathToExecutable := filepath.Join(cwd, "ginkgo-convert")

      cmd := exec.Command(pathToExecutable, "github.com/tjarratt/ginkgo-convert/fixtures")
			err = cmd.Run()
      Expect(err).NotTo(HaveOccurred())

			convertedFile, err := ioutil.ReadFile(filepath.Join(cwd, "fixtures/xunit_ginkgo_test.go"))
			Expect(err).NotTo(HaveOccurred())

			goldMaster, err := ioutil.ReadFile(filepath.Join(cwd, "fixtures/ginkgo_test_goldmaster.go"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(convertedFile)).To(Equal(string(goldMaster)))
    })
  })
}
