package k8s

import (
	"strings"

	"github.com/thoas/go-funk"
	"sigs.k8s.io/yaml"
)

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
