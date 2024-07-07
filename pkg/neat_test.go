//nolint:testpackage
package pkg

import (
	"path/filepath"
	"testing"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/util/homedir"
)

func TestDrift_NeatNew(t *testing.T) {
	t.Run("", func(t *testing.T) {
		drift := Drift{}
		drift.SetLogger("debug")
		drift.SetNamespace("sample")
		drift.SetKubeConfig(filepath.Join(homedir.HomeDir(), ".kube", "config"))
		dvn := deviation.Deviation{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
			Resource:   "sample",
		}

		out, err := drift.neat(dvn)
		assert.NoError(t, err)

		assert.Nil(t, out)
	})
}
