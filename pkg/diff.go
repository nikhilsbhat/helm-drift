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
	var (
		waitGroup sync.WaitGroup
		errChan   = make(chan error, len(renderedManifests.Deviations))
		diffs     = make([]deviation.Deviation, len(renderedManifests.Deviations))
	)

	waitGroup.Add(len(renderedManifests.Deviations))

	go func() {
		waitGroup.Wait()
		close(errChan)
	}()

	for index, dvn := range renderedManifests.Deviations {
		go func(index int, dvn deviation.Deviation) {
			defer waitGroup.Done()

			manifestPath := dvn.ManifestPath

			drift.log.Debugf("calculating diff for %s", manifestPath)

			arguments := []string{
				"--show-managed-fields=false",
				fmt.Sprintf("--concurrency=%d", drift.Concurrency),
				fmt.Sprintf("-f=%s", manifestPath),
			}

			nameSpace := drift.setNameSpace(renderedManifests, dvn)
			drift.log.Debugf("setting namespace to %s", nameSpace)

			isManagedByHPA, err := drift.IsManagedByHPA(dvn.Resource, dvn.Kind, nameSpace)
			if err != nil {
				drift.log.Error(err)

				errChan <- err
			}

			cmd := command.NewCommand("kubectl", drift.log)

			cmd.SetKubeDiffCmd(drift.kubeConfig, drift.kubeContext, nameSpace, arguments...)

			dft, err := cmd.RunKubeDiffCmd(dvn)
			if err != nil {
				drift.log.Error(err)

				errChan <- err
			}

			if !isManagedByHPA {
				if dft.HasDrift {
					renderedManifests.HasDrift = true
				}

				diffs[index] = dft

				return
			}

			wasHpaScaled, err := drift.WasScaledByHpa(dft.Deviations)
			if err != nil {
				drift.log.Error(err)

				errChan <- err
			}

			if (dft.HasDrift && !wasHpaScaled) || (dft.HasDrift && wasHpaScaled && !drift.IgnoreHPAChanges) {
				renderedManifests.HasDrift = true

				diffs[index] = dft
			}
		}(index, dvn)
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
