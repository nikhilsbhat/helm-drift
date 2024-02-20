package pkg

import (
	"fmt"
	"github.com/sters/yaml-diff/yamldiff"
	"os"
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

			cmd.SetKubeDiffCmd(drift.kubeConfig, drift.kubeContext, nameSpace, arguments...)

			dft, err := cmd.RunKubeDiffCmd(dvn)
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

func (drift *Drift) NewDiff(renderedManifests deviation.DriftedRelease) (deviation.DriftedRelease, error) {
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

			arguments := []string{dvn.Kind, dvn.Resource, "-o=yaml", "--ignore-not-found=true"}

			cmd := command.NewCommand("kubectl", drift.log)

			nameSpace := drift.setNameSpace(renderedManifests, dvn)
			drift.log.Debugf("setting namespace to %s", nameSpace)

			cmd.SetKubeGetCmd(drift.kubeConfig, drift.kubeContext, nameSpace, arguments...)

			dft, err := cmd.RunKubeGetCmd(dvn)
			if err != nil {
				drift.log.Errorf("running kubectl command to get manifests errored with:  %v", err)
				errChan <- err
			}

			kubeNeatFile, err := drift.neat(dft)
			if err != nil {
				errChan <- err
			}

			renderedFile, err := os.ReadFile(dvn.ManifestPath)
			if err != nil {
				errChan <- err
			}

			helmFile, err := yamldiff.Load(string(renderedFile))
			if err != nil {
				errChan <- err
			}

			kubeFile, err := yamldiff.Load(kubeNeatFile)
			if err != nil {
				errChan <- err
			}

			opts := make([]yamldiff.DoOptionFunc, 0)
			for _, diff := range yamldiff.Do(kubeFile, helmFile, opts...) {
				fmt.Println("#####################################################################################")
				fmt.Println(diff.Dump())
			}

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
