package deviation

import (
	"github.com/thoas/go-funk"
)

const (
	Failed  = "FAILED"
	Success = "SUCCESS"
)

type DriftedReleases struct {
	Namespace  string      `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Deviations []Deviation `json:"deviations,omitempty" yaml:"deviations,omitempty"`
}

type Deviation struct {
	Deviations   string `json:"deviations,omitempty" yaml:"deviations,omitempty"`
	HasDrift     bool   `json:"has_drift,omitempty" yaml:"has_drift,omitempty"`
	Kind         string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Resource     string `json:"resource,omitempty" yaml:"resource,omitempty"`
	TemplatePath string `json:"template_path,omitempty" yaml:"template_path,omitempty"`
	ManifestPath string `json:"manifest_path,omitempty" yaml:"manifest_path,omitempty"`
}

type Deviations []Deviation

func (dvn *Deviation) Drifted() string {
	if dvn.HasDrift {
		return "YES"
	}

	return "NO"
}

func (dvn Deviations) GetDriftAsMap(chart, release, time string) map[string]interface{} {
	return map[string]interface{}{
		"drifts":       dvn,
		"total_drifts": dvn.Count(),
		"time":         time,
		"release":      release,
		"chart":        chart,
		"status":       dvn.Status(),
	}
}

func (dvn Deviations) Status() string {
	hasDrift := funk.Contains(dvn, func(dft Deviation) bool {
		return dft.HasDrift
	})

	if hasDrift {
		return Failed
	}

	return Success
}

func (dvn Deviations) Count() int {
	var count int

	for _, dft := range dvn {
		if dft.HasDrift {
			count++
		}
	}

	return count
}
