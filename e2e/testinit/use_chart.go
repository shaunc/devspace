package testinit

import (
	"github.com/devspace-cloud/devspace/cmd"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/versions/latest"
	"github.com/devspace-cloud/devspace/pkg/devspace/docker"
	dockertypes "github.com/docker/docker/api/types"
)

// UseChart runs init test with "use helm chart" option
func UseChart(namespace string, pwd string) error {
	utils.ResetConfigs()

	dirPath, dirName, err := utils.CreateTempDir()
	if err != nil {
		return err
	}

	defer utils.DeleteTempDir(dirPath)

	utils.Copy(pwd+"/testinit/testdata", dirPath)

	err = utils.ChangeWorkingDir(dirPath)
	if err != nil {
		return err
	}

	testCase := &initTestCase{
		name: "Enter helm chart",
		fakeDockerClient: &docker.FakeClient{
			AuthConfig: &dockertypes.AuthConfig{
				Username: "user",
				Password: "pass",
			},
		},
		answers: []string{cmd.EnterHelmChartOption, "./chart"},
		expectedConfig: &latest.Config{
			Version: latest.Version,
			Deployments: []*latest.DeploymentConfig{
				&latest.DeploymentConfig{
					Name: dirName,
					Helm: &latest.HelmConfig{
						Chart: &latest.ChartConfig{
							Name: "./chart",
						},
					},
				},
			},
			Dev:    &latest.DevConfig{},
			Images: latest.NewRaw().Images,
		},
	}

	err = initializeTest(*testCase)
	if err != nil {
		return err
	}

	return nil
}
