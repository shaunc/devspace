package main

import (
	"fmt"
	"github.com/devspace-cloud/devspace/e2e/testenter"
	"os"
)

var testNamespace = "examples-test-namespace"

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//testdeploy.TestDeploy(testNamespace, pwd)
	//testinit.TestInit(pwd)
	testenter.TestEnter(pwd, testNamespace)
}
