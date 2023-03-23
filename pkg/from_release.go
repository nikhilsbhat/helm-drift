package pkg

import (
	"fmt"
	"log"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// getChartFromRelease should get the manifest from the selected release.
func (drift *Drift) getChartFromRelease() ([]byte, error) {
	settings := cli.New()

	drift.log.Debug(fmt.Sprintf("fetching chart manifest for release '%s' from kube cluster", drift.release))

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		drift.log.Error("oops initialising helm client errored with", err)

		return nil, err
	}

	client := action.NewGet(actionConfig)
	release, err := client.Run(drift.release)
	if err != nil {
		return nil, err
	}

	drift.log.Debug(fmt.Sprintf("chart manifest for release '%s' was successfully retrieved from kube cluster", drift.release))

	return []byte(release.Manifest), nil
}
