package pkg

import (
	"bufio"
	"fmt"
	"io"
	"os"

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
	Summary        bool
	Regex          string
	LogLevel       string
	FromRelease    bool
	NoColor        bool
	TempPath       string
	release        string
	chart          string
	namespace      string
	log            *logrus.Logger
	writer         *bufio.Writer
}

// SetRelease sets release for helm drift.
func (drift *Drift) SetRelease(release string) {
	drift.release = release
}

// SetChart sets chart name for helm drift.
func (drift *Drift) SetChart(chart string) {
	drift.chart = chart
}

// SetWriter sets writer to be used by helm drift.
func (drift *Drift) SetWriter(writer io.Writer) {
	drift.writer = bufio.NewWriter(writer)
}

// GetDrift gets all the drifts that the given release/chart has.
func (drift *Drift) GetDrift() error {
	if err := drift.cleanManifests(true); err != nil {
		drift.log.Fatalf("cleaning old rendered files failed with: %v", err)
	}

	if !drift.SkipValidation {
		if !drift.validatePrerequisite() {
			drift.log.Fatalf("validation failed, please install prerequisites to identify drifts")
		}
	}

	drift.log.Debug(
		fmt.Sprintf("got all required values to identify drifts from chart/release '%s' proceeding furter to fetch the same", drift.release),
	)

	drift.setNameSpace()

	chart, err := drift.getChartManifests()
	if err != nil {
		return err
	}

	kubeKindTemplates := drift.getTemplates(chart)

	deviations, err := drift.renderToDisk(kubeKindTemplates)
	if err != nil {
		return err
	}

	defer func(drift *Drift) {
		if err = drift.cleanManifests(false); err != nil {
			drift.log.Fatalf("cleaning rendered files failed with: %v", err)
		}
	}(drift)

	out, err := drift.Diff(deviations)
	if err != nil {
		return err
	}

	if len(out) == 0 {
		drift.log.Info("no drifts were identified")

		return nil
	}

	drift.render(out)

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

func (drift *Drift) setNameSpace() {
	drift.namespace = os.Getenv(helmNamespace)
}
