package pkg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueFiles(t *testing.T) {
	values := ValueFiles{}

	require.NoError(t, values.Set("values.yaml,override.yaml"))

	assert.Equal(t, ValueFiles{"values.yaml", "override.yaml"}, values)
	assert.Equal(t, "[values.yaml override.yaml]", values.String())
	assert.Equal(t, "ValueFiles", values.Type())
}

func TestValueFilesValid(t *testing.T) {
	valuesFile := filepath.Join(t.TempDir(), "values.yaml")
	require.NoError(t, os.WriteFile(valuesFile, []byte("key: value\n"), 0o600))

	values := ValueFiles{valuesFile, "-"}

	assert.NoError(t, values.Valid())
}

func TestValueFilesInvalid(t *testing.T) {
	values := ValueFiles{filepath.Join(t.TempDir(), "missing.yaml")}

	assert.Error(t, values.Valid())
}
