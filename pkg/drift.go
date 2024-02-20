package pkg

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/util/homedir"
)

const (
	// TemplateRegex is the default regex, that is used to split one big helm template to multiple templates.
	// Splitting templates eases the task of  identifying Kubernetes objects.
	TemplateRegex = `---\n# Source:\s.*.`
)

// Drift represents GetDrift.
type Drift struct {
	Values             []string
	StringValues       []string
	FileValues         []string
	ValueFiles         ValueFiles
	SkipTests          bool
	SkipValidation     bool
	SkipClean          bool
	Summary            bool
	Regex              string
	LogLevel           string
	FromRelease        bool
	NoColor            bool
	JSON               bool
	YAML               bool
	ExitWithError      bool
	Report             bool
	TempPath           string
	CustomDiff         string
	All                bool
	IsDefaultNamespace bool
	ConsiderHooks      bool
	New                bool
	Kind               []string
	SkipKinds          []string
	IgnoreHookTypes    []string
	Name               string
	release            string
	chart              string
	namespace          string
	kubeConfig         string
	kubeContext        string
	timeSpent          float64
	log                *logrus.Logger
	writer             *bufio.Writer
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
func (drift *Drift) GetDrift() {
	startTime := time.Now()

	if err := drift.cleanManifests(true); err != nil {
		drift.log.Fatalf("cleaning old rendered files failed with: %v", err)
	}

	drift.log.Debugf("got all required values to identify drifts from chart/release '%s' proceeding furter to fetch the same", drift.release)

	if err := drift.setExternalDiff(); err != nil {
		drift.log.Fatalf("%v", err)
	}

	chart, err := drift.getChartManifests()
	if err != nil {
		drift.log.Fatalf("%v", err)
	}

	kubeKindTemplates := drift.getTemplates(chart)

	renderedManifests, err := drift.renderToDisk(kubeKindTemplates, drift.chart, drift.release, drift.namespace)
	if err != nil {
		drift.log.Fatalf("%v", err)
	}

	defer func(drift *Drift) {
		if err = drift.cleanManifests(false); err != nil {
			drift.log.Fatalf("cleaning rendered files failed with: %v", err)
		}
	}(drift)

	var driftedReleases []deviation.DriftedRelease

	if drift.New {
		_, err = drift.NewDiff(renderedManifests)
		if err != nil {
			drift.log.Fatalf("%v", err)
		}

		return
	}

	out, err := drift.Diff(renderedManifests)
	if err != nil {
		drift.log.Fatalf("%v", err)
	}

	if len(out.Deviations) == 0 {
		drift.log.Info("no drifts were identified")
	} else {
		driftedReleases = append(driftedReleases, out)

		drift.timeSpent = time.Since(startTime).Seconds()

		if err = drift.render(driftedReleases); err != nil {
			drift.log.Fatalf("%v", err)
		}
	}
}

func (drift *Drift) getChartManifests() ([]byte, error) {
	if drift.FromRelease {
		drift.log.Debugf("from-release is selected, hence fetching manifests for '%s' from helm release", drift.release)

		return drift.getChartFromRelease()
	}

	drift.log.Debugf("fetching manifests for '%s' by rendering helm template locally", drift.release)

	return drift.getChartFromTemplate()
}

func (drift *Drift) SetNamespace(namespace string) {
	drift.namespace = namespace
	if len(drift.namespace) == 0 {
		drift.namespace = "default"
	}
}

func (drift *Drift) SetKubeConfig(kubeConfig string) {
	drift.kubeConfig = kubeConfig
	if len(drift.kubeConfig) == 0 {
		drift.kubeConfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
}

func (drift *Drift) SetKubeContext(kubeContext string) {
	drift.kubeContext = kubeContext
}

func (drift *Drift) setExternalDiff() error {
	if len(drift.CustomDiff) == 0 {
		return nil
	}

	return os.Setenv("KUBECTL_EXTERNAL_DIFF", drift.CustomDiff)
}
