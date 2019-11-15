package tests

import (
	"path/filepath"

	"github.com/devspace-cloud/devspace/cmd"
	"github.com/devspace-cloud/devspace/cmd/flags"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/configutil"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/generated"
	"github.com/devspace-cloud/devspace/pkg/devspace/kubectl"
	"github.com/devspace-cloud/devspace/pkg/devspace/services"
	"github.com/devspace-cloud/devspace/pkg/util/log"
	"github.com/pkg/errors"
)

// RunKaniko runs the test for the kaniko example
func RunKaniko(namespace string) error {
	var deployConfig = &cmd.DeployCmd{
		GlobalFlags: &flags.GlobalFlags{
			Namespace: namespace,
			NoWarn:    true,
		},
		ForceBuild:  true,
		ForceDeploy: true,
		SkipPush:    true,
	}

	// Create kubectl client
	client, err := kubectl.NewClientFromContext(deployConfig.KubeContext, deployConfig.Namespace, deployConfig.SwitchContext)
	if err != nil {
		return errors.Errorf("Unable to create new kubectl client: %v", err)
	}

	defer func() {
		// At last, we delete the current namespace
		utils.DeleteNamespaceAndWait(client, deployConfig.Namespace)
	}()

	var wd string
	wd, err = filepath.Abs("../examples/kaniko/")
	if err != nil {
		return err
	}

	utils.ChangeWorkingDir(wd)
	if err != nil {
		return err
	}

	err = deployConfig.Run(nil, nil)
	if err != nil {
		return err
	}

	// Checking if pods are running correctly
	utils.AnalyzePods(client, namespace)

	// Load generated config
	generatedConfig, err := generated.LoadConfig(deployConfig.Profile)
	if err != nil {
		return errors.Errorf("Error loading generated.yaml: %v", err)
	}

	// Add current kube context to context
	configOptions := deployConfig.ToConfigOptions()
	config, err := configutil.GetConfig(configOptions)
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
