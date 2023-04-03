package pkg

import (
	"fmt"

	"github.com/nikhilsbhat/helm-drift/pkg/command"
	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
)

func (drift *Drift) Diff(driftedReleases deviation.DriftedReleases) (deviation.DriftedReleases, error) {
	diffs := make([]deviation.Deviation, 0)

	for _, devn := range driftedReleases.Deviations {
		manifestPath := devn.ManifestPath

		drift.log.Debugf("calculating diff for %s", manifestPath)

		arguments := []string{fmt.Sprintf("-f=%s", manifestPath)}

		cmd := command.NewCommand("kubectl", drift.log)

		cmd.SetKubeCmd(driftedReleases.Namespace, arguments...)

		dvn, err := cmd.RunKubeCmd(devn)
		if err != nil {
			return driftedReleases, err
		}

		diffs = append(diffs, dvn)
	}

	driftedReleases.Deviations = diffs

	drift.log.Debug("ran diffs for all manifests successfully")

	return driftedReleases, nil
}
