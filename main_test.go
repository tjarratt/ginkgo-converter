package main_test

import (
	"crypto/rand"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

func init() {
	Describe("using ginkgo-converter", func() {
		BeforeEach(killAllConvertedGinkgoTests)
		AfterEach(killAllConvertedGinkgoTests)

		It("rewrites xunit tests as ginkgo tests", func() {
			runGinkgoConvert("--create-suite=false", "github.com/tjarratt/ginkgo-convert/fixtures")

			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			pathToFile := filepath.Join(cwd, "fixtures", "xunit_ginkgo_test.go")
			convertedFile, err := ioutil.ReadFile(pathToFile)
			Expect(err).NotTo(HaveOccurred())

			goldMaster, err := ioutil.ReadFile(filepath.Join(cwd, "goldmasters", "simple_test_goldmaster.go"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(convertedFile)).To(Equal(string(goldMaster)))
		})

		It("rewrites all tests in your package", func() {
			runGinkgoConvert("--create-suite=false", "github.com/tjarratt/ginkgo-convert/fixtures")

			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			pathToFile := filepath.Join(cwd, "fixtures", "nested", "nested_ginkgo_test.go")
			convertedFile, err := ioutil.ReadFile(pathToFile)
			Expect(err).NotTo(HaveOccurred())

			goldMaster, err := ioutil.ReadFile(filepath.Join(cwd, "goldmasters", "nested_goldmaster.go"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(convertedFile)).To(Equal(string(goldMaster)))
		})

		It("overwrites tests when the --destructive flag is provided", func() {
			withTempDir("overwrite", func(dir string) {
				cwd, err := os.Getwd()
				Expect(err).NotTo(HaveOccurred())

				bytes, err := ioutil.ReadFile(filepath.Join(cwd, "fixtures", "xunit_test.go"))
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(dir, "overwrite_test.go"), bytes, 0666)
				Expect(err).NotTo(HaveOccurred())

				runGinkgoConvert("--create-suite=false", "--destructive", "github.com/tjarratt/ginkgo-convert/fixtures/overwrite")

				expected, err := ioutil.ReadFile(filepath.Join(cwd, "goldmasters", "simple_test_goldmaster.go"))
				Expect(err).NotTo(HaveOccurred())

				converted, err := ioutil.ReadFile(filepath.Join(dir, "overwrite_test.go"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(converted)).To(Equal(string(expected)))
			})
		})

		It("creates a ginkgo test suite file when provided", func() {
			withTempDir("creates_suites", func(dir string) {
				cwd, err := os.Getwd()
				Expect(err).NotTo(HaveOccurred())

				bytes, err := ioutil.ReadFile(filepath.Join(cwd, "fixtures", "xunit_test.go"))
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(dir, "creates_suite_test.go"), bytes, 0666)
				Expect(err).NotTo(HaveOccurred())

				runGinkgoConvert("github.com/tjarratt/ginkgo-convert/fixtures/creates_suites")

				expected, err := ioutil.ReadFile(filepath.Join(cwd, "goldmasters", "test_suite_goldmaster.go"))
				Expect(err).NotTo(HaveOccurred())

				testSuite, err := ioutil.ReadFile(filepath.Join(dir, "creates_suites_suite_test.go"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(testSuite)).To(Equal(string(expected)))
			})
		})
	})
}

func withTempDir(dirname string, cb func(tmpDir string)) {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	tmpDir := filepath.Join(cwd, "fixtures", dirname)
	err = os.MkdirAll(tmpDir, os.ModeDir|os.ModeTemporary|os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	defer func() {
		err = os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	}()

	cb(tmpDir)
}

func uniqueKey(namePrefix string) string {
	salt, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt32))
	if err != nil {
		salt = big.NewInt(1)
	}

	return fmt.Sprintf("%s_%d_%d", namePrefix, time.Now().Unix(), salt)
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

func rewriteTestsInPackage(packageToRewrite string) {
	runGinkgoConvert(packageToRewrite)
	return
}

func overwriteTestsInPackage(packageToRewrite string) {
	runGinkgoConvert("--destructive", packageToRewrite)
	return
}

func runGinkgoConvert(args ...string) {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	pathToExecutable := filepath.Join(cwd, "bin", "ginkgo-convert")
	cmd := exec.Command(pathToExecutable, args...)
	err = cmd.Run()
	Expect(err).NotTo(HaveOccurred())
}
