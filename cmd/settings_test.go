package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nikhilsbhat/helm-drift/pkg"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvSettingsNew(t *testing.T) {
	for _, context := range []string{"", "wanted"} {
		t.Run("context="+context, func(t *testing.T) {
			paths := strings.Join([]string{
				filepath.Join(t.TempDir(), "first"),
				filepath.Join(t.TempDir(), "second"), "",
			}, string(os.PathListSeparator))
			t.Setenv("KUBECONFIG", paths)
			t.Setenv("HELM_KUBECONTEXT", context)
			t.Setenv("HELM_NAMESPACE", "sample")

			settings, err := new(EnvSettings).New()
			require.NoError(t, err)
			assert.Empty(t, settings.KubeConfig)
			assert.Equal(t, context, settings.KubeContext)
			assert.Equal(t, "sample", settings.Namespace)
			assert.Equal(t, paths, os.Getenv("KUBECONFIG"))
		})
	}
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
