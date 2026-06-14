package pkg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDriftSettersAndOutputFormats(t *testing.T) {
	drift := Drift{}

	drift.SetRelease("release")
	drift.SetChart("chart")
	drift.SetNamespace("")
	drift.SetKubeConfig("")
	drift.SetKubeContext("context")

	assert.Equal(t, "default", drift.namespace)
	assert.Equal(t, "context", drift.kubeContext)
	assert.Contains(t, drift.kubeConfig, ".kube/config")

	drift.OutputFormat = "json"
	drift.SetOutputFormats()
	assert.True(t, drift.json)

	drift = Drift{OutputFormat: "yaml"}
	drift.SetOutputFormats()
	assert.True(t, drift.yaml)

	drift = Drift{OutputFormat: "table"}
	drift.SetOutputFormats()
	assert.True(t, drift.table)
}

func TestSetReleasesToSkips(t *testing.T) {
	drift := Drift{SkipReleases: []string{"release=namespace"}}

	require.NoError(t, drift.SetReleasesToSkips())
	assert.Equal(t, []resourcesInfo{{name: "release", namespace: "namespace"}}, drift.releasesToSkip)

	drift = Drift{SkipReleases: []string{"invalid"}}
	assert.Error(t, drift.SetReleasesToSkips())
}

func TestSetExternalDiff(t *testing.T) {
	drift := Drift{CustomDiff: "dyff between"}

	require.NoError(t, drift.setExternalDiff())
	assert.Equal(t, "dyff between", os.Getenv("KUBECTL_EXTERNAL_DIFF"))
}

func TestAddNewLineAndCaption(t *testing.T) {
	drift := Drift{}
	drift.SetNamespace("sample")
	drift.SetRelease("release")

	assert.Equal(t, "message\n", addNewLine("message"))
	assert.Equal(t, "Namespace: 'sample'\nRelease: 'release'", drift.getCaption())
}

func TestIsAll(t *testing.T) {
	assert.True(t, (&Drift{}).isAll())
	assert.True(t, (&Drift{namespace: "default"}).isAll())
	assert.False(t, (&Drift{namespace: "default", IsDefaultNamespace: true}).isAll())
	assert.False(t, (&Drift{namespace: "sample"}).isAll())
}
