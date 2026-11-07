package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestKubeConfigLoading(t *testing.T) {
	dir := t.TempDir()
	first := clientcmdapi.NewConfig()
	first.CurrentContext = "gpu"
	first.Clusters["gpu"] = &clientcmdapi.Cluster{Server: "https://gpu.example.test"}
	first.AuthInfos["gpu"] = &clientcmdapi.AuthInfo{Token: "gpu-token"}
	first.Contexts["dev"] = &clientcmdapi.Context{Cluster: "dev", AuthInfo: "dev"}
	first.Clusters["dev"] = &clientcmdapi.Cluster{Server: "https://dev.example.test"}
	first.AuthInfos["dev"] = &clientcmdapi.AuthInfo{Token: "dev-token"}
	firstPath := filepath.Join(dir, "first")
	require.NoError(t, clientcmd.WriteToFile(*first, firstPath))

	// The selected context is in the second file, but its cluster and user are in the first.
	second := clientcmdapi.NewConfig()
	second.Contexts["gpu"] = &clientcmdapi.Context{Cluster: "gpu", AuthInfo: "gpu"}
	secondPath := filepath.Join(dir, "second")
	require.NoError(t, clientcmd.WriteToFile(*second, secondPath))

	explicit := clientcmdapi.NewConfig()
	explicit.CurrentContext = "explicit"
	explicit.Contexts["explicit"] = &clientcmdapi.Context{Cluster: "explicit", AuthInfo: "explicit"}
	explicit.Clusters["explicit"] = &clientcmdapi.Cluster{Server: "https://explicit.example.test"}
	explicit.AuthInfos["explicit"] = &clientcmdapi.AuthInfo{Token: "explicit-token"}
	explicitPath := filepath.Join(dir, "explicit")
	require.NoError(t, clientcmd.WriteToFile(*explicit, explicitPath))

	t.Setenv("KUBECONFIG", strings.Join([]string{firstPath, secondPath, ""}, string(os.PathListSeparator)))
	t.Setenv("HELM_KUBECONTEXT", "")
	t.Setenv("HELM_KUBETOKEN", "")
	t.Setenv("HELM_KUBEAPISERVER", "")

	for _, test := range []struct {
		name    string
		context string
		path    string
		target  string
	}{
		{name: "merged current context", target: "gpu"},
		{name: "explicit context with merged credentials", context: "gpu", target: "gpu"},
		{name: "context override", context: "dev", target: "dev"},
		{name: "explicit file overrides environment", path: explicitPath, target: "explicit"},
		{name: "missing context", context: "missing"},
	} {
		t.Run(test.name, func(t *testing.T) {
			drift := &Drift{}
			drift.SetKubeConfig(test.path)
			drift.SetKubeContext(test.context)

			for _, loader := range []struct {
				name string
				load func() (*rest.Config, error)
			}{
				{name: "client-go", load: func() (*rest.Config, error) {
					return buildConfigWithContextFromFlags(drift.kubeContext, drift.kubeConfig)
				}},
				{name: "helm", load: drift.helmSettings().RESTClientGetter().ToRESTConfig},
			} {
				t.Run(loader.name, func(t *testing.T) {
					config, err := loader.load()
					if test.target == "" {
						require.ErrorContains(t, err, "missing")

						return
					}

					require.NoError(t, err)
					assert.Equal(t, "https://"+test.target+".example.test", config.Host)
					assert.Equal(t, test.target+"-token", config.BearerToken)
				})
			}
		})
	}
}
