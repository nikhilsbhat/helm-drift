package k8s

import (
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/nikhilsbhat/helm-drift/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

type (
	Resource map[string]interface{}
)

// ResourceInterface implements methods to get resource name and kind.
type ResourceInterface interface {
	GetName(dataMap string, log *logrus.Logger) (string, error)
	GetKind(dataMap string, log *logrus.Logger) (string, error)
	GetNameSpace(name, kind, dataMap string, log *logrus.Logger) (string, error)
	IsHelmHook(dataMap string, hookKinds []string) (bool, error)
}

// GetName gets the name form the kubernetes resource.
func (resource *Resource) GetName(dataMap string, log *logrus.Logger) (string, error) {
	if err := yaml.Unmarshal([]byte(dataMap), resource); err != nil {
		return "", err
	}

	kindYaml := *resource

	metadata, metadataExists := kindYaml["metadata"].(map[string]interface{})
	if !metadataExists {
		log.Debug("failed to get 'metadata' from the manifest")

		return "", nil
	}

	value, failedManifest := metadata["name"].(string)
	if !failedManifest {
		return "", &errors.DriftError{Message: "failed to get name from the manifest, 'name' is not type string"}
	}

	return value, nil
}

// GetKind helps in identifying kind form the kubernetes resource.
func (resource *Resource) GetKind(dataMap string, _ *logrus.Logger) (string, error) {
	if err := yaml.Unmarshal([]byte(dataMap), resource); err != nil {
		return "", err
	}

	kindYaml := *resource

	value, failedManifest := kindYaml["kind"].(string)
	if !failedManifest {
		return "", &errors.DriftError{Message: "failed to get kube kind from the manifest, 'kind' is not type string"}
	}

	return value, nil
}

// GetNameSpace gets the namespace form the kubernetes resource.
func (resource *Resource) GetNameSpace(name, kind, dataMap string, log *logrus.Logger) (string, error) {
	if err := yaml.Unmarshal([]byte(dataMap), resource); err != nil {
		return "", err
	}

	kindYaml := *resource

	metadata, metadataExists := kindYaml["metadata"].(map[string]interface{})
	if !metadataExists {
		log.Debug("failed to get 'metadata' from the manifest")

		return "", nil
	}

	value, failedManifest := metadata["namespace"].(string)
	if !failedManifest {
		return "", &errors.NotFoundError{Key: "namespace", Manifest: fmt.Sprintf("%s/%s", name, kind)}
	}

	return value, nil
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
		} else {
			// Check if this is the last key and it is not nil
			// Key does not point to a map, so we can't check deeper.
			return i == len(keys)-1
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
	if err := yaml.Unmarshal([]byte(dataMap), resource); err != nil {
		return false, err
	}

	kindYaml := *resource

	if !isNestedKeyNotNil(kindYaml, "metadata.annotations.helm\\.sh/hook") || !isNestedKeyNotNil(kindYaml, "metadata.annotations.helm\\.sh/hook-delete-policy") {
		return false, nil
	}

	annotations, annotationsExists := kindYaml["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
	if !annotationsExists {
		return false, nil
	}

	hookType, deleteHookPolicyExists := annotations["helm.sh/hook-delete-policy"].(string)
	if !deleteHookPolicyExists {
		return false, nil
	}

	hookType = strings.TrimSpace(hookType)

	hookTypes := strings.Split(hookType, ",")

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
