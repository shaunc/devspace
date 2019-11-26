package testinit

import (
	"reflect"

	"github.com/devspace-cloud/devspace/cmd"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/devspace-cloud/devspace/pkg/devspace/builder/helper"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/configutil"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/versions/latest"
	"github.com/devspace-cloud/devspace/pkg/devspace/docker"
	"github.com/devspace-cloud/devspace/pkg/util/survey"
	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

type initTestCase struct {
	name string

	fakeDockerClient docker.ClientInterface
	answers          []string

	expectedConfig *latest.Config
}

// TestInit runs the e2e tests for the init cmd
func TestInit(ns string, pwd string) {
	err := CreateDockerfile(ns, pwd)
	utils.PrintTestResult("Create Dockerfile", err)

	err = UseExistingDockerfile(ns, pwd)
	utils.PrintTestResult("Use Existing Dockerfile", err)

	// err = UseDockerfile(ns, pwd)
	// utils.PrintTestResult("Use Dockerfile", err)

	// err = UseManifests(ns, pwd)
	// utils.PrintTestResult("Use Kubectl Manifests", err)

	// err = UseChart(ns, pwd)
	// utils.PrintTestResult("Use Helm Chart", err)
}

func initializeTest(testCase initTestCase) error {
	initConfig := cmd.InitCmd{
		Dockerfile:  helper.DefaultDockerfilePath,
		Reconfigure: false,
		Context:     "",
		Provider:    "",
	}
	docker.SetFakeClient(testCase.fakeDockerClient)

	for _, a := range testCase.answers {
		survey.SetNextAnswer(a)
	}

	// runs init cmd
	err := initConfig.Run(nil, nil)
	if err != nil {
		return err
	}

	if testCase.expectedConfig != nil {
		config, err := configutil.GetConfig(nil)
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
