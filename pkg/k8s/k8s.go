package k8s

import (
	"fmt"
	"strings"

	"github.com/nikhilsbhat/helm-drift/pkg/errors"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v3"
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

// IsHelmHook gets the namespace form the kubernetes resource.
func (resource *Resource) IsHelmHook(dataMap string, hookKinds []string) (bool, error) {
	var kindYaml map[string]interface{}
	if err := yaml.Unmarshal([]byte(dataMap), &kindYaml); err != nil {
		return false, err
	}

	if len(kindYaml) == 0 {
		return false, nil
	}

	if _, failedManifest := kindYaml["metadata"].(map[string]interface{})["annotations"]; !failedManifest {
		return false, nil
	}

	if _, failedManifest := kindYaml["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})["helm.sh/hook"].(string); !failedManifest {
		return false, &errors.NotFoundError{Key: "failed to identify the manifest as chart hook"}
	}

	hookType, failedManifest := kindYaml["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})["helm.sh/hook-delete-policy"].(string)
	if !failedManifest {
		return false, &errors.NotFoundError{Key: "failed to identify the the chart hook type from the manifest"}
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
