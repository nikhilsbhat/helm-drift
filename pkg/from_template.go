package pkg

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

func (drift *Drift) getChartFromTemplate() ([]byte, error) {
	flags := make([]string, 0)
	for _, value := range drift.Values {
		flags = append(flags, "--set", value)
	}

	for _, stringValue := range drift.StringValues {
		flags = append(flags, "--set-string", stringValue)
	}

	for _, fileValue := range drift.FileValues {
		flags = append(flags, "--set-file", fileValue)
	}

	for _, valueFile := range drift.ValueFiles {
		flags = append(flags, "--values", valueFile)
	}

	if strings.ToLower(drift.LogLevel) == logrus.DebugLevel.String() {
		flags = append(flags, "--debug")
	}

	if drift.SkipTests {
		flags = append(flags, "--skip-tests")
	}

	args := []string{"template", drift.release, drift.chart}
	args = append(args, flags...)

	drift.log.Debugf("rendering helm chart with following commands/flags '%s'", strings.Join(args, ", "))

	cmd := exec.Command(os.Getenv("HELM_BIN"), args...) //nolint:gosec
	output, err := cmd.Output()

	var exitErr *exec.ExitError

	if errors.As(err, &exitErr) {
		drift.log.Errorf("rendering template for release: '%s' errored with %v", drift.release, err)

		return nil, fmt.Errorf("%w: %s", exitErr, exitErr.Stderr)
	}

	var pathErr *fs.PathError

	if errors.As(err, &pathErr) {
		drift.log.Error("locating helm cli errored with", err)

		return nil, fmt.Errorf("%w: %s", pathErr, pathErr.Path)
	}

	return output, nil
}

func (drift *Drift) getTemplates(template []byte) []string {
	drift.log.Debugf("splitting helm manifests with regex pattern: '%s'", drift.Regex)
	temp := regexp.MustCompile(drift.Regex)
	kinds := temp.Split(string(template), -1)
	// Removing empty string at the beginning as splitting string always adds it in front.
	kinds = kinds[1:]

	return kinds
}
