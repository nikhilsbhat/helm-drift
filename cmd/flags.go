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
		"enabling this would render output with no color")
	cmd.PersistentFlags().BoolVarP(&drifts.SkipCRDS, "skip-crds", "", false,
		"setting this would set '--skip-crds' for helm template command while generating templates")
	cmd.PersistentFlags().BoolVarP(&drifts.Validate, "validate", "", false,
		"setting this would set '--validate' for helm template command while generating templates")
	cmd.PersistentFlags().StringVarP(&drifts.Version, "version", "", "",
		"specify a version constraint for the chart version to use, the value passed here would be used to set "+
			"--version for helm template command while generating templates")
	cmd.PersistentFlags().IntVarP(&drifts.Revision, "revision", "", 0,
		"revision of your release from which the drifts to be detected")
	cmd.PersistentFlags().IntVarP(&drifts.Concurrency, "concurrency", "", 1,
		"the value to be set for flag --concurrency of 'kubectl diff'")
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
	cmd.PersistentFlags().StringVarP(&drifts.OutputFormat, "output", "o", "",
		"the format to which the output should be rendered to, it should be one of yaml|json|table, if nothing specified it sets to default")
	cmd.PersistentFlags().BoolVarP(&drifts.DisableExitWithError, "disable-error-on-drift", "d", false,
		"enabling this would disable exiting with error if drifts were identified")
	cmd.PersistentFlags().StringVarP(&drifts.CustomDiff, "custom-diff", "", "",
		"custom diff command to use instead of default, the command passed here would be set under `KUBECTL_EXTERNAL_DIFF`."+
			"More information can be found here https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#diff")
	cmd.PersistentFlags().StringVarP(&drifts.Name, "name", "", "",
		"name of the kubernetes resource to limit the drift identification")
	cmd.PersistentFlags().StringSliceVarP(&drifts.Kind, "kind", "", nil,
		"kubernetes resource names to limit the drift identification (--kind takes higher precedence over --name)")
	cmd.PersistentFlags().StringSliceVarP(&drifts.SkipKinds, "skip", "", nil,
		"kubernetes resource names to skip the drift identification (ex: --skip Deployments)")
	cmd.PersistentFlags().BoolVarP(&drifts.ConsiderHooks, "consider-hooks", "", false,
		"when this is enabled, the flag 'ignore-hooks' holds no value")
	cmd.PersistentFlags().StringSliceVarP(&drifts.IgnoreHookTypes, "ignore-hooks", "", []string{"hook-succeeded", "hook-failed"},
		"list of hooks to ignore while identifying the drifts")
	cmd.PersistentFlags().BoolVarP(&drifts.IgnoreHPAChanges, "ignore-hpa-changes", "", false,
		"when enabled, the drifts caused on workload due to hpa scaling would be ignored")
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
