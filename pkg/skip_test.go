package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	helmRelease "helm.sh/helm/v3/pkg/release"
)

func TestFilterRelease(t *testing.T) {
	releases := []*helmRelease.Release{
		{Name: "keep", Namespace: "sample"},
		{Name: "skip", Namespace: "sample"},
		{Name: "skip", Namespace: "other"},
	}

	filtered := resourcesToSkip{{name: "skip", namespace: "sample"}}.filterRelease(releases)

	assert.Equal(t, []*helmRelease.Release{
		{Name: "keep", Namespace: "sample"},
		{Name: "skip", Namespace: "other"},
	}, filtered)
}
