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
				switch exerr.ExitCode() {
				case 1:
					deviation.HasDrift = true
					deviation.Deviations = string(out)
					drift.log.Debugf("found diffs for '%s' with name '%s'", deviation.Kind, deviation.Kind)
				default:
					return nil, fmt.Errorf("running kubectl diff errored with exit code: %w ,with message: %s", err, string(out))
				}
			}
		} else {
			drift.log.Debugf("no diffs found for '%s' with name '%s'", deviation.Kind, deviation.Kind)
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

func (drift *Drift) getDriftMap(drifts []Deviation) map[string]interface{} {
	return map[string]interface{}{
		"drifts":       drifts,
		"total_drifts": drift.driftCount(drifts),
		"time":         fmt.Sprintf("%v", drift.timeSpent),
		"release":      drift.release,
		"chart":        drift.chart,
		"status":       drift.status(drifts),
	}
}

func (drift *Drift) driftCount(drifts []Deviation) int {
	var count int
	for _, dft := range drifts {
		if dft.HasDrift {
			count++
		}
	}

	return count
}
