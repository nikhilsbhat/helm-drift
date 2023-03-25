package pkg

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/thoas/go-funk"
)

type Deviation struct {
	Deviations   string
	HasDrift     bool
	Kind         string
	Resource     string
	TemplatePath string
	ManifestPath string
}

const (
	Failed  = "FAILED"
	Success = "SUCCESS"
)

func (drift *Drift) Diff(deviations []Deviation) ([]Deviation, error) {
	diffs := make([]Deviation, 0)
	for _, deviation := range deviations {
		manifestPath := deviation.ManifestPath

		drift.log.Debugf("calculating diff for %s", manifestPath)

		arguments := []string{fmt.Sprintf("-f=%s", manifestPath)}
		cmd := drift.kubeCmd(arguments...)

		drift.log.Debugf("envionment variables that would be used: %v", cmd.Environ())

		out, err := cmd.CombinedOutput()
		if err != nil {
			var exerr *exec.ExitError
			if errors.As(err, &exerr) {
				if exerr.ExitCode() != 1 {
					return nil, fmt.Errorf("running kubectl diff errored with exit code: %w ,with message: %s", err, string(out))
				}
			}
		}

		if len(out) != 0 {
			drift.log.Debugf("found diffs for '%s' with name '%s'", deviation.Kind, deviation.Kind)
			deviation.HasDrift = true
			deviation.Deviations = string(out)
		}
		diffs = append(diffs, deviation)
	}

	drift.log.Debug("ran diffs for all manifests successfully")

	return diffs, nil
}

func (drift *Drift) status(drifts []Deviation) string {
	hasDrift := funk.Contains(drifts, func(dft Deviation) bool {
		return dft.HasDrift
	})

	if hasDrift {
		return Failed
	}

	return Success
}

func (drift *Deviation) hasDrift() string {
	if drift.HasDrift {
		return "YES"
	}

	return "NO"
}
