package pkg

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	"github.com/nikhilsbhat/helm-drift/pkg/errors"
	helmRelease "helm.sh/helm/v3/pkg/release"
)

func (drift *Drift) GetAllDrift() {
	startTime := time.Now()

	if err := drift.cleanManifests(true); err != nil {
		drift.log.Fatalf("cleaning old rendered files failed with: %v", err)
	}

	drift.log.Debugf("got all required values to identify drifts from chart/release '%s' proceeding furter to fetch the same", drift.release)

	drift.setNameSpace()

	if err := drift.setExternalDiff(); err != nil {
		drift.log.Fatalf("%v", err)
	}

	releases, err := drift.getChartsFromReleases()
	if err != nil {
		drift.log.Fatalf("%v", err)
	}

	defer func(drift *Drift) {
		if err = drift.cleanManifests(false); err != nil {
			drift.log.Fatalf("cleaning rendered files failed with: %v", err)
		}
	}(drift)

	driftedReleases := make([]deviation.DriftedRelease, 0)

	var waitGroup sync.WaitGroup

	errChan := make(chan error)

	waitGroup.Add(len(releases))

	go func() {
		waitGroup.Wait()
		close(errChan)
	}()

	for _, release := range releases {
		go func(release *helmRelease.Release) {
			defer waitGroup.Done()
			drift.log.Debugf("identifying drifts for release '%s'", release.Name)

			kubeKindTemplates := drift.getTemplates([]byte(release.Manifest))

			deviations, err := drift.renderToDisk(kubeKindTemplates, "", release.Name, release.Namespace)
			if err != nil {
				errChan <- err
			}

			out, err := drift.Diff(deviations)
			if err != nil {
				errChan <- err
			}

			if len(out.Deviations) == 0 {
				drift.log.Infof("no drifts identified for relase '%s'", release.Name)

				return
			}

			driftedReleases = append(driftedReleases, out)
		}(release)
	}

	var driftErrors []string

	for errCh := range errChan {
		if errCh != nil {
			driftErrors = append(driftErrors, errCh.Error())
		}
	}

	if len(driftErrors) != 0 {
		drift.log.Fatalf("%v", &errors.DriftError{Message: fmt.Sprintf("identifying drifts errored with: %s", strings.Join(driftErrors, "\n"))})
	}

	drift.timeSpent = time.Since(startTime).Seconds()

	if err = drift.render(driftedReleases); err != nil {
		drift.log.Fatalf("%v", err)
	}
}
