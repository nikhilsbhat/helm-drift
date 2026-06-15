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

	if err := drift.setExternalDiff(); err != nil {
		drift.log.Fatalf("%v", err)
	}

	releases, err := drift.getChartsFromReleases()
	if err != nil {
		drift.log.Fatalf("%v", err)
	}

	releases = resourcesToSkip(drift.releasesToSkip).filterRelease(releases)

	defer func(drift *Drift) {
		if err = drift.cleanManifests(false); err != nil {
			drift.log.Fatalf("cleaning rendered files failed with: %v", err)
		}
	}(drift)

	driftedReleases := make([]*deviation.DriftedRelease, len(releases))

	sem := make(chan struct{}, drift.releaseConcurrencyLimit(releases))

	if drift.Limit != 0 {
		drift.log.Debugf("limit on concurrency is set to '%d', so batching the executions of helm releases", drift.Limit)
	}

	var waitGroup sync.WaitGroup

	errChan := make(chan error)

	waitGroup.Add(len(releases))

	go func() {
		waitGroup.Wait()
		close(errChan)
	}()

	for index, release := range releases {
		sem <- struct{}{}

		go func(index int, release *helmRelease.Release) {
			defer waitGroup.Done()
			defer func() { <-sem }()

			drift.log.Debugf("identifying drifts for release '%s'", release.Name)

			kubeKindTemplates := drift.getTemplates([]byte(release.Manifest))

			deviations, err := drift.renderToDisk(kubeKindTemplates, "", release.Name, release.Namespace)
			if err != nil {
				errChan <- err

				return
			}

			out, err := drift.Diff(deviations)
			if err != nil {
				errChan <- err

				return
			}

			if len(out.Deviations) == 0 && err == nil {
				drift.log.Infof("no drifts identified for relase '%s'", release.Name)

				return
			}

			driftedReleases[index] = out
		}(index, release)
	}

	if driftErrors := collectErrors(errChan); len(driftErrors) != 0 {
		drift.log.Fatalf("%v", &errors.DriftError{Message: fmt.Sprintf("identifying drifts errored with: %s", strings.Join(driftErrors, "\n"))})
	}

	drift.timeSpent = time.Since(startTime).Seconds()

	if err = drift.render(filterDriftedReleases(driftedReleases)); err != nil {
		drift.log.Fatalf("%v", err)
	}
}

func (drift *Drift) releaseConcurrencyLimit(releases []*helmRelease.Release) int {
	if drift.Limit != 0 {
		return drift.Limit
	}

	return len(releases)
}

func filterDriftedReleases(driftedReleases []*deviation.DriftedRelease) []*deviation.DriftedRelease {
	filteredDriftedReleases := make([]*deviation.DriftedRelease, 0, len(driftedReleases))

	for _, driftedRelease := range driftedReleases {
		if driftedRelease != nil {
			filteredDriftedReleases = append(filteredDriftedReleases, driftedRelease)
		}
	}

	return filteredDriftedReleases
}
