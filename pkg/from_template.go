package pkg

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

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

func (drift *Drift) getChartFromTemplate() ([]byte, error) {
	flags := make([]string, 0)
	for _, value := range drift.Values {
		flags = append(flags, "--set", value)
	}

	for _, stringValue := range drift.StringValues {
		flags = append(flags, "--set-string", stringValue)
	}

	for _, fileValue := range drift.FileValues {
		flags = append(flags, "--set-file", fileValue)
	}

	for _, valueFile := range drift.ValueFiles {
		flags = append(flags, "--values", valueFile)
	}

	if strings.ToLower(drift.LogLevel) == logrus.DebugLevel.String() {
		flags = append(flags, "--debug")
	}

	if drift.SkipTests {
		flags = append(flags, "--skip-tests")
	}

	if drift.SkipCRDS {
		flags = append(flags, "--skip-crds")
	}

	if drift.Validate {
		flags = append(flags, "--validate")
	}

	if len(drift.Version) != 0 {
		flags = append(flags, "--version", drift.Version)
	}

	args := []string{"template", drift.release, drift.chart}
	args = append(args, flags...)

	drift.log.Debugf("rendering helm chart with following commands/flags '%s'", strings.Join(args, ", "))

	cmd := exec.Command(os.Getenv("HELM_BIN"), args...) //nolint:gosec
	output, err := cmd.Output()

	var exitErr *exec.ExitError

	if errors.As(err, &exitErr) {
		drift.log.Errorf("rendering template for release: '%s' errored with %v", drift.release, err)

		return nil, fmt.Errorf("%w: %s", exitErr, exitErr.Stderr)
	}

	var pathErr *fs.PathError

	if errors.As(err, &pathErr) {
		drift.log.Error("locating helm cli errored with", err)

		return nil, fmt.Errorf("%w: %s", pathErr, pathErr.Path)
	}

	return output, nil
}

func (drift *Drift) getTemplates(template []byte) []string {
	drift.log.Debugf("splitting helm manifests with regex pattern: '%s'", drift.Regex)
	temp := regexp.MustCompile(drift.Regex)
	kinds := temp.Split(string(template), -1)
	// Removing empty string at the beginning as splitting string always adds it in front.
	kinds = kinds[1:]

	return kinds
}

func (templates *HelmTemplates) FilterBySkip(drift *Drift) []string {
	return funk.Filter(*templates, func(tmpl string) bool {
		if len(drift.SkipKinds) == 0 {
			return true
		}

		kind, err := k8s.NewResource().GetKind(tmpl, nil)
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

		kind, err := k8s.NewResource().GetKind(tmpl, drift.log)
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

		name, err := k8s.NewResource().GetName(tmpl, drift.log)
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

func (templates *HelmTemplates) Get(log *logrus.Logger) ([]deviation.Deviation, error) {
	deviations := make([]deviation.Deviation, 0)

	for _, manifest := range *templates {
		name, err := k8s.NewResource().GetName(manifest, log)
		if err != nil {
			return nil, err
		}

		kind, err := k8s.NewResource().GetKind(manifest, log)
		if err != nil {
			return nil, err
		}

		deviations = append(deviations, deviation.Deviation{Resource: name, Kind: kind})
	}

	return deviations, nil
}

func (template *HelmTemplate) Get(log *logrus.Logger) (deviation.Deviation, error) {
	name, err := k8s.NewResource().GetName(string(*template), log)
	if err != nil {
		return deviation.Deviation{}, err
	}

	kind, err := k8s.NewResource().GetKind(string(*template), log)
	if err != nil {
		return deviation.Deviation{}, err
	}

	dvn := deviation.Deviation{Resource: name, Kind: kind}

	nameSpace, err := k8s.NewResource().GetNameSpace(name, kind, string(*template), log)

	notFoundErrType := &driftError.NotFoundError{}

	if errors.Is(err, notFoundErrType) {
		return deviation.Deviation{}, err
	}

	if len(nameSpace) != 0 {
		dvn.NameSpace = nameSpace
	}

	return dvn, nil
}

func NewHelmTemplate(template string) *HelmTemplate {
	helmTemplate := HelmTemplate(template)

	return &helmTemplate
}

func NewHelmTemplates(templates []string) *HelmTemplates {
	helmTemplates := HelmTemplates(templates)

	return &helmTemplates
}
