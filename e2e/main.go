package main

import (
	"github.com/devspace-cloud/devspace/e2e/tests"
	"github.com/devspace-cloud/devspace/e2e/utils"
)

var testNamespace = "examples-test-namespace"

func main() {
	// err := tests.RunQuickstart(testNamespace)
	// utils.PrintTestResult("Quickstart", err)

	// err = tests.RunKustomize(testNamespace)
	// utils.PrintTestResult("Kustomize", err)

	// err := tests.RunProfiles(testNamespace)
	// utils.PrintTestResult("Profiles", err)

	// err := tests.RunMicroservices(testNamespace)
	// utils.PrintTestResult("Microservices", err)

	err := tests.RunMinikube(testNamespace)
	utils.PrintTestResult("Minikube", err)

	// err := tests.RunQuickstartKubectl(testNamespace)
	// utils.PrintTestResult("Quickstart Kubectl", err)

	// Portforwarding not working
	// err := tests.RunPhpMysql(testNamespace)
	// utils.PrintTestResult("Php Mysql", err)
}
