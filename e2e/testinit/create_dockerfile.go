package testinit

import (
	"errors"
	"os"

	"github.com/devspace-cloud/devspace/cmd"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/versions/latest"
	"github.com/devspace-cloud/devspace/pkg/devspace/docker"
	"github.com/devspace-cloud/devspace/pkg/util/log"
	"github.com/devspace-cloud/devspace/pkg/util/ptr"
	dockertypes "github.com/docker/docker/api/types"
)

// CreateDockerfile runs init test with "create docker file" option
func CreateDockerfile(namespace string, pwd string) error {
	log.Info("Create Dockerfile Test")

	utils.ResetConfigs()

	dirPath, dirName, err := utils.CreateTempDir()
	if err != nil {
		return err
	}

	defer utils.DeleteTempDir(dirPath)

	err = utils.ChangeWorkingDir(dirPath)
	if err != nil {
		return err
	}

	// Copy the testdata into the temp dir
	utils.Copy(pwd+"/testinit/testdata/main.go", dirPath+"/main.go")

	port := 8080
	testCase := &initTestCase{
		name: "Create Dockerfile",
		fakeDockerClient: &docker.FakeClient{
			AuthConfig: &dockertypes.AuthConfig{
				Username: "user",
				Password: "pass",
			},
		},
		answers: []string{cmd.CreateDockerfileOption, "go", "Use hub.docker.com => you are logged in as user", "user/" + dirName, "8080"},
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
						ExcludePaths: []string{"Dockerfile", "devspace.yaml"},
					},
				},
			},
		},
	}

	err = initializeTest(*testCase)
	if err != nil {
		return err
	}

	// Check if Dockerfile has not been created
	if _, err := os.Stat(dirPath + "/Dockerfile"); os.IsNotExist(err) {
		return errors.New("Dockerfile was not created")
	}

	return nil
}
