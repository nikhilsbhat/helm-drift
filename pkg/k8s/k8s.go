package k8s

import (
	"fmt"
	"strings"

	"github.com/nikhilsbhat/helm-drift/pkg/errors"
	"github.com/thoas/go-funk"
	"sigs.k8s.io/yaml"
)

type (
	Resource map[string]interface{}
)

// ResourceInterface implements methods to get resource name and kind.
type ResourceInterface interface {
	GetName(dataMap string) (string, error)
	GetKind(dataMap string) (string, error)
	GetNameSpace(name, kind, dataMap string) (string, error)
	IsHelmHook(dataMap string, hookKinds []string) (bool, error)
}

// GetName gets the name form the kubernetes resource.
func (resource *Resource) GetName(dataMap string) (string, error) {
	var kindYaml map[string]interface{}
	if err := yaml.Unmarshal([]byte(dataMap), &kindYaml); err != nil {
		return "", err
	}

	if len(kindYaml) != 0 {
		value, failedManifest := kindYaml["metadata"].(map[string]interface{})["name"].(string)
		if !failedManifest {
			return "", &errors.DriftError{Message: "failed to get name from the manifest, 'name' is not type string"}
		}

		return value, nil
	}

	return "", nil
}

// GetKind helps in identifying kind form the kubernetes resource.
func (resource *Resource) GetKind(dataMap string) (string, error) {
	var kindYaml map[string]interface{}
	if err := yaml.Unmarshal([]byte(dataMap), &kindYaml); err != nil {
		return "", err
	}

	if len(kindYaml) != 0 {
		value, failedManifest := kindYaml["kind"].(string)
		if !failedManifest {
			return "", &errors.DriftError{Message: "failed to get kube kind from the manifest, 'kind' is not type string"}
		}

		return value, nil
	}

	return "", nil
}

// GetNameSpace gets the namespace form the kubernetes resource.
func (resource *Resource) GetNameSpace(name, kind, dataMap string) (string, error) {
	var kindYaml map[string]interface{}
	if err := yaml.Unmarshal([]byte(dataMap), &kindYaml); err != nil {
		return "", err
	}

	if len(kindYaml) != 0 {
		value, failedManifest := kindYaml["metadata"].(map[string]interface{})["namespace"].(string)
		if !failedManifest {
			return "", &errors.NotFoundError{Key: "namespace", Manifest: fmt.Sprintf("%s/%s", name, kind)}
		}

		return value, nil
	}

	return "", nil
}

func isNestedKeyNotNil(data map[string]interface{}, key string) bool {
	if len(data) == 0 {
		return false
	}

	keys := splitKey(key, ".", "\\")

	// Traverse the nested structure
	for i, k := range keys { //nolint:varnamelen
		value, ok := data[k]
		if !ok || value == nil {
			return false
		}

		if nestedMap, ok := value.(map[string]interface{}); ok {
			data = nestedMap
		} else if i+1 < len(keys) {
			// Key does not point to a map, so we can't check deeper.
			return false
		} else {
			// Last key is valid and not nil.
			return true
		}
	}

	return false
}

func splitKey(key string, delimiter string, escapedchar string) []string {
	// Split the key using the specified delimiter
	parts := strings.Split(key, delimiter)

	// Merge any escaped delimiters with the previous part
	var result []string

	for i := 0; i < len(parts); i++ { //nolint:varnamelen
		if strings.HasSuffix(parts[i], escapedchar) {
			// Remove the trailing backslash and merge it with the next part
			parts[i] = strings.TrimSuffix(parts[i], escapedchar)
			if i+1 < len(parts) {
				result = append(result, parts[i]+delimiter+parts[i+1])
				i++ // Skip the next part
			} else {
				// If there's no next part, just append the escaped part
				result = append(result, parts[i])
			}
		} else {
			result = append(result, parts[i])
		}
	}

	return result
}

// IsHelmHook gets the namespace form the kubernetes resource.
func (resource *Resource) IsHelmHook(dataMap string, hookKinds []string) (bool, error) {
	var kindYaml map[string]interface{}
	if err := yaml.Unmarshal([]byte(dataMap), &kindYaml); err != nil {
		return false, err
	}

	if !isNestedKeyNotNil(kindYaml, "metadata.annotations.helm\\.sh/hook") || !isNestedKeyNotNil(kindYaml, "metadata.annotations.helm\\.sh/hook-delete-policy") {
		return false, nil
	}

	hookType, failedManifest := kindYaml["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})["helm.sh/hook-delete-policy"].(string)
	if !failedManifest {
		return false, nil
	}

	hookType = strings.TrimSpace(hookType)

	hookTypes := make([]string, 0)

	if len(strings.Split(hookType, ",")) > 1 {
		hookTypes = strings.Split(hookType, ",")
	}

	for _, hkType := range hookTypes {
		if funk.Contains(hookKinds, hkType) {
			return true, nil
		}
	}

	return false, nil
}

// NewResource returns aa new instance of ResourceInterface.
func NewResource() ResourceInterface {
	return &Resource{}
}
