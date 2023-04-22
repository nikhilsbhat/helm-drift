package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	"github.com/nikhilsbhat/helm-drift/pkg/errors"
)

const (
	templatePathPermission = 0o755
	manifestFilePermission = 0o644
)

func (drift *Drift) renderToDisk(manifests []string, chartName, releaseName, releaseNamespace any) (deviation.DriftedRelease, error) {
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
		return releaseDrifted, err
	}

	deviations := make([]deviation.Deviation, 0)

	var waitGroup sync.WaitGroup

	errChan := make(chan error)

	waitGroup.Add(len(manifests))

	go func() {
		waitGroup.Wait()
		close(errChan)
	}()

	for _, manifest := range manifests {
		go func(manifest string) {
			defer waitGroup.Done()

			template, err := NewHelmTemplate(manifest).Get()
			if err != nil {
				errChan <- err
			}

			drift.log.Debugf("generating manifest '%s'", template.Resource)

			manifestPath := filepath.Join(templatePath, fmt.Sprintf("%s.%s.yaml", template.Resource, template.Kind))
			if err = os.WriteFile(manifestPath, []byte(manifest), manifestFilePermission); err != nil {
				errChan <- err
			}

			drift.log.Debugf("manifest for '%s' generated successfully", template.Resource)

			dvn := deviation.Deviation{
				Kind:         template.Kind,
				Resource:     template.Resource,
				TemplatePath: templatePath,
				ManifestPath: manifestPath,
			}
			deviations = append(deviations, dvn)
		}(manifest)
	}

	var diskErrors []string

	for err := range errChan {
		if err != nil {
			diskErrors = append(diskErrors, err.Error())
		}
	}

	if len(diskErrors) != 0 {
		return deviation.DriftedRelease{},
			&errors.DriftError{Message: fmt.Sprintf("rendering helm manifests to disk errored: %s", strings.Join(diskErrors, "\n"))}
	}

	if len(deviations) != len(manifests) {
		resourceFromManifests, err := NewHelmTemplates(manifests).Get()
		if err != nil {
			return deviation.DriftedRelease{}, err
		}

		return deviation.DriftedRelease{}, &errors.NotAllError{Manifests: resourceFromManifests, ResourceFromDeviations: deviations}
	}

	releaseDrifted.Deviations = deviations

	drift.log.Debug("all manifests rendered to disk successfully")

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
