package logs

import (
	"github.com/devspace-cloud/devspace/cmd"
	"github.com/devspace-cloud/devspace/cmd/flags"
	"github.com/devspace-cloud/devspace/e2e/utils"
	"github.com/pkg/errors"
	"strings"
)

func runDefault(f *customFactory) error {
	lc := &cmd.LogsCmd{
		GlobalFlags: &flags.GlobalFlags{
			Namespace: f.namespace,
			NoWarn:    true,
			Silent:    true,
		},
		LastAmountOfLines: 1,
	}

	done := utils.Capture()

	err := lc.RunLogs(nil, nil)
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
