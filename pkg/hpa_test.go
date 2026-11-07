package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDrift_IsManagedByHPA(t *testing.T) {
	t.Run("uses cached hpa targets", func(t *testing.T) {
		drift := Drift{
			hpaCache: map[string]map[string]struct{}{
				"sample": {
					hpaTargetKey("sample", "Deployment"): {},
				},
			},
		}
		drift.SetLogger("error")

		output, err := drift.IsManagedByHPA("sample", "Deployment", "sample")
		require.NoError(t, err)
		assert.True(t, output)

		output, err = drift.IsManagedByHPA("sample", "Service", "sample")
		require.NoError(t, err)
		assert.False(t, output)
	})
}

func TestDrift_HasOnlyChangesScaledByHpa(t *testing.T) {
	tests := []struct {
		name       string
		diffOutput string
		expected   bool
	}{
		{
			name: "replica changes only",
			diffOutput: `--- before
+++ after
-  replicas: 2
+  replicas: 3`,
			expected: true,
		},
		{
			name: "non hpa change",
			diffOutput: `--- before
+++ after
-  image: nginx:1.24
+  image: nginx:1.25`,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("KUBECTL_EXTERNAL_DIFF", "")

			drift := Drift{}
			drift.SetLogger("error")

			output, err := drift.HasOnlyChangesScaledByHpa(test.diffOutput)
			require.NoError(t, err)
			assert.Equal(t, test.expected, output)
		})
	}
}
