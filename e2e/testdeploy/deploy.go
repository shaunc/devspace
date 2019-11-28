package testdeploy

import (
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/devspace-cloud/devspace/pkg/util/factory"
)

type customFactory struct {
	*factory.DefaultFactoryImpl
	namespace string
	pwd       string

	FakeLogger *fakelog.FakeLogger
}

// GetLog implements interface
func (c *customFactory) GetLog() log.Logger {
	return c.FakeLogger
}

// TestDeploy starts the tests of the deploy cmd for the examples
func TestDeploy(ns string, pwd string) {
	myFactory := &customFactory{
		namespace: ns,
		pwd:       pwd,
	}
	myFactory.FakeLogger = fakelog.NewFakeLogger()

	err := RunQuickstart(myFactory)
	utils.PrintTestResult("Quickstart", err)

	// err = RunKustomize(*myFactory)
	// utils.PrintTestResult("Kustomize", err)

	// err = RunProfiles(*myFactory)
	// utils.PrintTestResult("Profiles", err)

	// TODO: Need to reset helm client somehow
	// err = RunMicroservices(*myFactory)
	// utils.PrintTestResult("Microservices", err)

	// TODO: Need to reset helm client somehow
	// err = RunMinikube(*myFactory)
	// utils.PrintTestResult("Minikube", err)

	// err = RunQuickstartKubectl(*myFactory)
	// utils.PrintTestResult("Quickstart Kubectl", err)

	// TODO: Need to reset helm client somehow
	// err = RunPhpMysql(*myFactory)
	// utils.PrintTestResult("Php Mysql", err)
}
