package pkg

import (
	"fmt"
	"strings"
	"sync"

	"github.com/nikhilsbhat/helm-drift/pkg/command"
	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	"github.com/nikhilsbhat/helm-drift/pkg/errors"
)

func (drift *Drift) Diff(renderedManifests deviation.DriftedRelease) (deviation.DriftedRelease, error) {
	diffs := make([]deviation.Deviation, 0)

	var drifted bool

	var waitGroup sync.WaitGroup

	errChan := make(chan error)

	waitGroup.Add(len(renderedManifests.Deviations))

	go func() {
		waitGroup.Wait()
		close(errChan)
	}()

	for _, dvn := range renderedManifests.Deviations {
		go func(dvn deviation.Deviation) {
			defer waitGroup.Done()

			manifestPath := dvn.ManifestPath

			drift.log.Debugf("calculating diff for %s", manifestPath)

			arguments := []string{fmt.Sprintf("-f=%s", manifestPath)}

			cmd := command.NewCommand("kubectl", drift.log)

			nameSpace := drift.setNameSpace(renderedManifests, dvn)
			drift.log.Debugf("setting namespace to %s", nameSpace)

			cmd.SetKubeCmd(nameSpace, arguments...)

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

	renderedManifests.Deviations = diffs
	renderedManifests.HasDrift = drifted

	drift.log.Debugf("ran diffs for all manifests for release '%s' successfully", renderedManifests.Release)

	return renderedManifests, nil
}

func (drift *Drift) setNameSpace(releaseNameSpace deviation.DriftedRelease, manifestNameSpace deviation.Deviation) string {
	if len(manifestNameSpace.NameSpace) != 0 {
		drift.log.Debugf("manifest is not deployed in a helm release's namespace, it is set to '%s'. "+
			"So considering this namespace for identifying drifts in manifest '%s'", manifestNameSpace.NameSpace, manifestNameSpace.TemplatePath)

		return manifestNameSpace.NameSpace
	}

	return releaseNameSpace.Namespace
}
