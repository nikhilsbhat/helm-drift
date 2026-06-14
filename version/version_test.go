package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBuildInfo(t *testing.T) {
	Version = "1.2.3"
	Revision = "abc123"
	Env = "production"
	BuildDate = "2026-06-14"
	GoVersion = "go1.25"
	Platform = "darwin/arm64"

	info := GetBuildInfo()

	assert.Equal(t, "1.2.3", info.Version)
	assert.Equal(t, "abc123", info.Revision)
	assert.Equal(t, "production", info.Environment)
	assert.Equal(t, "2026-06-14", info.BuildDate)
	assert.Equal(t, "go1.25", info.GoVersion)
	assert.Equal(t, "darwin/arm64", info.Platform)
}

func TestGetBuildInfoDefaultsNonProductionEnv(t *testing.T) {
	Env = "local"

	info := GetBuildInfo()

	assert.Equal(t, "alfa", info.Environment)
}
