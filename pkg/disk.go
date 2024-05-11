package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	"github.com/nikhilsbhat/helm-drift/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	templatePathPermission = 0o755
	manifestFilePermission = 0o644
)

func (drift *Drift) renderToDisk(manifests []string, chartName, releaseName, releaseNamespace any) (deviation.DriftedRelease, error) {
	manifests = NewHelmTemplates(manifests).FilterByHelmHook(drift)
	manifests = NewHelmTemplates(manifests).FilterBySkip(drift)
	manifests = NewHelmTemplates(manifests).FilterByKind(drift)
	manifests = NewHelmTemplates(manifests).FilterByName(drift)

	releaseDrifted := deviation.DriftedRelease{
		Namespace: releaseNamespace.(string),
		Release:   releaseName.(string),
		Chart:     chartName.(string),
	}

	templatePath := filepath.Join(drift.TempPath, drift.release)
	if drift.All {
		templatePath = filepath.Join(drift.TempPath, "all", releaseName.(string))
	}

	drift.log.Debugf("rendering helm manifests to disk under %s", templatePath)
	drift.log.Debugf("creating directories '%s' to generate manifests", templatePath)

	if err := os.MkdirAll(templatePath, templatePathPermission); err != nil {
		log.Errorf("creating template path '%s' errored with '%v'", templatePath, err)

		return releaseDrifted, err
	}

	templates := make([]deviation.Deviation, 0)

	for _, manifest := range manifests {
		drift.log.Debugf("rendering manifest to disc, hold on for a moment....")

		template, err := NewHelmTemplate(manifest).Get(drift.log)
		if err != nil {
			log.Errorf("getting manifest information from template errored with '%v'", err)

			return deviation.DriftedRelease{}, err
		}

		drift.log.Debugf("generating manifest '%s'", template.Resource)

		manifestPath := filepath.Join(templatePath, fmt.Sprintf("%s.%s.%s.yaml", template.Resource, template.Kind, releaseName))
		if err = os.WriteFile(manifestPath, []byte(manifest), manifestFilePermission); err != nil {
			log.Errorf("writing manifest '%s' to disk errored with '%v'", manifestPath, err)

			return deviation.DriftedRelease{}, err
		}

		drift.log.Debugf("manifest for '%s' generated successfully", template.Resource)

		dvn := deviation.Deviation{
			Kind:         template.Kind,
			Resource:     template.Resource,
			NameSpace:    template.NameSpace,
			TemplatePath: templatePath,
			ManifestPath: manifestPath,
		}

		templates = append(templates, dvn)
	}

	if len(templates) != len(manifests) {
		resourceFromManifests, err := NewHelmTemplates(manifests).Get(drift.log)
		if err != nil {
			log.Errorf("getting manifests information from templates errored with '%v'", err)

			return deviation.DriftedRelease{}, err
		}

		return deviation.DriftedRelease{}, &errors.NotAllError{Manifests: resourceFromManifests, ResourceFromDeviations: templates}
	}

	releaseDrifted.Deviations = templates

	drift.log.Debugf("all manifests from release '%s' was successfully rendered to disk...", releaseName.(string))

	return releaseDrifted, nil
}

func (drift *Drift) cleanManifests(force bool) error {
	templatePath := filepath.Join(drift.TempPath, drift.release)
	if drift.All {
		templatePath = filepath.Join(drift.TempPath, "all")
	}

	if !drift.SkipClean || force {
		if err := os.RemoveAll(templatePath); err != nil {
			return err
		}

		drift.log.Debug("all manifests rendered to disk was cleaned")
	} else {
		drift.log.Debug("rendered manifests deletion skipped as it was disabled")
	}

	return nil
}
