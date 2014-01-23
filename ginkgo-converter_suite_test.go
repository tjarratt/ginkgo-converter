package ginkgo-converter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGinkgo-Converter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ginkgo-Converter Suite")
}
