package init

import (
	"fmt"
	"reflect"

	"github.com/devspace-cloud/devspace/pkg/util/factory"

	"github.com/devspace-cloud/devspace/cmd"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/devspace-cloud/devspace/pkg/devspace/build/builder/helper"

	"github.com/devspace-cloud/devspace/pkg/devspace/config/versions/latest"
	"github.com/devspace-cloud/devspace/pkg/util/log"

	"github.com/devspace-cloud/devspace/pkg/devspace/docker"
	fakelog "github.com/devspace-cloud/devspace/pkg/util/log/testing"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

type initTestCase struct {
	name    string
	answers []string

	expectedConfig *latest.Config
}

type customFactory struct {
	*factory.DefaultFactoryImpl

	FakeLogger *fakelog.FakeLogger
	// Config          *latest.Config
	// GeneratedConfig *generated.Config
}

// func (c *customFactory) NewConfigLoader(options *loader.ConfigOptions, log log.Logger) loader.ConfigLoader {
// 	return fakeconfigloader.NewFakeConfigLoader(c.GeneratedConfig, c.Config, log)
// }

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

// TestInit runs the e2e tests for the init cmd
func TestInit(pwd string) {
	myFactory := &customFactory{}
	myFactory.FakeLogger = fakelog.NewFakeLogger()

	err := CreateDockerfile(myFactory, pwd)
	utils.PrintTestResult("Create Dockerfile", err)

	err = UseExistingDockerfile(myFactory, pwd)
	utils.PrintTestResult("Use Existing Dockerfile", err)

	err = UseDockerfile(myFactory, pwd)
	utils.PrintTestResult("Use Dockerfile", err)

	err = UseManifests(myFactory, pwd)
	utils.PrintTestResult("Use Kubectl Manifests", err)

	err = UseChart(myFactory, pwd)
	utils.PrintTestResult("Use Helm Chart", err)
}

func initializeTest(f *customFactory, testCase initTestCase) error {
	initConfig := cmd.InitCmd{
		Dockerfile:  helper.DefaultDockerfilePath,
		Reconfigure: false,
		Context:     "",
		Provider:    "",
	}

	c, err := f.NewDockerClient(f.GetLog())
	if err != nil {
		return err
	}
	docker.SetFakeClient(c)

	for _, a := range testCase.answers {
		fmt.Println("SetNextAnswer:", a)
		f.FakeLogger.Survey.SetNextAnswer(a)
	}

	// runs init cmd
	err = initConfig.Run(f, nil, nil)
	if err != nil {
		return err
	}

	if testCase.expectedConfig != nil {
		config, err := f.NewConfigLoader(nil, nil).Load()
		if err != nil {
			return err
		}

		isEqual := reflect.DeepEqual(config, testCase.expectedConfig)
		if !isEqual {
			configYaml, _ := yaml.Marshal(config)
			expectedYaml, _ := yaml.Marshal(testCase.expectedConfig)

			return errors.Errorf("TestCase '%v': Got\n %s\n\n, but expected\n\n %s\n", testCase.name, configYaml, expectedYaml)
		}
	}

	return nil
}
