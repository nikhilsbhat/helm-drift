package pkg

import (
	"errors"
	"log"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	driftError "github.com/nikhilsbhat/helm-drift/pkg/errors"
	"github.com/nikhilsbhat/helm-drift/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

type (
	HelmTemplates []string
	HelmTemplate  string
)

func (templates *HelmTemplates) FilterBySkip2(drift *Drift) []string {
	return funk.Filter(*templates, func(tmpl string) bool {
		if len(drift.SkipKinds) == 0 {
			return true
		}

		kind, err := k8s.NewResource().Get(tmpl, "kind", nil)
		if err != nil {
			log.Fatal(err)
		}

		return !funk.Contains(drift.SkipKinds, kind)
	}).([]string)
}

func (templates *HelmTemplates) FilterBySkip(drift *Drift) []string {
	return funk.Filter(*templates, func(tmpl string) bool {
		if len(drift.SkipKinds) == 0 {
			return true
		}

		kind, err := k8s.NewResource().Get(tmpl, "kind", nil)
		if err != nil {
			log.Fatal(err)
		}

		return !funk.Contains(drift.SkipKinds, kind)
	}).([]string)
}

func (templates *HelmTemplates) FilterByKind(drift *Drift) []string {
	return funk.Filter(*templates, func(tmpl string) bool {
		if len(drift.Kind) == 0 {
			return true
		}

		kind, err := k8s.NewResource().Get(tmpl, "kind", drift.log)
		if err != nil {
			log.Fatal(err)
		}

		return funk.Contains(drift.Kind, kind)
	}).([]string)
}

func (templates *HelmTemplates) FilterByName(drift *Drift) []string {
	return funk.Filter(*templates, func(tmpl string) bool {
		if len(drift.Name) == 0 {
			return true
		}

		name, err := k8s.NewResource().GetMetadata(tmpl, "name", drift.log)
		if err != nil {
			log.Fatal(err)
		}

		return name == drift.Name
	}).([]string)
}

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

func (templates *HelmTemplates) Get(log *logrus.Logger) ([]*deviation.Deviation, error) {
	deviations := make([]*deviation.Deviation, 0)

	for _, manifest := range *templates {
		name, err := k8s.NewResource().GetMetadata(manifest, "name", log)
		if err != nil {
			return nil, err
		}

		kind, err := k8s.NewResource().Get(manifest, "kind", log)
		if err != nil {
			return nil, err
		}

		deviations = append(deviations, &deviation.Deviation{Resource: name, Kind: kind})
	}

	return deviations, nil
}

func (template *HelmTemplate) Get(log *logrus.Logger) (*deviation.Deviation, error) {
	name, err := k8s.NewResource().GetMetadata(string(*template), "name", log)
	if err != nil {
		return nil, err
	}

	kind, err := k8s.NewResource().Get(string(*template), "kind", log)
	if err != nil {
		return nil, err
	}

	apiVersion, err := k8s.NewResource().Get(string(*template), "apiVersion", log)
	if err != nil {
		return nil, err
	}

	dvn := &deviation.Deviation{Resource: name, Kind: kind, APIVersion: apiVersion}

	nameSpace, err := k8s.NewResource().GetMetadata(string(*template), "namespace", log)

	notFoundErrType := &driftError.NotFoundError{}

	if errors.Is(err, notFoundErrType) {
		return nil, err
	}

	if len(nameSpace) != 0 {
		dvn.NameSpace = nameSpace
	}

	return dvn, nil
}

func (template *HelmTemplate) GetTemplate() string {
	val := *template

	return string(val)
}

func NewHelmTemplate(template string) *HelmTemplate {
	helmTemplate := HelmTemplate(template)

	return &helmTemplate
}

func NewHelmTemplates(templates []string) *HelmTemplates {
	helmTemplates := HelmTemplates(templates)

	return &helmTemplates
}
