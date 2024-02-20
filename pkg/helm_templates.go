package pkg

import (
	"errors"
	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	driftErr "github.com/nikhilsbhat/helm-drift/pkg/errors"
	"github.com/nikhilsbhat/helm-drift/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v3"
	"log"
	"strings"
)

type (
	HelmTemplates []string
	HelmTemplate  string
)

// FilterBySkip filters the helm templates by list of skip kinds.
func (templates *HelmTemplates) FilterBySkip(drift *Drift) []string {
	return funk.Filter(*templates, func(tmpl string) bool {
		if len(drift.SkipKinds) == 0 {
			return true
		}

		kind, err := k8s.NewResource().GetKind(tmpl)
		if err != nil {
			log.Fatal(err)
		}

		return !funk.Contains(drift.SkipKinds, kind)
	}).([]string)
}

// FilterByKind filters the helm templates by selected kind.
func (templates *HelmTemplates) FilterByKind(drift *Drift) []string {
	return funk.Filter(*templates, func(tmpl string) bool {
		if len(drift.Kind) == 0 {
			return true
		}

		kind, err := k8s.NewResource().GetKind(tmpl)
		if err != nil {
			log.Fatal(err)
		}

		return funk.Contains(drift.Kind, kind)
	}).([]string)
}

// FilterByName filters the helm templates based on the name of the workload.
func (templates *HelmTemplates) FilterByName(drift *Drift) []string {
	return funk.Filter(*templates, func(tmpl string) bool {
		if len(drift.Name) == 0 {
			return true
		}

		name, err := k8s.NewResource().GetName(tmpl)
		if err != nil {
			log.Fatal(err)
		}

		return name == drift.Name
	}).([]string)
}

// FilterByHelmHook filters the helm templates based on the helm hooks part of the template.
func (templates *HelmTemplates) FilterByHelmHook(drift *Drift) []string {
	if drift.ConsiderHooks {
		return *templates
	}

	return funk.Filter(*templates, func(tmpl string) bool {
		hook, err := k8s.NewResource().IsHelmHook(tmpl, drift.IgnoreHookTypes)
		if err != nil {
			log.Fatal(err)
		}

		return !hook
	}).([]string)
}

func (templates *HelmTemplates) Get() ([]deviation.Deviation, error) {
	deviations := make([]deviation.Deviation, 0)

	for _, manifest := range *templates {
		name, err := k8s.NewResource().GetName(manifest)
		if err != nil {
			return nil, err
		}

		kind, err := k8s.NewResource().GetKind(manifest)
		if err != nil {
			return nil, err
		}

		deviations = append(deviations, deviation.Deviation{Resource: name, Kind: kind})
	}

	return deviations, nil
}

func (template *HelmTemplate) Get() (deviation.Deviation, error) {
	name, err := k8s.NewResource().GetName(string(*template))
	if err != nil {
		return deviation.Deviation{}, err
	}

	kind, err := k8s.NewResource().GetKind(string(*template))
	if err != nil {
		return deviation.Deviation{}, err
	}

	dvn := deviation.Deviation{Resource: name, Kind: kind}

	nameSpace, err := k8s.NewResource().GetNameSpace(name, kind, string(*template))

	notFoundErrType := &driftErr.NotFoundError{}

	if errors.Is(err, notFoundErrType) {
		return deviation.Deviation{}, err
	}

	if len(nameSpace) != 0 {
		dvn.NameSpace = nameSpace
	}

	return dvn, nil
}

func (template *HelmTemplate) DropStandardLabels(log *logrus.Logger) (string, error) {
	dvn, err := template.Get()
	if err != nil {
		return "", err
	}

	var manifestYaml map[string]interface{}
	if err = yaml.Unmarshal([]byte(*template), &manifestYaml); err != nil {
		return "", err
	}

	if len(manifestYaml) == 0 {
		return "", nil
	}

	_, annotationMissing := manifestYaml["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
	_, labelMissing := manifestYaml["metadata"].(map[string]interface{})["labels"].(map[string]interface{})

	if !annotationMissing || !labelMissing {
		log.Debugf(
			"failed to fetch annotations or labels from helm manifests, either annotations/labels are not present in resource %s with name %s", dvn.Kind, dvn.Resource,
		)

		yamlOut, err := yaml.Marshal(manifestYaml)
		if err != nil {
			return "", err
		}

		return string(yamlOut), nil
	}

	for key, _ := range manifestYaml["metadata"].(map[string]interface{})["annotations"].(map[string]interface{}) {
		if strings.Contains(key, "meta.helm") {
			delete(manifestYaml["metadata"].(map[string]interface{})["annotations"].(map[string]interface{}), key)
		}
	}

	for key, _ := range manifestYaml["metadata"].(map[string]interface{})["labels"].(map[string]interface{}) {
		switch key {
		case "app.kubernetes.io/name", "helm.sh/chart", "app.kubernetes.io/managed-by", "app.kubernetes.io/instance", "app.kubernetes.io/version", "app.kubernetes.io/component", "app.kubernetes.io/part-of":
			delete(manifestYaml["metadata"].(map[string]interface{})["labels"].(map[string]interface{}), key)
		}
	}

	yamlOut, err := yaml.Marshal(manifestYaml)
	if err != nil {
		return "", err
	}

	return string(yamlOut), nil
}

// NewHelmTemplate returns HelmTemplate of template string passed to it.
func NewHelmTemplate(template string) *HelmTemplate {
	helmTemplate := HelmTemplate(template)

	return &helmTemplate
}

// NewHelmTemplates returns HelmTemplates of template slices passed to it.
func NewHelmTemplates(templates []string) *HelmTemplates {
	helmTemplates := HelmTemplates(templates)

	return &helmTemplates
}
