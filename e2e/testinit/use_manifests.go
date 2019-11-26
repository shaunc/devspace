package testinit

import (
	"github.com/devspace-cloud/devspace/cmd"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/versions/latest"
	"github.com/devspace-cloud/devspace/pkg/devspace/docker"
	dockertypes "github.com/docker/docker/api/types"
)

// UseManifests runs init test with "use kubernetes manifests" option
func UseManifests(namespace string, pwd string) error {
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
		name: "Enter kubernetes manifests",
		fakeDockerClient: &docker.FakeClient{
			AuthConfig: &dockertypes.AuthConfig{
				Username: "user",
				Password: "pass",
			},
		},
		answers: []string{cmd.EnterManifestsOption, "kube/**"},
		expectedConfig: &latest.Config{
			Version: latest.Version,
			Deployments: []*latest.DeploymentConfig{
				&latest.DeploymentConfig{
					Name: dirName,
					Kubectl: &latest.KubectlConfig{
						Manifests: []string{
							"kube/**",
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
