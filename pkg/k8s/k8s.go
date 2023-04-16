package k8s

import (
	"github.com/nikhilsbhat/helm-drift/pkg/errors"

	"gopkg.in/yaml.v3"
)

type (
	Resource map[string]interface{}
)

// ResourceInterface implements methods to get resource name and kind.
type ResourceInterface interface {
	GetName(dataMap string) (string, error)
	GetKind(dataMap string) (string, error)
}

// GetName gets the name form the kubernetes resource.
func (resource *Resource) GetName(dataMap string) (string, error) {
	var kindYaml map[string]interface{}
	if err := yaml.Unmarshal([]byte(dataMap), &kindYaml); err != nil {
		return "", err
	}

	if len(kindYaml) != 0 {
		value, ok := kindYaml["metadata"].(map[string]interface{})["name"].(string)
		if !ok {
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
		value, ok := kindYaml["kind"].(string)
		if !ok {
			return "", &errors.DriftError{Message: "failed to get kube kind from the manifest, 'kind' is not type string"}
		}

		return value, nil
	}

	return "", nil
}

// NewResource returns aa new instance of ResourceInterface
func NewResource() ResourceInterface {
	return &Resource{}
}
