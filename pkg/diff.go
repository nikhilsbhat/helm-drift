package pkg

import (
	"fmt"
	"strings"
	"sync"

	"github.com/nikhilsbhat/helm-drift/pkg/command"
	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	"github.com/nikhilsbhat/helm-drift/pkg/errors"
)

func (drift *Drift) Diff(driftedRelease deviation.DriftedRelease) (deviation.DriftedRelease, error) {
	diffs := make([]deviation.Deviation, 0)

	var drifted bool

	var waitGroup sync.WaitGroup

	errChan := make(chan error)

	waitGroup.Add(len(driftedRelease.Deviations))

	go func() {
		waitGroup.Wait()
		close(errChan)
	}()

	for _, dvn := range driftedRelease.Deviations {
		go func(dvn deviation.Deviation) {
			defer waitGroup.Done()

			manifestPath := dvn.ManifestPath

			drift.log.Debugf("calculating diff for %s", manifestPath)

			arguments := []string{fmt.Sprintf("-f=%s", manifestPath)}

			cmd := command.NewCommand("kubectl", drift.log)

			cmd.SetKubeCmd(driftedRelease.Namespace, arguments...)

			dft, err := cmd.RunKubeCmd(dvn)
			if err != nil {
				errChan <- err
			}

			if dft.HasDrift {
				drifted = dft.HasDrift
			}

			diffs = append(diffs, dft)
		}(dvn)
	}

	var diffErrors []string

	for errCh := range errChan {
		if errCh != nil {
			diffErrors = append(diffErrors, errCh.Error())
		}
	}

	if len(diffErrors) != 0 {
		return deviation.DriftedRelease{}, &errors.DriftError{Message: fmt.Sprintf("calculating diff errored with: %s", strings.Join(diffErrors, "\n"))}
	}

	driftedRelease.Deviations = diffs
	driftedRelease.HasDrift = drifted

	drift.log.Debugf("ran diffs for all manifests for release '%s' successfully", driftedRelease.Release)

	return driftedRelease, nil
}
