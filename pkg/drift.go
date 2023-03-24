package pkg

import (
	"bufio"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

const (
	// TemplateRegex is the default regex, that is used to split one big helm template to multiple templates.
	// Splitting templates eases the task of  identifying Kubernetes objects.
	TemplateRegex = `---\n# Source:\s.*.`
)

// Drift represents GetDrift.
type Drift struct {
	Values         []string
	StringValues   []string
	FileValues     []string
	ValueFiles     ValueFiles
	SkipTests      bool
	SkipValidation bool
	SkipClean      bool
	Regex          string
	LogLevel       string
	FromRelease    bool
	TempPath       string
	release        string
	chart          string
	log            *logrus.Logger
	writer         *bufio.Writer
}

func (drift *Drift) SetRelease(release string) {
	drift.release = release
}

func (drift *Drift) SetChart(chart string) {
	drift.chart = chart
}

func (drift *Drift) SetWriter(writer io.Writer) {
	drift.writer = bufio.NewWriter(writer)
}

func (drift *Drift) GetDrift() error {
	if !drift.SkipValidation {
		if !drift.validatePrerequisite() {
			drift.log.Fatalf("validation failed, please install prerequisites to identify drifts")
		}
	}

	drift.log.Debug(
		fmt.Sprintf("got all required values to identify drifts from chart/release '%s' proceeding furter to fetch the same", drift.release),
	)

	chart, err := drift.getChartManifests()
	if err != nil {
		return err
	}

	kubeKindTemplates := drift.getTemplates(chart)

	if err = drift.renderToDisk(kubeKindTemplates); err != nil {
		return err
	}

	defer func(drift *Drift) {
		if err = drift.cleanManifests(); err != nil {
			drift.log.Fatalf("cleaning rendered files failed with: %v", err)
		}
	}(drift)

	out, err := drift.Diff()
	if err != nil {
		return err
	}

	if len(out) == 0 {
		drift.log.Info("no drifts were identified")

		return nil
	}

	for file, diff := range out {
		drift.render(addNewLine("------------------------------------------------------------------------------------"))
		drift.render(addNewLine(addNewLine(fmt.Sprintf("Identified drifts in: '%s'", file))))
		drift.render(addNewLine("-----------"))
		drift.render(diff)
		drift.render(addNewLine(addNewLine("-----------")))
	}

	return nil
}

func (drift *Drift) getChartManifests() ([]byte, error) {
	if drift.FromRelease {
		drift.log.Debug(fmt.Sprintf("from-release is selected, hence fetching manifests for '%s' from helm release", drift.release))

		return drift.getChartFromRelease()
	}

	drift.log.Debug(fmt.Sprintf("fetching manifests for '%s' by rendering helm template locally", drift.release))

	return drift.getChartFromTemplate()
}
