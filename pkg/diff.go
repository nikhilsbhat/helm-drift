package pkg

import (
	"fmt"
	"strings"
	"sync"

	"github.com/nikhilsbhat/helm-drift/pkg/command"
	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	"github.com/nikhilsbhat/helm-drift/pkg/errors"
)

func (drift *Drift) Diff(renderedManifests *deviation.DriftedRelease) (*deviation.DriftedRelease, error) {
	var (
		waitGroup sync.WaitGroup
		errChan   = make(chan error, len(renderedManifests.Deviations))
		diffs     = make([]*deviation.Deviation, len(renderedManifests.Deviations))
		sem       = make(chan struct{}, func() int {
			if drift.Limit != 0 {
				return drift.Limit
			}

			return len(renderedManifests.Deviations)
		}())
	)

	if drift.Limit != 0 {
		drift.log.Infof(
			"limit on concurrency is set to '%d', so batching the 'kubectl diff' executions", drift.Limit,
		)
	}

	waitGroup.Add(len(renderedManifests.Deviations))

	go func() {
		waitGroup.Wait()
		close(errChan)
	}()

	handleError := func(err error) {
		if err != nil {
			drift.log.Error(err)
			errChan <- err
		}
	}

	for index, dvn := range renderedManifests.Deviations {
		sem <- struct{}{}

		go func(index int, dvn *deviation.Deviation) {
			defer waitGroup.Done()
			defer func() { <-sem }()

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
			handleError(err)

			cmd := command.NewCommand("kubectl", drift.log)

			cmd.SetKubeDiffCmd(drift.kubeConfig, drift.kubeContext, nameSpace, arguments...)

			dft, err := cmd.RunKubeDiffCmd(dvn)
			handleError(err)

			if !isManagedByHPA {
				if dft.HasDrift {
					renderedManifests.HasDrift = true
				}

				diffs[index] = dft

				return
			}

			hasOnlyChangesScaledByHpa, err := drift.HasOnlyChangesScaledByHpa(dft.Deviations)
			handleError(err)

			if dft.HasDrift && (!hasOnlyChangesScaledByHpa || !drift.IgnoreHPAChanges) {
				renderedManifests.HasDrift = true
			} else {
				dft.HasDrift = false
				dft.Deviations = ""
			}

			diffs[index] = dft
		}(index, dvn)
	}

	var diffErrors []string

	for errCh := range errChan {
		if errCh != nil {
			diffErrors = append(diffErrors, errCh.Error())
		}
	}

	if len(diffErrors) != 0 {
		return nil, &errors.DriftError{Message: fmt.Sprintf("calculating diff errored with: %s", strings.Join(diffErrors, "\n"))}
	}

	renderedManifests.Deviations = diffs

	drift.log.Debugf("ran diffs for all manifests for release '%s' successfully", renderedManifests.Release)

	return renderedManifests, nil
}

func (drift *Drift) setNameSpace(releaseNameSpace *deviation.DriftedRelease, manifestNameSpace *deviation.Deviation) string {
	if len(manifestNameSpace.NameSpace) != 0 {
		drift.log.Debugf("manifest is not deployed in a helm release's namespace, it is set to '%s'. "+
			"So considering this namespace for identifying drifts in manifest '%s'", manifestNameSpace.NameSpace, manifestNameSpace.TemplatePath)

		return manifestNameSpace.NameSpace
	}

	return releaseNameSpace.Namespace
}
