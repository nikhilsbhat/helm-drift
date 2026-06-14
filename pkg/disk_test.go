package pkg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderToDiskAndCleanManifests(t *testing.T) {
	tempDir := t.TempDir()
	drift := Drift{TempPath: tempDir}
	drift.SetLogger("error")
	drift.SetRelease("release")

	rendered, err := drift.renderToDisk([]string{deploymentManifest}, "chart", "release", "sample")
	require.NoError(t, err)
	require.Len(t, rendered.Deviations, 1)

	manifestPath := rendered.Deviations[0].ManifestPath
	assert.FileExists(t, manifestPath)
	assert.Equal(t, "release", rendered.Release)
	assert.Equal(t, "sample", rendered.Namespace)
	assert.Equal(t, "chart", rendered.Chart)
	assert.Equal(t, "Deployment", rendered.Deviations[0].Kind)

	require.NoError(t, drift.cleanManifests(false))
	assert.NoFileExists(t, manifestPath)
}

func TestCleanManifestsSkipClean(t *testing.T) {
	tempDir := t.TempDir()
	drift := Drift{TempPath: tempDir, SkipClean: true}
	drift.SetLogger("error")
	drift.SetRelease("release")

	releaseDir := filepath.Join(tempDir, "release")
	require.NoError(t, os.MkdirAll(releaseDir, 0o755))

	require.NoError(t, drift.cleanManifests(false))
	assert.DirExists(t, releaseDir)

	require.NoError(t, drift.cleanManifests(true))
	assert.NoDirExists(t, releaseDir)
}
