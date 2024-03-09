package command

//go:generate mockgen -destination ../mocks/command/set.go -package mockCommand -source ./set.go
import (
	"fmt"
	"os"
)

// SetKubeDiffCmd sets the kubectl diff command with all predefined arguments.
func (cmd *command) SetKubeDiffCmd(kubeConfig string, kubeContext string, namespace string, args ...string) {
	cmd.baseCmd.Env = os.Environ()
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, "diff")
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, args...)
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, cmd.getNamespace(namespace))
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, getContext(kubeConfig, kubeContext)...)

	cmd.log.Debugf("running command '%s' to find diff", cmd.baseCmd.String())
}

func (cmd *command) SetKubeGetCmd(kubeConfig string, kubeContext string, namespace string, args ...string) {
	cmd.baseCmd.Env = os.Environ()
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, "get")
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, args...)
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, cmd.getNamespace(namespace))
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, getContext(kubeConfig, kubeContext)...)

	cmd.log.Debugf("running command '%s' to find get", cmd.baseCmd.String())
}

func (cmd *command) getNamespace(nameSpace string) string {
	return fmt.Sprintf("-n=%s", nameSpace)
}

func getContext(kubeConfig string, kubeContext string) []string {
	cmds := make([]string, 0)
	if len(kubeContext) != 0 {
		cmds = append(cmds, fmt.Sprintf("--context=%s", kubeContext))
	}

	if len(kubeConfig) != 0 {
		cmds = append(cmds, fmt.Sprintf("--kubeconfig=%s", kubeConfig))
	}

	return cmds
}
