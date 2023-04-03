package command

import (
	"context"
	"os/exec"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	log "github.com/sirupsen/logrus"
)

const (
	HelmContext   = "HELM_KUBECONTEXT"
	HelmNamespace = "HELM_NAMESPACE"
	KubeConfig    = "KUBECONFIG"
)

type Exec interface {
	SetKubeCmd(namespace string, args ...string)
	RunKubeCmd(deviation deviation.Deviation) (deviation.Deviation, error)
}

type command struct {
	baseCmd *exec.Cmd
	log     *log.Logger
}

func NewCommand(cmd string, logger *log.Logger) Exec {
	commandClient := command{
		baseCmd: exec.CommandContext(context.Background(), cmd),
		log:     logger,
	}

	return &commandClient
}
