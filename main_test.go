package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

func init() {
	Describe("using ginkgo-converter", func() {
		BeforeEach(killAllConvertedGinkgoTests)
		AfterEach(killAllConvertedGinkgoTests)

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

func killAllConvertedGinkgoTests() {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	fixtures := filepath.Join(cwd, "fixtures")
	dirContents, err := ioutil.ReadDir(fixtures)
	Expect(err).NotTo(HaveOccurred())

	regex := regexp.MustCompile(".+_ginkgo_test.go")
	for _, file := range dirContents {
		if !regex.MatchString(file.Name()) {
			continue
		}

		pathToFile := filepath.Join(fixtures, file.Name())
		err = os.Remove(pathToFile)
		Expect(err).NotTo(HaveOccurred())
	}
}
