package testenter

import (
	"fmt"
	"github.com/devspace-cloud/devspace/cmd"
	"github.com/devspace-cloud/devspace/cmd/flags"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/devspace-cloud/devspace/pkg/devspace/docker"
	"github.com/devspace-cloud/devspace/pkg/util/factory"
	"github.com/devspace-cloud/devspace/pkg/util/log"
	fakelog "github.com/devspace-cloud/devspace/pkg/util/log/testing"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type customFactory struct {
	*factory.DefaultFactoryImpl

	FakeLogger *fakelog.FakeLogger
}

// NewDockerClient implements interface
func (c *customFactory) NewDockerClient(log log.Logger) (docker.Client, error) {
	fakeDockerClient := &docker.FakeClient{
		AuthConfig: &dockertypes.AuthConfig{
			Username: "user",
			Password: "pass",
		},
	}
	return fakeDockerClient, nil
}

// GetLog implements interface
func (c *customFactory) GetLog() log.Logger {
	return c.FakeLogger
}

func TestEnter(pwd string, ns string) {
	err := runTest(pwd, ns)
	if err != nil {
		utils.PrintTestResult("Enter Test", err)
	}
}

func runTest(pwd string, ns string) error {
	f := &customFactory{}
	f.FakeLogger = fakelog.NewFakeLogger()

	deployConfig := &cmd.DeployCmd{
		GlobalFlags: &flags.GlobalFlags{
			Namespace: ns,
			NoWarn:    true,
		},
		ForceBuild:  true,
		ForceDeploy: true,
		SkipPush:    true,
	}

	enterConfig := &cmd.EnterCmd{
		GlobalFlags: &flags.GlobalFlags{
			Namespace: ns,
			NoWarn:    true,
		},
		Wait: true,
	}

	err := utils.ChangeWorkingDir(pwd + "/../examples/quickstart")
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
	err = utils.AnalyzePods(client, ns)
	if err != nil {
		return err
	}

	pods, err := client.KubeClient().CoreV1().Pods(ns).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	enterConfig.Pod = pods.Items[0].Name

	err = enterConfig.Run(f, nil, []string{"echo", "blabla"})
	if err != nil {
		return err
	}

	return nil
}
