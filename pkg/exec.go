package pkg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

const (
	helmContext   = "HELM_KUBECONTEXT"
	helmNamespace = "HELM_NAMESPACE"
	kubeContext   = "HELM_KUBECONTEXT"
	kubeNamespace = "HELM_NAMESPACE"
	kubeConfig    = "KUBECONFIG"
)

func (drift *Drift) kubeCmd(args ...string) *exec.Cmd {
	cmd := exec.CommandContext(context.Background(), "kubectl", args...)
	cmd.Env = drift.getKubeEnvironments()
	drift.log.Debugf("running command '%s' to find diff", cmd.String())

	return cmd
}

func (drift *Drift) getKubeEnvironments() []string {
	contexts := os.Getenv(helmContext)
	namespace := os.Getenv(helmNamespace)
	config := os.Getenv(kubeConfig)

	var envs []string
	if len(contexts) != 0 {
		envs = append(envs, constructEnv(kubeContext, contexts))
	}
	if len(namespace) != 0 {
		envs = append(envs, constructEnv(kubeNamespace, namespace))
	}
	if len(config) != 0 {
		envs = append(envs, constructEnv(kubeConfig, config))
	}

	return envs
}

func constructEnv(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}
