package k8s

import (
	"fmt"
	"strings"

	"github.com/nikhilsbhat/helm-drift/pkg/errors"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

type (
	Resource map[string]interface{}
)

// ResourceInterface implements methods to get resource name and kind.
type ResourceInterface interface {
	Get(dataMap string, key string, log *logrus.Logger) (string, error)
	GetMetadata(dataMap string, key string, log *logrus.Logger) (string, error)
	IsHelmHook(dataMap string, hookKinds []string) (bool, error)
}

// Get helps in identifying kind form the kubernetes resource.
func (resource *Resource) Get(dataMap string, key string, log *logrus.Logger) (string, error) {
	if err := yaml.Unmarshal([]byte(dataMap), resource); err != nil {
		return "", err
	}

	kindYaml := *resource

	value, manifestExists := kindYaml[key].(string)
	if !manifestExists {
		log.Warnf("failed to get '%s' from the manifest", key)
		return "", nil
	}

	return value, nil
}

// GetMetadata gets the namespace form the kubernetes resource.
func (resource *Resource) GetMetadata(dataMap string, key string, log *logrus.Logger) (string, error) {
	if err := yaml.Unmarshal([]byte(dataMap), resource); err != nil {
		return "", err
	}

	kindYaml := *resource

	metadata, metadataExists := kindYaml["metadata"].(map[string]interface{})
	if !metadataExists {
		log.Debug("failed to get 'metadata' from the manifest")

		return "", nil
	}

	value, failedManifest := metadata[key].(string)
	if !failedManifest {
		return "", &errors.DriftError{Message: fmt.Sprintf("failed to get %s from the metadata, '%s' is not type string", key, key)}
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

// NewResource returns aa new instance of ResourceInterface.
func NewResource() ResourceInterface {
	return &Resource{}
}
