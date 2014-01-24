package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestGinkgoConverter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ginkgo-Converter Suite")
}
