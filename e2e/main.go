package main

import (
	"fmt"
	"os"

	"github.com/devspace-cloud/devspace/e2e/deploy"
	"github.com/devspace-cloud/devspace/e2e/testinit"
	"github.com/devspace-cloud/devspace/e2e/utils"
)

var testNamespace = "examples-test-namespace"

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// deployTest(pwd)
	testinit.TestInit(testNamespace, pwd)
}

func deployTest(pwd string) {

	err := deploy.RunQuickstart(testNamespace, pwd)
	utils.PrintTestResult("Quickstart", err)

	err = deploy.RunKustomize(testNamespace, pwd)
	utils.PrintTestResult("Kustomize", err)

	err = deploy.RunProfiles(testNamespace, pwd)
	utils.PrintTestResult("Profiles", err)

	// TODO: Need to reset helm client somehow
	// err = deploy.RunMicroservices(testNamespace, pwd)
	// utils.PrintTestResult("Microservices", err)

	// TODO: Need to reset helm client somehow
	// err = deploy.RunMinikube(testNamespace, pwd)
	// utils.PrintTestResult("Minikube", err)

	err = deploy.RunQuickstartKubectl(testNamespace, pwd)
	utils.PrintTestResult("Quickstart Kubectl", err)

	// TODO: Need to reset helm client somehow
	// err = deploy.RunPhpMysql(testNamespace, pwd)
	// utils.PrintTestResult("Php Mysql", err)
}
