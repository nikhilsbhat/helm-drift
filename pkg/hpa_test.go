package pkg_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nikhilsbhat/helm-drift/pkg"
	"github.com/stretchr/testify/assert"
)

func TestDrift_IsManagedByHPA(t *testing.T) {
	t.Run("", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		assert.NoError(t, err)

		drift := pkg.Drift{}
		drift.SetKubeConfig(filepath.Join(homeDir, ".kube/config"))

		output, err := drift.IsManagedByHPA("sample", "Deployment", "sample")
		assert.NoError(t, err)
		assert.False(t, output)
	})
}
