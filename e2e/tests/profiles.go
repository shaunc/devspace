package tests

import (
	"github.com/devspace-cloud/devspace/cmd"
	"github.com/devspace-cloud/devspace/cmd/flags"
	"github.com/devspace-cloud/devspace/cmd/use"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/configutil"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/generated"
	"github.com/devspace-cloud/devspace/pkg/devspace/kubectl"
	"github.com/devspace-cloud/devspace/pkg/devspace/services"
	"github.com/devspace-cloud/devspace/pkg/util/log"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RunProfiles runs the test for the kustomize example
func RunProfiles(namespace string, pwd string) error {
	log.Info("Run Profiles")

	utils.ResetConfigs()

	var deployConfig = &cmd.DeployCmd{
		GlobalFlags: &flags.GlobalFlags{
			Namespace: namespace,
			NoWarn:    true,
		},
		ForceBuild:  true,
		ForceDeploy: true,
		SkipPush:    true,
	}

	err := utils.ChangeWorkingDir(pwd + "/../examples/profiles")
	if err != nil {
		return err
	}

	// Create kubectl client
	var client kubectl.Client
	client, err = kubectl.NewClientFromContext(deployConfig.KubeContext, deployConfig.Namespace, deployConfig.SwitchContext)
	if err != nil {
		return errors.Errorf("Unable to create new kubectl client: %v", err)
	}

	// At last, we delete the current namespace
	defer utils.DeleteNamespaceAndWait(client, deployConfig.Namespace)

	err = runProfile(deployConfig, "dev-service2-only", client, namespace, []string{"service-2"}, false)
	if err != nil {
		return err
	}

	err = runProfile(deployConfig, "", client, namespace, []string{"service-1", "service-2"}, true)
	if err != nil {
		return err
	}

	return nil
}

func runProfile(deployConfig *cmd.DeployCmd, profile string, client kubectl.Client, namespace string, expectedPodLabels []string, reset bool) error {
	utils.ResetConfigs()

	var profileConfig = &use.ProfileCmd{
		Reset: reset,
	}

	if profile == "" {
		err := profileConfig.RunUseProfile(nil, nil)
		if err != nil {
			return err
		}
	} else {
		err := profileConfig.RunUseProfile(nil, []string{profile})
		if err != nil {
			return err
		}
	}

	err := profileConfig.RunUseProfile(nil, []string{profile})
	if err != nil {
		return err
	}

	err = deployConfig.Run(nil, nil)
	if err != nil {
		return err
	}

	// Checking if pods are running correctly
	utils.AnalyzePods(client, namespace)

	pods, errp := client.KubeClient().CoreV1().Pods(namespace).List(v1.ListOptions{})
	if errp != nil {
		return err
	}

	var rp []string
	for _, x := range expectedPodLabels {
		for _, y := range pods.Items {
			if x == y.ObjectMeta.Labels["app.kubernetes.io/component"] {
				rp = append(rp, x)
			}
		}
	}

	if !utils.Equal(rp, expectedPodLabels) {
		return errors.New("The expected pods are not running")
	}

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
