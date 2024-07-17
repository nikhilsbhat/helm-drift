package pkg

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nikhilsbhat/common/errors"
	"github.com/nikhilsbhat/common/renderer"
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
	ValueFiles           ValueFiles `json:"value_files,omitempty"             yaml:"value_files,omitempty"`
	SkipTests            bool       `json:"skip_tests,omitempty"              yaml:"skip_tests,omitempty"`
	SkipValidation       bool       `json:"skip_validation,omitempty"         yaml:"skip_validation,omitempty"`
	SkipClean            bool       `json:"skip_clean,omitempty"              yaml:"skip_clean,omitempty"`
	FromRelease          bool       `json:"from_release,omitempty"            yaml:"from_release,omitempty"`
	NoColor              bool       `json:"no_color,omitempty"                yaml:"no_color,omitempty"`
	DisableExitWithError bool       `json:"disable_exit_with_error,omitempty" yaml:"disable_exit_with_error,omitempty"`
	All                  bool       `json:"all,omitempty"                     yaml:"all,omitempty"`
	IsDefaultNamespace   bool       `json:"is_default_namespace,omitempty"    yaml:"is_default_namespace,omitempty"`
	ConsiderHooks        bool       `json:"consider_hooks,omitempty"          yaml:"consider_hooks,omitempty"`
	SkipCRDS             bool       `json:"skipCRDS,omitempty"                yaml:"skipCRDS,omitempty"`
	Validate             bool       `json:"validate,omitempty"                yaml:"validate,omitempty"`
	IgnoreHPAChanges     bool       `json:"ignore_hpa_changes,omitempty"      yaml:"ignore_hpa_changes,omitempty"`
	Revision             int        `json:"revision,omitempty"                yaml:"revision,omitempty"`
	Concurrency          int        `json:"concurrency,omitempty"             yaml:"concurrency,omitempty"`
	Kind                 []string   `json:"kind,omitempty"                    yaml:"kind,omitempty"`
	SkipReleases         []string   `json:"skip_releases,omitempty"           yaml:"skip_releases,omitempty"`
	SkipKinds            []string   `json:"skip_kinds,omitempty"              yaml:"skip_kinds,omitempty"`
	IgnoreHookTypes      []string   `json:"ignore_hook_types,omitempty"       yaml:"ignore_hook_types,omitempty"`
	Values               []string   `json:"values,omitempty"                  yaml:"values,omitempty"`
	StringValues         []string   `json:"string_values,omitempty"           yaml:"string_values,omitempty"`
	FileValues           []string   `json:"file_values,omitempty"             yaml:"file_values,omitempty"`
	Version              string     `json:"version,omitempty"                 yaml:"version,omitempty"`
	Regex                string     `json:"regex,omitempty"                   yaml:"regex,omitempty"`
	LogLevel             string     `json:"log_level,omitempty"               yaml:"log_level,omitempty"`
	TempPath             string     `json:"temp_path,omitempty"               yaml:"temp_path,omitempty"`
	CustomDiff           string     `json:"custom_diff,omitempty"             yaml:"custom_diff,omitempty"`
	Name                 string     `json:"name,omitempty"                    yaml:"name,omitempty"`
	OutputFormat         string     `json:"output_format,omitempty"           yaml:"output_format,omitempty"`
	releasesToSkip       []resourcesInfo
	json                 bool
	yaml                 bool
	csv                  bool
	table                bool
	release              string
	chart                string
	namespace            string
	kubeConfig           string
	kubeContext          string
	timeSpent            float64
	log                  *logrus.Logger
	writer               *bufio.Writer
	renderer             renderer.Config
}

type resourcesInfo struct {
	name      string
	namespace string
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

// SetRenderer sets renderer to Images.
func (drift *Drift) SetRenderer() {
	render := renderer.GetRenderer(os.Stdout, drift.log, drift.NoColor, drift.yaml, drift.json, drift.csv, drift.table)
	drift.renderer = render
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

	out, err := drift.Diff(renderedManifests)
	if err != nil {
		drift.log.Fatalf("%v", err)
	}

	if len(out.Deviations) == 0 {
		drift.log.Info("no drifts were identified")
	} else {
		drift.timeSpent = time.Since(startTime).Seconds()

		if err = drift.render([]*deviation.DriftedRelease{out}); err != nil {
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

func (drift *Drift) SetReleasesToSkips() error {
	const resourceLength = 2

	releasesToBeSkipped := make([]resourcesInfo, len(drift.SkipReleases))

	for index, skipRelease := range drift.SkipReleases {
		parsedRelease := strings.SplitN(skipRelease, "=", resourceLength)
		if len(parsedRelease) != resourceLength {
			return &errors.CommonError{Message: fmt.Sprintf("unable to parse release skip '%s'", skipRelease)}
		}

		releasesToBeSkipped[index] = resourcesInfo{name: parsedRelease[0], namespace: parsedRelease[1]}
	}

	drift.releasesToSkip = releasesToBeSkipped

	return nil
}
