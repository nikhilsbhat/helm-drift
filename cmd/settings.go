package cmd

import (
	"os"

	"github.com/spf13/pflag"
)

type EnvSettings struct {
	KubeConfig  string
	KubeContext string
	Namespace   string
}

func (s *EnvSettings) New() (*EnvSettings, error) {
	// Leave the explicit path empty so Helm and kubectl merge KUBECONFIG themselves.
	envSetting := EnvSettings{
		Namespace:   os.Getenv("HELM_NAMESPACE"),
		KubeContext: os.Getenv("HELM_KUBECONTEXT"),
	}

	return &envSetting, nil
}

func (s *EnvSettings) AddFlags(_ *pflag.FlagSet) {
}
