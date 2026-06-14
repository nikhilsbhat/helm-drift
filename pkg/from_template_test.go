package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplates(t *testing.T) {
	drift := Drift{Regex: TemplateRegex}
	drift.SetLogger("error")

	templates := drift.getTemplates([]byte(`---
# Source: sample/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sample
---
# Source: sample/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: sample
`))

	assert.Len(t, templates, 2)
	assert.Contains(t, templates[0], "kind: Deployment")
	assert.Contains(t, templates[1], "kind: Service")
}
