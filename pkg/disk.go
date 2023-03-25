package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nikhilsbhat/helm-drift/pkg/k8s"
)

const (
	templatePathPermission = 0o755
	manifestFilePermission = 0o644
)

func (drift *Drift) renderToDisk(manifests []string) ([]Deviation, error) {
	templatePath := filepath.Join(drift.TempPath, drift.release)

	drift.log.Debugf("rendering helm manifests to disk under %s", templatePath)
	drift.log.Debugf("creating directories '%s' to generate manifests", templatePath)
	if err := os.MkdirAll(templatePath, templatePathPermission); err != nil {
		return nil, err
	}

	deviations := make([]Deviation, 0)
	for _, manifest := range manifests {
		name, err := k8s.NewName().Get(manifest)
		if err != nil {
			return nil, err
		}

		kind, err := k8s.NewKind().Get(manifest)
		if err != nil {
			return nil, err
		}

		drift.log.Debugf("generating manifest %s", name)

		manifestPath := filepath.Join(templatePath, fmt.Sprintf("%s.yaml", name))
		if err = os.WriteFile(manifestPath, []byte(manifest), manifestFilePermission); err != nil {
			return nil, err
		}

		deviation := Deviation{
			Kind:         kind,
			Resource:     name,
			TemplatePath: templatePath,
			ManifestPath: manifestPath,
		}
		deviations = append(deviations, deviation)
	}

	drift.log.Debug("all manifests rendered to disk successfully")

	return deviations, nil
}

func (drift *Drift) cleanManifests() error {
	templatePath := filepath.Join(drift.TempPath, drift.release)

	if !drift.SkipClean {
		if err := os.RemoveAll(templatePath); err != nil {
			return err
		}
		drift.log.Debug("all manifests rendered to disk was cleaned")
	} else {
		drift.log.Debug("rendered manifests deletion skipped as it was disabled")
	}

	return nil
}
