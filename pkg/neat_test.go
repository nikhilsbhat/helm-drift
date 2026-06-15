//nolint:testpackage
package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCleanResource(t *testing.T) {
	t.Run("drops cluster managed fields", func(t *testing.T) {
		resource := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "ServiceAccount",
				"metadata": map[string]any{
					"name":              "sample",
					"namespace":         "sample",
					"managedFields":     []any{},
					"uid":               "uid",
					"resourceVersion":   "1",
					"generation":        int64(1),
					"creationTimestamp": "2026-06-14T00:00:00Z",
				},
				"status": map[string]any{},
			},
		}

		cleanedResource := cleanResource(resource)

		assert.NotContains(t, cleanedResource.Object, "status")
		assert.NotContains(t, cleanedResource.Object["metadata"], "managedFields")
		assert.NotContains(t, cleanedResource.Object["metadata"], "uid")
		assert.NotContains(t, cleanedResource.Object["metadata"], "resourceVersion")
		assert.NotContains(t, cleanedResource.Object["metadata"], "generation")
		assert.NotContains(t, cleanedResource.Object["metadata"], "creationTimestamp")
	})
}

func TestDrift_DropStandardHelmLabels(t *testing.T) {
	t.Run("drops helm labels and annotations", func(t *testing.T) {
		drift := Drift{}
		drift.SetLogger("debug")

		resource := &unstructured.Unstructured{
			Object: map[string]any{
				"metadata": map[string]any{
					"labels": map[string]any{
						"app.kubernetes.io/name": "sample",
						"custom":                 "keep",
					},
					"annotations": map[string]any{
						"meta.helm.sh/release-name": "sample",
						"custom":                    "keep",
					},
				},
			},
		}

		cleanedResource := drift.dropStandardHelmLabels(resource)
		labels, _, _ := unstructured.NestedStringMap(cleanedResource.Object, "metadata", "labels")
		annotations, _, _ := unstructured.NestedStringMap(cleanedResource.Object, "metadata", "annotations")

		assert.NotContains(t, labels, "app.kubernetes.io/name")
		assert.Equal(t, "keep", labels["custom"])
		assert.NotContains(t, annotations, "meta.helm.sh/release-name")
		assert.Equal(t, "keep", annotations["custom"])
	})
}
