package testdeploy

import (
	"github.com/devspace-cloud/devspace/cmd"
	"github.com/devspace-cloud/devspace/cmd/flags"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/devspace-cloud/devspace/pkg/devspace/services"
	"github.com/devspace-cloud/devspace/pkg/util/factory"
	"github.com/devspace-cloud/devspace/pkg/util/log"
	fakelog "github.com/devspace-cloud/devspace/pkg/util/log/testing"
	"github.com/pkg/errors"
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

	err = RunKustomize(myFactory)
	utils.PrintTestResult("Kustomize", err)

	err = RunProfiles(myFactory)
	utils.PrintTestResult("Profiles", err)

	err = RunMicroservices(myFactory)
	utils.PrintTestResult("Microservices", err)

	err = RunMinikube(myFactory)
	utils.PrintTestResult("Minikube", err)

	err = RunQuickstartKubectl(myFactory)
	utils.PrintTestResult("Quickstart Kubectl", err)

	err = RunPhpMysql(myFactory)
	utils.PrintTestResult("Php Mysql", err)

	err = RunDependencies(myFactory)
	utils.PrintTestResult("Dependencies", err)

	//err := RunKaniko(myFactory)
	//utils.PrintTestResult("Kaniko", err)
}

func RunTest(f *customFactory, dir string, deployConfig *cmd.DeployCmd) error {
	if deployConfig == nil {
		deployConfig = &cmd.DeployCmd{
			GlobalFlags: &flags.GlobalFlags{
				Namespace: f.namespace,
				NoWarn:    true,
			},
			ForceBuild:  true,
			ForceDeploy: true,
			SkipPush:    true,
		}
	}

	err := utils.ChangeWorkingDir(f.pwd + "/../examples/" + dir)
	if err != nil {
		return err
	}

	// Create kubectl client
	client, err := f.NewKubeClientFromContext(deployConfig.KubeContext, deployConfig.Namespace, deployConfig.SwitchContext)
	if err != nil {
		return errors.Errorf("Unable to create new kubectl client: %v", err)
	}

	// At last, we delete the current namespace
	defer utils.DeleteNamespaceAndWait(client, deployConfig.Namespace)

	err = deployConfig.Run(f, nil, nil)
	if err != nil {
		return err
	}

	// Checking if pods are running correctly
	err = utils.AnalyzePods(client, f.namespace)
	if err != nil {
		return err
	}

	// Load generated config
	generatedConfig, err := f.NewConfigLoader(nil, nil).Generated()
	if err != nil {
		return errors.Errorf("Error loading generated.yaml: %v", err)
	}

	// Add current kube context to context
	configOptions := deployConfig.ToConfigOptions()
	config, err := f.NewConfigLoader(configOptions, f.GetLog()).Load()
	if err != nil {
		return err
	}

	servicesClient := services.NewClient(config, generatedConfig, client, nil, log.GetInstance())

	// Port-forwarding
	err = utils.PortForwardAndPing(servicesClient)
	if err != nil {
		return err
	}

	return nil
}
