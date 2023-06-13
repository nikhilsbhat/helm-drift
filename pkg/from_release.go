package pkg

import (
	"log"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

// getChartFromRelease should get the manifest from the selected release.
func (drift *Drift) getChartFromRelease() ([]byte, error) {
	settings := cli.New()

	drift.log.Debugf("fetching chart manifest for release '%s' from kube cluster", drift.release)

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		drift.log.Error("oops initialising helm client errored with", err)

		return nil, err
	}

	client := action.NewGet(actionConfig)

	helmRelease, err := client.Run(drift.release)
	if err != nil {
		drift.log.Errorf("fetching helm release '%s' errored with '%v'", drift.release, err)

		return nil, err
	}

	drift.log.Debugf("chart manifest for release '%s' was successfully retrieved from kube cluster", drift.release)

	return []byte(helmRelease.Manifest), nil
}

func (drift *Drift) getChartsFromReleases() ([]*release.Release, error) {
	settings := cli.New()

	drift.log.Debug("fetching all helm releases from kube cluster")

	var namespace string

	if drift.isAll() {
		drift.log.Debug("no namespace specified, fetching all helm releases from the the cluster")
	} else {
		drift.log.Debugf("retrieving charts from the namespace '%s'", drift.namespace)
		namespace = drift.namespace
	}

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		drift.log.Error("oops initialising helm client errored with", err)

		return nil, err
	}

	client := action.NewList(actionConfig)

	return client.Run()
}

func (drift *Drift) isAll() bool {
	if drift.namespace == "default" {
		return !(drift.IsDefaultNamespace)
	}

	if len(drift.namespace) != 0 {
		return false
	}

	return true
}
