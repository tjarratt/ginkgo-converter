package creates_suites_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCreates_suites(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Creates_suites Suite")
}
