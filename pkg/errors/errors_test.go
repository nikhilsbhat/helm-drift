package errors_test

import (
	"errors"
	"testing"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	pkgErr "github.com/nikhilsbhat/helm-drift/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestSimpleErrors(t *testing.T) {
	assert.Equal(t, "pre", (&pkgErr.PreValidationError{Message: "pre"}).Error())
	assert.Equal(t, "drift", (&pkgErr.DriftError{Message: "drift"}).Error())
	assert.Equal(t,
		"failed to get key 'metadata.name' from the manifest 'manifest'",
		(&pkgErr.NotFoundError{Key: "metadata.name", Manifest: "manifest"}).Error(),
	)
}

func TestDiskError(t *testing.T) {
	errs := make(chan error, 3)
	errs <- errors.New("first")
	errs <- nil
	errs <- errors.New("second")
	close(errs)

	message, hasErrors := (&pkgErr.DiskError{Errors: errs}).HasDiskError()

	assert.True(t, hasErrors)
	assert.Contains(t, message, "first")
	assert.Contains(t, message, "second")
}

func TestDiskErrorWithoutErrors(t *testing.T) {
	errs := make(chan error)
	close(errs)

	message, hasErrors := (&pkgErr.DiskError{Errors: errs}).HasDiskError()

	assert.False(t, hasErrors)
	assert.Equal(t, "rendering helm manifests to disk errored: ", message)
}

func TestNotAllError(t *testing.T) {
	err := (&pkgErr.NotAllError{
		ResourceFromDeviations: []*deviation.Deviation{{Resource: "rendered", Kind: "Service"}},
		Manifests: []*deviation.Deviation{
			{Resource: "rendered", Kind: "Service"},
			{Resource: "missing", Kind: "Deployment"},
		},
	}).Error()

	assert.Contains(t, err, "not all manifests were rendered to disk successfully")
	assert.Contains(t, err, "missing")
	assert.NotContains(t, err, `"resource": "rendered"`)
}
