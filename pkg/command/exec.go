package command

//go:generate mockgen -destination ../mocks/command/exec.go -package mockCommand -source ./exec.go
import (
	"context"
	"os/exec"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	log "github.com/sirupsen/logrus"
)

// Exec implements methods that set's and run's the kubectl command.
type Exec interface {
	SetKubeDiffCmd(kubeConfig string, kubeContext string, namespace string, args ...string)
	RunKubeDiffCmd(deviation *deviation.Deviation) (*deviation.Deviation, error)
	SetKubeGetCmd(kubeConfig string, kubeContext string, namespace string, args ...string)
	RunKubeCmd(deviation *deviation.Deviation) ([]byte, error)
}

type command struct {
	baseCmd *exec.Cmd
	log     *log.Logger
}

// NewCommand returns new instance of Exec.
func NewCommand(cmd string, logger *log.Logger) Exec {
	commandClient := command{
		baseCmd: exec.CommandContext(context.Background(), cmd),
		log:     logger,
	}

	return &commandClient
}
