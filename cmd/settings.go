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

func (s *EnvSettings) New() *EnvSettings {
	envSetting := EnvSettings{
		Namespace:   os.Getenv("HELM_NAMESPACE"),
		KubeContext: os.Getenv("HELM_KUBECONTEXT"),
		KubeConfig:  os.Getenv("KUBECONFIG"),
	}

	return &envSetting
}

func (s *EnvSettings) AddFlags(_ *pflag.FlagSet) {
}
