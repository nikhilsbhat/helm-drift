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
	kubeConfig    = "KUBECONFIG"
)

func (drift *Drift) kubeCmd(args ...string) *exec.Cmd {
	cmd := exec.CommandContext(context.Background(), "kubectl")
	cmd.Env = drift.getKubeEnvironments()
	cmd.Args = append(cmd.Args, "diff")
	cmd.Args = append(cmd.Args, args...)
	cmd.Args = append(cmd.Args, drift.getNamespace())

	if len(setContext()) != 0 {
		cmd.Args = append(cmd.Args, setContext())
	}

	drift.log.Debugf("running command '%s' to find diff", cmd.String())

	return cmd
}

func (drift *Drift) getKubeEnvironments() []string {
	config := os.Getenv(kubeConfig)

	os.Environ()
	var envs []string

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

func setContext() string {
	kubeContext := os.Getenv(helmContext)
	if len(kubeContext) != 0 {
		return fmt.Sprintf("--kube-context=%s", kubeContext)
	}

	return kubeContext
}

func (drift *Drift) getNamespace() string {
	return fmt.Sprintf("-n=%s", drift.namespace)
}
