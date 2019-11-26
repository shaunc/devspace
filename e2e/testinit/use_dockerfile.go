package testinit

import (
	"github.com/devspace-cloud/devspace/cmd"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/versions/latest"
	"github.com/devspace-cloud/devspace/pkg/devspace/docker"
	"github.com/devspace-cloud/devspace/pkg/util/ptr"
	dockertypes "github.com/docker/docker/api/types"
)

// UseDockerfile runs init test with "use existing dockerfile" option
func UseDockerfile(namespace string, pwd string) error {
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

	// Reset configs after changing working dir
	utils.ResetConfigs()

	port := 8080
	testCase := &initTestCase{
		name: "Enter existing Dockerfile",
		fakeDockerClient: &docker.FakeClient{
			AuthConfig: &dockertypes.AuthConfig{
				Username: "user",
				Password: "pass",
			},
		},
		answers: []string{cmd.EnterDockerfileOption, "./Dockerfile", "Use hub.docker.com => you are logged in as user", "user/" + dirName, "8080"},
		expectedConfig: &latest.Config{
			Version: latest.Version,
			Images: map[string]*latest.ImageConfig{
				"default": &latest.ImageConfig{
					Image: "user/" + dirName,
				},
			},
			Deployments: []*latest.DeploymentConfig{
				&latest.DeploymentConfig{
					Name: dirName,
					Helm: &latest.HelmConfig{
						ComponentChart: ptr.Bool(true),
						Values: map[interface{}]interface{}{
							"containers": []interface{}{
								map[interface{}]interface{}{
									"image": "user/" + dirName,
								},
							},
							"service": map[interface{}]interface{}{
								"ports": []interface{}{
									map[interface{}]interface{}{
										"port": port,
									},
								},
							},
						},
					},
				},
			},
			Dev: &latest.DevConfig{
				Ports: []*latest.PortForwardingConfig{
					&latest.PortForwardingConfig{
						ImageName: "default",
						PortMappings: []*latest.PortMapping{
							&latest.PortMapping{
								LocalPort: &port,
							},
						},
					},
				},
				Open: []*latest.OpenConfig{
					&latest.OpenConfig{
						URL: "http://localhost:8080",
					},
				},
				Sync: []*latest.SyncConfig{
					&latest.SyncConfig{
						ImageName:    "default",
						ExcludePaths: []string{"devspace.yaml"},
					},
				},
			},
		},
	}

	err = initializeTest(*testCase)
	if err != nil {
		return err
	}

	return nil
}
