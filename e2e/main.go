package main

import (
	"github.com/devspace-cloud/devspace/e2e/tests"
	"github.com/devspace-cloud/devspace/e2e/utils"
)

var testNamespace = "examples-test-namespace"

func main() {
	err := tests.RunQuickstart(testNamespace)
	utils.PrintTestResult("Quickstart", err)

	// err := tests.RunKustomize(testNamespace)
	// utils.PrintTestResult("Kustomize", err)
}
