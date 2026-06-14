package pkg

import (
	"bytes"
	"testing"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	"github.com/stretchr/testify/assert"
)

func TestRunTableAndAllTable(t *testing.T) {
	drift := Drift{NoColor: true}
	drift.SetLogger("error")
	drift.SetWriter(new(bytes.Buffer))

	table := drift.tableSchema()
	hasDrift := drift.runTable(table, []*deviation.DriftedRelease{{
		Release: "release",
		Deviations: []*deviation.Deviation{
			{Kind: "Deployment", Resource: "sample", HasDrift: true},
			{Kind: "Service", Resource: "sample"},
		},
	}})

	assert.True(t, hasDrift)

	table = drift.tableSchema()
	hasDrift = drift.allTable(table, []*deviation.DriftedRelease{
		{Release: "clean", Namespace: "sample"},
		{Release: "drifted", Namespace: "sample", HasDrift: true},
	})

	assert.True(t, hasDrift)
}

func TestWriteAndFlush(t *testing.T) {
	buffer := new(bytes.Buffer)
	drift := Drift{}
	drift.SetLogger("error")
	drift.SetWriter(buffer)

	drift.write("hello")
	drift.flush()

	assert.Equal(t, "hello", buffer.String())
}
