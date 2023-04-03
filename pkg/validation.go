package pkg

import (
	"errors"
	"os/exec"
)

func (drift *Drift) ValidatePrerequisite() bool {
	success := true

	if goPath := exec.Command("go"); goPath.Err != nil {
		if !errors.Is(goPath.Err, exec.ErrDot) {
			drift.log.Infof("%v", goPath.Err.Error())
			drift.log.Info("helm-drift requires 'kubectl' to identify drifts")

			success = false
		}
	}

	return success
}
