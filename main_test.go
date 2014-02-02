package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

func init() {
	Describe("using ginkgo-converter", func() {
		BeforeEach(deleteFilesInTmp)

		It("rewrites xunit tests as ginkgo tests", func() {
			withTempDir(func(tempDir string) {
				runGinkgoConvert()

				convertedFile := readConvertedFileNamed(tempDir, "xunit_test.go")
				goldmaster := readGoldMasterNamed("xunit_test.go")
				Expect(convertedFile).To(Equal(goldmaster))
			})
		})

		It("rewrites all usages of *testing.T as mr.T()", func() {
			withTempDir(func(tempDir string) {
				runGinkgoConvert()

				convertedFile := readConvertedFileNamed(tempDir, "extra_functions_test.go")
				goldmaster := readGoldMasterNamed("extra_functions_test.go")
				Expect(convertedFile).To(Equal(goldmaster))
			})
		})

		It("rewrites tests in the package dir that belong to other packages", func() {
			withTempDir(func(tempDir string) {
				runGinkgoConvert()

				convertedFile := readConvertedFileNamed(tempDir, "outside_package_test.go")
				goldMaster := readGoldMasterNamed("outside_package_test.go")
				Expect(convertedFile).To(Equal(goldMaster))
		  })
		})

		It("rewrites tests in nested packages", func() {
			withTempDir(func(dir string) {
				runGinkgoConvert()

				convertedFile := readConvertedFileNamed(dir, filepath.Join("nested", "nested_test.go"))
				goldMaster := readGoldMasterNamed("nested_test.go")
				Expect(convertedFile).To(Equal(goldMaster))
			})
		})

		Context("ginkgo test suite files", func() {
			It("creates a ginkgo test suite file for the package you specified", func() {
				withTempDir(func(dir string) {
					runGinkgoConvert()

					testsuite := readConvertedFileNamed(dir, "tmp_suite_test.go")
					goldmaster := readGoldMasterNamed("suite_test.go")
					Expect(testsuite).To(Equal(goldmaster))
				})
			})

			It("creates ginkgo test suites for all nested packages", func() {
				withTempDir(func(dir string) {
					runGinkgoConvert()

					testsuite := readConvertedFileNamed(dir, "nested", "nested_suite_test.go")
					goldmaster := readGoldMasterNamed("nested_suite_test.go")
					Expect(testsuite).To(Equal(goldmaster))
				})
			})

			It("gracefully handles existing test suite files", func() {
				withTempDir(func (dir string) {
					cwd, err := os.Getwd()
					bytes, err := ioutil.ReadFile(filepath.Join(cwd, "goldmasters", "fixtures_suite_test.go"))
					Expect(err).NotTo(HaveOccurred())
					err = ioutil.WriteFile(filepath.Join(cwd, "tmp", "tmp_suite_test.go"), bytes, 0600)
					Expect(err).NotTo(HaveOccurred())

					runGinkgoConvert()
				})
			})
		})
	})
}

func withTempDir(cb func(tempDir string)) {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	tempDir := filepath.Join(cwd, "tmp")
	err = os.MkdirAll(tempDir, os.ModeDir|os.ModeTemporary|os.ModePerm)
	Expect(err).NotTo(HaveOccurred())

	copyFixturesIntoTempDir("fixtures", tempDir)

	cb(tempDir)
}

func copyFixturesIntoTempDir(relativePathToFixtures, tempDir string) {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	files, err := ioutil.ReadDir(filepath.Join(cwd, relativePathToFixtures))
	Expect(err).NotTo(HaveOccurred())

	for _, f := range files {
		if f.IsDir() {
			nestedFixturesDir := filepath.Join(relativePathToFixtures, f.Name())
			nestedTempDir := filepath.Join(tempDir, f.Name())
			err = os.MkdirAll(nestedTempDir, os.ModeDir|os.ModeTemporary|os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			copyFixturesIntoTempDir(nestedFixturesDir, nestedTempDir)
			continue
		}

		bytes, err := ioutil.ReadFile(filepath.Join(cwd, relativePathToFixtures, f.Name()))
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(tempDir, f.Name()), bytes, 0600)
		Expect(err).NotTo(HaveOccurred())
	}
}

func runGinkgoConvert() {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	pathToExecutable := filepath.Join(cwd, "bin", "ginkgo-convert")
	cmd := exec.Command(pathToExecutable, "github.com/tjarratt/ginkgo-convert/tmp")
	out, err := cmd.Output()

	if err != nil {
		println("ginkgo-convert failed:", string(out))
	}
	Expect(err).NotTo(HaveOccurred())
}

func readGoldMasterNamed(filename string) string {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	bytes, err := ioutil.ReadFile(filepath.Join(cwd, "goldmasters", filename))
	Expect(err).NotTo(HaveOccurred())

	return string(bytes)
}

func readConvertedFileNamed(pathComponents ...string) string {
	pathToFile := filepath.Join(pathComponents...)
	bytes, err := ioutil.ReadFile(pathToFile)
	Expect(err).NotTo(HaveOccurred())

	return string(bytes)
}

func deleteFilesInTmp() {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	tempDir := filepath.Join(cwd, "tmp")

 	err = os.RemoveAll(tempDir)
 	Expect(err).NotTo(HaveOccurred())
}
