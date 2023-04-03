package pkg

import (
	"fmt"

	"github.com/nikhilsbhat/helm-drift/pkg/command"
	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
)

func (drift *Drift) Diff(driftedRelease deviation.DriftedRelease) (deviation.DriftedRelease, error) {
	diffs := make([]deviation.Deviation, 0)

	var drifted bool

	for _, dvn := range driftedRelease.Deviations {
		manifestPath := dvn.ManifestPath

		drift.log.Debugf("calculating diff for %s", manifestPath)

		arguments := []string{fmt.Sprintf("-f=%s", manifestPath)}

		cmd := command.NewCommand("kubectl", drift.log)

		cmd.SetKubeCmd(driftedRelease.Namespace, arguments...)

		dft, err := cmd.RunKubeCmd(dvn)
		if err != nil {
			return driftedRelease, err
		}

		if dft.HasDrift {
			drifted = dft.HasDrift
		}

		diffs = append(diffs, dft)
	}

	driftedRelease.Deviations = diffs
	driftedRelease.HasDrift = drifted

	drift.log.Debugf("ran diffs for all manifests for release '%s' successfully", driftedRelease.Release)

	return driftedRelease, nil
}
