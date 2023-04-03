package command

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
)

func (cmd *command) SetKubeCmd(namespace string, args ...string) {
	cmd.baseCmd.Env = cmd.prepareKubeEnvironments()
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, "diff")
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, args...)
	cmd.baseCmd.Args = append(cmd.baseCmd.Args, cmd.getNamespace(namespace))

	if len(setContext()) != 0 {
		cmd.baseCmd.Args = append(cmd.baseCmd.Args, setContext())
	}

	cmd.log.Debugf("running command '%s' to find diff", cmd.baseCmd.String())
}

func (cmd *command) prepareKubeEnvironments() []string {
	config := os.Getenv(KubeConfig)

	os.Environ()
	var envs []string

	if len(config) != 0 {
		envs = append(envs, constructEnv(KubeConfig, config))
	} else {
		envs = append(envs, constructEnv(KubeConfig, filepath.Join(homedir.HomeDir(), ".kube", "config")))
	}

	envs = append(envs, os.Environ()...)

	return envs
}

func (cmd *command) getNamespace(nameSpace string) string {
	return fmt.Sprintf("-n=%s", nameSpace)
}

func constructEnv(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

func setContext() string {
	kubeContext := os.Getenv(HelmContext)
	if len(kubeContext) != 0 {
		return fmt.Sprintf("--context=%s", kubeContext)
	}

	return kubeContext
}
