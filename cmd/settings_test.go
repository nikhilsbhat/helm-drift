package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nikhilsbhat/helm-drift/pkg"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	expanded, err := expandHome("~/config")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "config"), expanded)

	plain, err := expandHome("/tmp/config")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/config", plain)
}

func TestFindKubeConfigForContext(t *testing.T) {
	kubeConfig := filepath.Join(t.TempDir(), "config")
	require.NoError(t, os.WriteFile(kubeConfig, []byte(`
apiVersion: v1
kind: Config
clusters:
- name: local
  cluster:
    server: https://127.0.0.1
contexts:
- name: wanted
  context:
    cluster: local
    user: local
current-context: wanted
users:
- name: local
  user: {}
`), 0o600))
	t.Setenv("KUBECONFIG", kubeConfig)

	found, err := findKubeConfigForContext("wanted")
	require.NoError(t, err)
	assert.Equal(t, kubeConfig, found)

	found, err = findKubeConfigForContext("")
	require.NoError(t, err)
	assert.Equal(t, kubeConfig, found)

	_, err = findKubeConfigForContext("missing")
	require.Error(t, err)
}

func TestFindKubeConfigForContextWithoutKubeConfig(t *testing.T) {
	t.Setenv("KUBECONFIG", "")

	found, err := findKubeConfigForContext("")

	require.NoError(t, err)
	assert.Empty(t, found)
}

func TestEnvSettingsNew(t *testing.T) {
	kubeConfig := filepath.Join(t.TempDir(), "config")
	require.NoError(t, os.WriteFile(kubeConfig, []byte(`
apiVersion: v1
kind: Config
clusters:
- name: local
  cluster:
    server: https://127.0.0.1
contexts:
- name: wanted
  context:
    cluster: local
    user: local
current-context: wanted
users:
- name: local
  user: {}
`), 0o600))
	t.Setenv("KUBECONFIG", kubeConfig)
	t.Setenv("HELM_KUBECONTEXT", "wanted")
	t.Setenv("HELM_NAMESPACE", "sample")

	settings, err := new(EnvSettings).New()
	require.NoError(t, err)

	assert.Equal(t, kubeConfig, settings.KubeConfig)
	assert.Equal(t, "wanted", settings.KubeContext)
	assert.Equal(t, "sample", settings.Namespace)
}

func TestValidateAndSetArgs(t *testing.T) {
	t.Cleanup(func() { drifts = pkg.Drift{} })

	drifts = pkg.Drift{}

	err := validateAndSetArgs(&cobra.Command{}, []string{"release"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "[RELEASE] or [CHART] cannot be empty")

	drifts = pkg.Drift{}
	err = validateAndSetArgs(&cobra.Command{}, []string{"release", "chart"})
	require.NoError(t, err)

	drifts = pkg.Drift{FromRelease: true}
	err = validateAndSetArgs(&cobra.Command{}, []string{"release"})
	require.NoError(t, err)
}

func TestUsageTemplate(t *testing.T) {
	usage := getUsageTemplate()

	assert.Contains(t, usage, "Usage:")
	assert.Contains(t, usage, "Available Commands:")
}
