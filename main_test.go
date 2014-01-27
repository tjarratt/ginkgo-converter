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

			rewriteTestsInPackage("github.com/tjarratt/ginkgo-convert/fixtures")
			pathToFile := filepath.Join(cwd, "fixtures/xunit_ginkgo_test.go")
			convertedFile, err := ioutil.ReadFile(pathToFile)
			Expect(err).NotTo(HaveOccurred())

			goldMaster, err := ioutil.ReadFile(filepath.Join(cwd, "fixtures/ginkgo_test_goldmaster.go"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(convertedFile)).To(Equal(string(goldMaster)))
		})

		It("rewrites all tests in your package", func() {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			rewriteTestsInPackage("github.com/tjarratt/ginkgo-convert/fixtures")
			pathToFile := filepath.Join(cwd, "fixtures/nested/nested_ginkgo_test.go")
			convertedFile, err := ioutil.ReadFile(pathToFile)
			Expect(err).NotTo(HaveOccurred())

			goldMaster, err := ioutil.ReadFile(filepath.Join(cwd, "fixtures/nested/nested_goldmaster.go"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(convertedFile)).To(Equal(string(goldMaster)))
		})
	})
}

func killAllConvertedGinkgoTests() {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	fixtures := filepath.Join(cwd, "fixtures")
	killTestsInDir(fixtures)
}

func killTestsInDir(dir string) {
	dirContents, err := ioutil.ReadDir(dir)
	Expect(err).NotTo(HaveOccurred())

	regex := regexp.MustCompile(".+_ginkgo_test.go")
	for _, file := range dirContents {
		if file.IsDir() {
			killTestsInDir(filepath.Join(dir, file.Name()))
			continue
		}

		if !regex.MatchString(file.Name()) {
			continue
		}

		pathToFile := filepath.Join(dir, file.Name())
		err = os.Remove(pathToFile)
		Expect(err).NotTo(HaveOccurred())
	}
}

func rewriteTestsInPackage(packageToRewrite string) (convertedFile []byte) {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	pathToExecutable := filepath.Join(cwd, "ginkgo-convert")
	cmd := exec.Command(pathToExecutable, packageToRewrite)
	err = cmd.Run()
	Expect(err).NotTo(HaveOccurred())

	return
}
