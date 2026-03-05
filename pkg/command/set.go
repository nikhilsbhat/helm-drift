package command

//go:generate mockgen -destination ../mocks/command/set.go -package mockCommand -source ./set.go
import (
	"fmt"
	"os"
)

func (cmd *command) setKubeCmd(action string, kubeConfig string, kubeContext string, namespace string, args ...string) {
	cmd.baseCmd.Env = os.Environ()
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, action)
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, args...)
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, cmd.getNamespace(namespace))
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, getContext(kubeConfig, kubeContext)...)

	cmd.log.Debugf("running command '%s' to execute '%s'", cmd.baseCmd.String(), action)
}

// SetKubeDiffCmd sets the kubectl diff command with all predefined arguments.
func (cmd *command) SetKubeDiffCmd(kubeConfig string, kubeContext string, namespace string, args ...string) {
	cmd.setKubeCmd("diff", kubeConfig, kubeContext, namespace, args...)
}

// SetKubeGetCmd sets the kubectl get command with all predefined arguments.
func (cmd *command) SetKubeGetCmd(kubeConfig string, kubeContext string, namespace string, args ...string) {
	cmd.setKubeCmd("get", kubeConfig, kubeContext, namespace, args...)
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
