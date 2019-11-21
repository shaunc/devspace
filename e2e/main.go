package main

import (
	"fmt"
	"os"

	"github.com/devspace-cloud/devspace/e2e/tests"
	"github.com/devspace-cloud/devspace/e2e/utils"
)

var testNamespace = "examples-test-namespace"

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// err := tests.RunQuickstart(testNamespace, pwd)
	// utils.PrintTestResult("Quickstart", err)

	// err = tests.RunKustomize(testNamespace, pwd)
	// utils.PrintTestResult("Kustomize", err)

	err = tests.RunProfiles(testNamespace, pwd)
	utils.PrintTestResult("Profiles", err)

	// err := tests.RunMicroservices(testNamespace, pwd)
	// utils.PrintTestResult("Microservices", err)

	// err := tests.RunMinikube(testNamespace, pwd)
	// utils.PrintTestResult("Minikube", err)

	// err := tests.RunQuickstartKubectl(testNamespace, pwd)
	// utils.PrintTestResult("Quickstart Kubectl", err)

	// Portforwarding not working
	// err := tests.RunPhpMysql(testNamespace, pwd)
	// utils.PrintTestResult("Php Mysql", err)
}
