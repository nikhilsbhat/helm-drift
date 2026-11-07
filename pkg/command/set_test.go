package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetKubeDiffCmd(t *testing.T) {
	cmd := NewCommand("kubectl", logrus.New()).(*command)

	cmd.SetKubeDiffCmd("/tmp/kubeconfig", "kind-kind", "sample", "-f=manifest.yaml")

	assert.Equal(t, []string{
		"kubectl",
		"diff",
		"-f=manifest.yaml",
		"-n=sample",
		"--context=kind-kind",
		"--kubeconfig=/tmp/kubeconfig",
	}, cmd.baseCmd.Args)
}

func TestSetKubeGetCmd(t *testing.T) {
	cmd := NewCommand("kubectl", logrus.New()).(*command)

	cmd.SetKubeGetCmd("", "", "default", "pods")

	assert.Equal(t, []string{"kubectl", "get", "pods", "-n=default"}, cmd.baseCmd.Args)
}

func TestGetContext(t *testing.T) {
	assert.Equal(t, []string{"--context=ctx"}, getContext("", "ctx"))
	assert.Equal(t, []string{"--kubeconfig=/tmp/config"}, getContext("/tmp/config", ""))
	assert.Empty(t, getContext("", ""))
}

func TestKubeCommandsInheritKubeConfig(t *testing.T) {
	paths := strings.Join([]string{filepath.Join(t.TempDir(), "first"), filepath.Join(t.TempDir(), "second"), ""}, string(os.PathListSeparator))
	t.Setenv("KUBECONFIG", paths)

	for _, action := range []string{"diff", "get"} {
		t.Run(action, func(t *testing.T) {
			cmd := NewCommand("kubectl", logrus.New()).(*command)
			cmd.setKubeCmd(action, "", "gpu", "sample", "-f=manifest.yaml")
			assert.Equal(t, []string{"kubectl", action, "-f=manifest.yaml", "-n=sample", "--context=gpu"}, cmd.baseCmd.Args)
			assert.Contains(t, cmd.baseCmd.Env, "KUBECONFIG="+paths)
		})
	}
}
