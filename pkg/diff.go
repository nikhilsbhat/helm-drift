package pkg

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (drift *Drift) Diff() (map[string]string, error) {
	templatePath := filepath.Join(drift.TempPath, drift.release)

	drift.log.Debugf("reading rendered manifsets under '%s' for calculating diff", templatePath)
	manifests, err := os.ReadDir(templatePath)
	if err != nil {
		return nil, err
	}

	diffs := make(map[string]string, 0)
	for _, manifest := range manifests {
		manifestPath := filepath.Join(templatePath, manifest.Name())

		drift.log.Debugf("calculating diff for %s", manifestPath)

		arguments := []string{fmt.Sprintf("-f=%s", manifestPath)}
		cmd := drift.kubeCmd(arguments...)

		drift.log.Debugf("envionment variables that would be used: %v", cmd.Environ())

		out, err := cmd.CombinedOutput()
		if err != nil {
			var exerr *exec.ExitError
			if errors.As(err, &exerr) {
				if exerr.ExitCode() != 1 {
					return nil, fmt.Errorf("running kubectl diff errored with exit code: %w ,with message: %s", err, string(out))
				}
			}
		}

		if len(out) != 0 {
			drift.log.Debugf("found diffs for %s", manifest.Name())
			diffs[manifestPath] = string(out)
		}
	}

	drift.log.Debug("ran diffs for all manifests successfully")

	return diffs, nil
}
