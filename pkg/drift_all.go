package pkg

import (
	"time"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
)

func (drift *Drift) GetAllDrift() error {
	startTime := time.Now()

	if err := drift.cleanManifests(true); err != nil {
		drift.log.Fatalf("cleaning old rendered files failed with: %v", err)
	}

	drift.log.Debugf("got all required values to identify drifts from chart/release '%s' proceeding furter to fetch the same", drift.release)

	drift.setNameSpace()

	if err := drift.setExternalDiff(); err != nil {
		return err
	}

	releases, err := drift.getChartsFromReleases()
	if err != nil {
		return err
	}

	defer func(drift *Drift) {
		if err = drift.cleanManifests(false); err != nil {
			drift.log.Fatalf("cleaning rendered files failed with: %v", err)
		}
	}(drift)

	driftedReleases := make([]deviation.DriftedRelease, 0)

	for _, release := range releases {
		drift.log.Debugf("identifying drifts for release '%s'", release.Name)

		kubeKindTemplates := drift.getTemplates([]byte(release.Manifest))

		deviations, err := drift.renderToDisk(kubeKindTemplates, "", release.Name, release.Namespace)
		if err != nil {
			return err
		}

		out, err := drift.Diff(deviations)
		if err != nil {
			return err
		}

		if len(out.Deviations) == 0 {
			drift.log.Infof("no drifts identified for relase '%s'", release.Name)

			continue
		}

		driftedReleases = append(driftedReleases, out)
	}

	drift.timeSpent = time.Since(startTime).Seconds()

	return drift.render(driftedReleases)
}
