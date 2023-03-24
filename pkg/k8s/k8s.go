package k8s

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type (
	Name map[string]interface{}
)

type NameInterface interface {
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

func NewName() NameInterface {
	return &Name{}
}
