package command

//go:generate mockgen -destination ../mocks/command/run.go -package mockCommand -source ./run.go
import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
)

// RunKubeDiffCmd runs the kubectl command with all predefined arguments.
func (cmd *command) RunKubeDiffCmd(deviation *deviation.Deviation) (*deviation.Deviation, error) {
	cmd.log.Debugf("envionment variables that would be used: %v", cmd.baseCmd.Environ())

	out, err := cmd.baseCmd.CombinedOutput()
	if err != nil {
		var exerr *exec.ExitError
		if errors.As(err, &exerr) {
			switch exerr.ExitCode() {
			case 1:
				deviation.HasDrift = true
				deviation.Deviations = string(out)
				cmd.log.Debugf("found diffs for '%s' with name '%s'", deviation.Kind, deviation.Kind)
			default:
				return deviation, fmt.Errorf("running kubectl diff errored with exit code: %w ,with message: %s", err, string(out))
			}
		}
	} else {
		cmd.log.Debugf("no diffs found for '%s' with name '%s'", deviation.Kind, deviation.Kind)
	}

	return deviation, nil
}

func (cmd *command) RunKubeCmd(deviation *deviation.Deviation) ([]byte, error) {
	cmd.log.Debugf("envionment variables that would be used: %v", cmd.baseCmd.Environ())

	out, err := cmd.baseCmd.CombinedOutput()
	if err != nil {
		cmd.log.Errorf("fetching manifests for '%s' with name '%s' errored with: '%s'", deviation.Kind, deviation.Kind, string(out))

		return nil, err
	}

	return out, nil
}
