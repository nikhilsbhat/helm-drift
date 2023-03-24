package pkg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
)

const (
	helmContext   = "HELM_KUBECONTEXT"
	helmNamespace = "HELM_NAMESPACE"
	kubeContext   = "HELM_KUBECONTEXT"
	kubeNamespace = "HELM_NAMESPACE"
	kubeConfig    = "KUBECONFIG"
)

func (drift *Drift) kubeCmd(args ...string) *exec.Cmd {
	cmd := exec.CommandContext(context.Background(), "kubectl")
	cmd.Args = append(cmd.Args, "diff")
	cmd.Args = append(cmd.Args, args...)
	cmd.Args = append(cmd.Args, setNamespace())
	cmd.Env = drift.getKubeEnvironments()

	drift.log.Debugf("running command '%s' to find diff", cmd.String())

	return cmd
}

func (drift *Drift) getKubeEnvironments() []string {
	contexts := os.Getenv(helmContext)
	namespace := os.Getenv(helmNamespace)
	config := os.Getenv(kubeConfig)

	os.Environ()
	var envs []string
	if len(contexts) != 0 {
		envs = append(envs, constructEnv(kubeContext, contexts))
	}
	if len(namespace) != 0 {
		envs = append(envs, constructEnv(kubeNamespace, namespace))
	}
	if len(config) != 0 {
		envs = append(envs, constructEnv(kubeConfig, config))
	} else {
		envs = append(envs, constructEnv(kubeConfig, filepath.Join(homedir.HomeDir(), ".kube", "config")))
	}

	envs = append(envs, os.Environ()...)

	return envs
}

func constructEnv(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

func setNamespace() string {
	return fmt.Sprintf("-n=%s", os.Getenv(helmNamespace))
}
