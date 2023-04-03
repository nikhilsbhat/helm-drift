package deviation

import (
	"github.com/thoas/go-funk"
)

//nolint:varnamelen
const (
	Failed  = "FAILED"
	Success = "SUCCESS"
	Yes     = "YES"
	No      = "NO"
)

// DriftedRelease holds drift information of the selected release/chart.
type DriftedRelease struct {
	Chart      string      `json:"chart,omitempty" yaml:"chart,omitempty"`
	Namespace  string      `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Release    string      `json:"release,omitempty" yaml:"release,omitempty"`
	HasDrift   bool        `json:"has_drift,omitempty" yaml:"has_drift,omitempty"`
	Deviations []Deviation `json:"deviations,omitempty" yaml:"deviations,omitempty"`
}

// Deviation holds drift information of all manifests from the selected release/chart.
type Deviation struct {
	Deviations   string `json:"deviations,omitempty" yaml:"deviations,omitempty"`
	HasDrift     bool   `json:"has_drift,omitempty" yaml:"has_drift,omitempty"`
	Kind         string `json:"kind,omitempty" yaml:"kind,omitempty"`
	Resource     string `json:"resource,omitempty" yaml:"resource,omitempty"`
	TemplatePath string `json:"template_path,omitempty" yaml:"template_path,omitempty"`
	ManifestPath string `json:"manifest_path,omitempty" yaml:"manifest_path,omitempty"`
}

type (
	Deviations      []Deviation
	DriftedReleases []DriftedRelease
)

// Drifted returns Yes if the release has Drifted.
func (dvn *DriftedRelease) Drifted() string {
	if dvn.HasDrift {
		return Yes
	}

	return No
}

// Status returns Failed if at least one of the release has Drifted.
func (dvn DriftedReleases) Status() string {
	hasDrift := funk.Contains(dvn, func(dft DriftedRelease) bool {
		return dft.HasDrift
	})

	if hasDrift {
		return Failed
	}

	return Success
}

// Count returns total number of drifted release.
func (dvn DriftedReleases) Count() int {
	var count int

	for _, dft := range dvn {
		if dft.HasDrift {
			count++
		}
	}

	return count
}

// Drifted returns Yes if at least one of the release has Drifted.
func (dvn *DriftedReleases) Drifted() string {
	hasDrift := funk.Contains(*dvn, func(dft DriftedRelease) bool {
		return dft.HasDrift
	})

	if hasDrift {
		return Yes
	}

	return No
}

// Drifted returns Yes if at least one of the manifest from a release has Drifted.
func (dvn *Deviation) Drifted() string {
	if dvn.HasDrift {
		return Yes
	}

	return No
}

// GetDriftAsMap returns the map equivalent of drifted release configuration.
func (dvn *Deviations) GetDriftAsMap(chart, release, time string) map[string]interface{} {
	return map[string]interface{}{
		"drifts":       dvn,
		"total_drifts": dvn.Count(),
		"time":         time,
		"release":      release,
		"chart":        chart,
		"status":       dvn.Status(),
	}
}

// Status returns Failed if at least one of the manifest in release has Drifted.
func (dvn Deviations) Status() string {
	hasDrift := funk.Contains(dvn, func(dft Deviation) bool {
		return dft.HasDrift
	})

	if hasDrift {
		return Failed
	}

	return Success
}

// Count returns total number of drifts in release.
func (dvn Deviations) Count() int {
	var count int

	for _, dft := range dvn {
		if dft.HasDrift {
			count++
		}
	}

	return count
}
