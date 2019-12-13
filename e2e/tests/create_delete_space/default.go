package create_delete_space

import (
	"github.com/devspace-cloud/devspace/cmd"
	"github.com/devspace-cloud/devspace/cmd/create"
	"github.com/devspace-cloud/devspace/cmd/flags"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/pkg/errors"
	"strings"
)

func runDefault(f *customFactory) error {
	cs := &create.SpaceCmd{}

	deployConfig := &cmd.DeployCmd{
		GlobalFlags: &flags.GlobalFlags{
			Namespace: f.namespace,
			NoWarn:    true,
		},
		ForceBuild:  true,
		ForceDeploy: true,
		SkipPush:    true,
	}

	err := deployConfig.Run(f, nil, nil)
	if err != nil {
		return err
	}

	client, err := f.NewKubeClientFromContext("", f.namespace, false)
	if err != nil {
		return errors.Errorf("Unable to create new kubectl client: %v", err)
	}

	// Checking if pods are running correctly
	err = utils.AnalyzePods(client, f.namespace)
	if err != nil {
		return err
	}

	lc := &cmd.LogsCmd{
		GlobalFlags: &flags.GlobalFlags{
			Namespace: f.namespace,
			NoWarn:    true,
			Silent:    true,
		},
		LastAmountOfLines: 1,
	}

	done := utils.Capture()

	err = lc.RunLogs(nil, nil)
	if err != nil {
		return err
	}

	capturedOutput, err := done()
	if err != nil {
		return err
	}

	if strings.Index(capturedOutput, "blabla world") == -1 {
		return errors.Errorf("capturedOutput '%v' is different than output 'blabla world' for the enter cmd", capturedOutput)
	}

	return nil
}
