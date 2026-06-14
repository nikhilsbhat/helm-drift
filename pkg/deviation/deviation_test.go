package deviation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDriftedReleaseStatus(t *testing.T) {
	releases := DriftedReleases{
		{Release: "clean"},
		{Release: "drifted", HasDrift: true},
	}

	assert.Equal(t, Yes, releases[1].Drifted())
	assert.Equal(t, No, releases[0].Drifted())
	assert.Equal(t, Failed, releases.Status())
	assert.Equal(t, 1, releases.Count())
	assert.True(t, releases.Drifted())
}

func TestDeviationStatus(t *testing.T) {
	deviations := Deviations{
		{Resource: "clean"},
		{Resource: "drifted", HasDrift: true},
	}

	assert.Equal(t, Yes, deviations[1].Drifted())
	assert.Equal(t, No, deviations[0].Drifted())
	assert.Equal(t, Failed, deviations.Status())
	assert.Equal(t, 1, deviations.Count())

	driftMap := deviations.GetDriftAsMap("chart", "release", "1s")
	assert.Equal(t, &deviations, driftMap["drifts"])
	assert.Equal(t, 1, driftMap["total_drifts"])
	assert.Equal(t, "1s", driftMap["time"])
	assert.Equal(t, "release", driftMap["release"])
	assert.Equal(t, "chart", driftMap["chart"])
	assert.Equal(t, Failed, driftMap["status"])
}

func TestSuccessfulStatuses(t *testing.T) {
	releases := DriftedReleases{{Release: "clean"}}
	deviations := Deviations{{Resource: "clean"}}

	assert.Equal(t, Success, releases.Status())
	assert.False(t, releases.Drifted())
	assert.Equal(t, Success, deviations.Status())
	assert.Equal(t, 0, deviations.Count())
}
