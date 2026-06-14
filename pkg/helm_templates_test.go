package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const deploymentManifest = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: sample
  namespace: workloads
`

const serviceManifest = `apiVersion: v1
kind: Service
metadata:
  name: sample
`

const hookManifest = `apiVersion: batch/v1
kind: Job
metadata:
  name: hook
  annotations:
    helm.sh/hook: test
`

func TestHelmTemplatesFilters(t *testing.T) {
	drift := Drift{
		Kind:      []string{"Deployment", "Service"},
		SkipKinds: []string{"Service"},
		Name:      "sample",
	}
	drift.SetLogger("error")

	templates := NewHelmTemplates([]string{deploymentManifest, serviceManifest, hookManifest})

	filtered := NewHelmTemplates(templates.FilterByHelmHook(&drift)).FilterByKind(&drift)
	filtered = NewHelmTemplates(filtered).FilterBySkip(&drift)
	filtered = NewHelmTemplates(filtered).FilterByName(&drift)

	assert.Equal(t, []string{deploymentManifest}, filtered)
}

func TestHelmTemplateGet(t *testing.T) {
	drift := Drift{}
	drift.SetLogger("error")

	template, err := NewHelmTemplate(deploymentManifest).Get(drift.log)
	require.NoError(t, err)

	assert.Equal(t, "apps/v1", template.APIVersion)
	assert.Equal(t, "Deployment", template.Kind)
	assert.Equal(t, "sample", template.Resource)
	assert.Equal(t, "workloads", template.NameSpace)
	assert.Equal(t, deploymentManifest, NewHelmTemplate(deploymentManifest).GetTemplate())
}

func TestHelmTemplatesGet(t *testing.T) {
	drift := Drift{}
	drift.SetLogger("error")

	deviations, err := NewHelmTemplates([]string{deploymentManifest, serviceManifest}).Get(drift.log)
	require.NoError(t, err)

	assert.Len(t, deviations, 2)
	assert.Equal(t, "Deployment", deviations[0].Kind)
	assert.Equal(t, "Service", deviations[1].Kind)
}
