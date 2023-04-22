package cmd

import (
	"path/filepath"

	"github.com/nikhilsbhat/helm-drift/pkg"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

// Registers all global flags to utility itself.
func registerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayVar(&drifts.Values, "set", []string{},
		"set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.PersistentFlags().StringArrayVar(&drifts.StringValues, "set-string", []string{},
		"set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.PersistentFlags().StringArrayVar(&drifts.FileValues, "set-file", []string{},
		"set values from respective files specified via the command line "+
			"(can specify multiple or separate values with commas: key1=path1,key2=path2)")
	cmd.PersistentFlags().VarP(&drifts.ValueFiles, "values", "f",
		"specify values in a YAML file (can specify multiple)")
	cmd.PersistentFlags().BoolVarP(&drifts.SkipTests, "skip-tests", "", false,
		"setting this would set '--skip-tests' for helm template command while generating templates")
	cmd.PersistentFlags().StringVarP(&drifts.LogLevel, "log-level", "l", "info",
		"log level for the plugin helm drift (defaults to info)")
	cmd.PersistentFlags().BoolVarP(&drifts.NoColor, "no-color", "", false,
		"enabling this would render summary with no color")
}

// Registers flags to support command run/all.
func registerCommonFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&drifts.Regex, "regex", "", pkg.TemplateRegex,
		"regex used to split helm template rendered")
	cmd.PersistentFlags().StringVarP(&drifts.TempPath, "temp-path", "", filepath.Join(homedir.HomeDir(), ".helm-drift", "templates"),
		"path on disk where the helm templates would be rendered on to (the same would be used be used by 'kubectl diff')")
	cmd.PersistentFlags().BoolVarP(&drifts.SkipValidation, "skip-validation", "", false,
		"enable the flag if prerequisite validation needs to be skipped")
	cmd.PersistentFlags().BoolVarP(&drifts.SkipClean, "skip-cleaning", "", false,
		"enable the flag to skip cleaning the manifests rendered on to disk")
	cmd.PersistentFlags().BoolVarP(&drifts.Summary, "summary", "", false,
		"if enabled, prints a quick summary in table format without printing actual drifts")
	cmd.PersistentFlags().BoolVarP(&drifts.JSON, "json", "j", false,
		"enable the flag to render drifts in json format (disabled by default)")
	cmd.PersistentFlags().BoolVarP(&drifts.YAML, "yaml", "y", false,
		"enable the flag to render drifts in yaml format (disabled by default)")
	cmd.PersistentFlags().BoolVarP(&drifts.ExitWithError, "disable-error-on-drift", "d", false,
		"enabling this would disable exiting with error if drifts were identified (works only when --summary is enabled)")
	cmd.PersistentFlags().BoolVarP(&drifts.Report, "report", "", false,
		"when enabled the summary report would be rendered on to a file (this works only if --yaml or --json is enabled along with summary)")
	cmd.PersistentFlags().StringVarP(&drifts.CustomDiff, "custom-diff", "", "",
		"custom diff command to use instead of default, the command passed here would be set under `KUBECTL_EXTERNAL_DIFF`."+
			"More information can be found here https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#diff")
	cmd.PersistentFlags().StringVarP(&drifts.Name, "name", "", "",
		"name of the kubernetes resource to limit the drift identification")
	cmd.PersistentFlags().StringSliceVarP(&drifts.Kind, "kind", "", nil,
		"kubernetes resource names to limit the drift identification (--kind takes higher precedence over --name)")
	cmd.PersistentFlags().StringSliceVarP(&drifts.SkipKinds, "skip", "", nil,
		"kubernetes resource names to skip the drift identification (ex: --skip Deployments)")
}

// Registers flags specific to command, run.
func registerDriftFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&drifts.FromRelease, "from-release", "", false,
		"enable the flag to identify drifts from a release instead (disabled by default, works with command 'run' not with 'all')")
}

// Registers flags specific to command, all.
func registerDriftAllFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&drifts.IsDefaultNamespace, "is-default-namespace", "", false,
		"set this flag if drifts have to be checked specifically in 'default' namespace")
}
