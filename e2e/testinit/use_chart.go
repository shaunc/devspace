package testinit

import (
	"github.com/devspace-cloud/devspace/cmd"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/versions/latest"
)

// UseChart runs init test with "use helm chart" option
func UseChart(factory *customFactory, namespace string, pwd string) error {
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
		name:    "Enter helm chart",
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

	err = initializeTest(factory, *testCase)
	if err != nil {
		return err
	}

	return nil
}
