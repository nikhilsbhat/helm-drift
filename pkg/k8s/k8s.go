package k8s

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type (
	Name map[string]interface{}
	Kind map[string]interface{}
)

type NameInterface interface {
	Get(dataMap string) (string, error)
}

type KindInterface interface {
	Get(dataMap string) (string, error)
}

//nolint:goerr113
func (name *Name) Get(dataMap string) (string, error) {
	var kindYaml map[string]interface{}
	if err := yaml.Unmarshal([]byte(dataMap), &kindYaml); err != nil {
		return "", err
	}
	if len(kindYaml) != 0 {
		value, ok := kindYaml["metadata"].(map[string]interface{})["name"].(string)
		if !ok {
			return "", fmt.Errorf("failed to get name from the manifest, 'name' is not type string")
		}

		return value, nil
	}

	return "", nil
}

//nolint:goerr113
func (kind *Kind) Get(dataMap string) (string, error) {
	var kindYaml map[string]interface{}
	if err := yaml.Unmarshal([]byte(dataMap), &kindYaml); err != nil {
		return "", err
	}
	if len(kindYaml) != 0 {
		value, ok := kindYaml["kind"].(string)
		if !ok {
			return "", fmt.Errorf("failed to get kube kind from the manifest, 'kind' is not type string")
		}

		return value, nil
	}

	return "", nil
}

func NewName() NameInterface {
	return &Name{}
}

func NewKind() KindInterface {
	return &Kind{}
}
