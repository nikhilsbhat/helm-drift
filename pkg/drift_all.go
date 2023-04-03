package pkg

import (
	"encoding/json"
	"fmt"
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

	charts, err := drift.getChartsFromReleases()
	if err != nil {
		return err
	}

	defer func(drift *Drift) {
		if err = drift.cleanManifests(false); err != nil {
			drift.log.Fatalf("cleaning rendered files failed with: %v", err)
		}
	}(drift)

	chartsDeviations := make(map[string]deviation.DriftedReleases)

	for _, chart := range charts {
		kubeKindTemplates := drift.getTemplates([]byte(chart.Manifest))

		deviations, err := drift.renderToDisk(kubeKindTemplates, chart.Name, chart.Namespace)
		if err != nil {
			return err
		}

		out, err := drift.Diff(deviations)
		if err != nil {
			return err
		}

		if len(out.Deviations) == 0 {
			drift.log.Infof("no drifts identified for relase '%s'", chart.Name)

			continue
		}

		chartsDeviations[chart.Name] = out
	}

	drift.timeSpent = time.Since(startTime).Seconds()

	jsonOut, err := json.MarshalIndent(chartsDeviations, "", " ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonOut))

	return nil
}
