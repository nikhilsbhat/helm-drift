package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	"github.com/nikhilsbhat/helm-drift/pkg/k8s"
)

const (
	templatePathPermission = 0o755
	manifestFilePermission = 0o644
)

func (drift *Drift) renderToDisk(manifests []string, releaseName, releaseNamespace any) (deviation.DriftedReleases, error) {
	releaseDrifted := deviation.DriftedReleases{
		Namespace: releaseNamespace.(string),
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

	for _, manifest := range manifests {
		name, err := k8s.NewName().Get(manifest)
		if err != nil {
			return releaseDrifted, err
		}

		kind, err := k8s.NewKind().Get(manifest)
		if err != nil {
			return releaseDrifted, err
		}

		drift.log.Debugf("generating manifest %s", name)

		manifestPath := filepath.Join(templatePath, fmt.Sprintf("%s.%s.yaml", name, kind))
		if err = os.WriteFile(manifestPath, []byte(manifest), manifestFilePermission); err != nil {
			return releaseDrifted, err
		}

		dvn := deviation.Deviation{
			Kind:         kind,
			Resource:     name,
			TemplatePath: templatePath,
			ManifestPath: manifestPath,
		}
		deviations = append(deviations, dvn)
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
