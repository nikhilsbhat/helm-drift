package pkg

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

const (
	docLink = "https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/" +
		"#append-home-kube-config-to-your-kubeconfig-environment-variable"
)

func (drift *Drift) ValidatePrerequisite() bool {
	success := true

	if len(strings.Split(drift.kubeConfig, ":")) > 1 {
		drift.log.Info("since drift uses kubectl underneath with arg '--kubeconfig' " +
			"drift cannot run with multiple kubeconfig files set under environment var 'KUBECONFIG', " +
			"like mentioned in below document")
		fmt.Printf("%s\n", docLink)

		success = false
	}

	if goPath := exec.Command("kubectl"); goPath.Err != nil {
		if !errors.Is(goPath.Err, exec.ErrDot) {
			drift.log.Infof("%v", goPath.Err.Error())
			drift.log.Info("helm-drift requires 'kubectl' to identify drifts")

			success = false
		}
	}

	return success
}
