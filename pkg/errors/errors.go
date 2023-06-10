package errors

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	"github.com/thoas/go-funk"
)

type PreValidationError struct {
	Message string
}

type DriftError struct {
	Message string
}

type NotAllError struct {
	ResourceFromDeviations []deviation.Deviation
	Manifests              []deviation.Deviation
}

type DiskError struct {
	Errors chan error
}

type NotFoundError struct {
	Key      string
	Manifest string
}

func (e *PreValidationError) Error() string {
	return e.Message
}

func (e *DriftError) Error() string {
	return e.Message
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("failed to get key '%s' from the manifest '%s'", e.Key, e.Manifest)
}

func (e *DiskError) HasDiskError() (string, bool) {
	var diskErrors []string

	var hasErrors bool

	for err := range e.Errors {
		if err != nil {
			diskErrors = append(diskErrors, err.Error())
		}
	}

	if len(diskErrors) != 0 {
		hasErrors = true
	}

	return fmt.Sprintf("rendering helm manifests to disk errored: %s", strings.Join(diskErrors, "\n")), hasErrors
}

func (e *NotAllError) Error() string {
	var diffs []deviation.Deviation

	for _, resource := range e.Manifests {
		rs := resource
		if !funk.Contains(e.ResourceFromDeviations, func(dvn deviation.Deviation) bool {
			return (dvn.Resource == rs.Resource) && (dvn.Kind == rs.Kind)
		}) {
			diffs = append(diffs, rs)
		}
	}

	diffJSON, err := json.MarshalIndent(diffs, " ", " ")
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("not all manifests were rendered to disk successfully, manifests failed to render: \n%v", string(diffJSON))
}
