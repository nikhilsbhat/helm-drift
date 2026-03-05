package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nikhilsbhat/helm-drift/pkg/errors"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

type EnvSettings struct {
	KubeConfig  string
	KubeContext string
	Namespace   string
}

func (s *EnvSettings) New() (*EnvSettings, error) {
	envSetting := EnvSettings{
		Namespace:   os.Getenv("HELM_NAMESPACE"),
		KubeContext: os.Getenv("HELM_KUBECONTEXT"),
		KubeConfig:  os.Getenv("KUBECONFIG"),
	}

	kubeConfig, err := findKubeConfigForContext(envSetting.KubeContext)
	if err != nil {
		return nil, err
	}

	envSetting.KubeConfig = kubeConfig

	return &envSetting, nil
}

func (s *EnvSettings) AddFlags(_ *pflag.FlagSet) {
}

func findKubeConfigForContext(context string) (string, error) {
	KubeConfigFromEnv := os.Getenv("KUBECONFIG")
	if KubeConfigFromEnv == "" {
		return "", &errors.DriftError{Message: "'KUBECONFIG' env variable is not set"}
	}

	paths := strings.Split(KubeConfigFromEnv, ":")

	for _, p := range paths {
		expanded, err := expandHome(p)
		if err != nil {
			continue
		}

		cfg, err := clientcmd.LoadFromFile(expanded)
		if err != nil {
			continue
		}

		if _, ok := cfg.Contexts[context]; ok {
			return expanded, nil
		}
	}

	return "", &errors.DriftError{Message: fmt.Sprintf("context %q not found in any kubeconfig file", context)}
}

func expandHome(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, path[1:]), nil
	}

	return path, nil
}
