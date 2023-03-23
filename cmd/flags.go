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
		"set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)") //nolint:lll
	cmd.PersistentFlags().VarP(&drifts.ValueFiles, "values", "f",
		"specify values in a YAML file (can specify multiple)")
	cmd.PersistentFlags().BoolVarP(&drifts.SkipTests, "skip-tests", "", false,
		"setting this would set '--skip-tests' for helm template command while generating templates")
	cmd.PersistentFlags().StringVarP(&drifts.LogLevel, "log-level", "l", "info",
		"log level for the plugin helm drift (defaults to info)")
}

// Registers all flags to command, get.
func registerRunFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&drifts.Regex, "regex", "", pkg.TemplateRegex,
		"regex used to split helm template rendered")
	cmd.PersistentFlags().StringVarP(&drifts.TempPath, "temp-path", "", filepath.Join(homedir.HomeDir(), ".helm-drift", "templates"),
		"path on disk where the helm templates would be rendered on to (the same would be used be used by 'kubectl diff')")
	cmd.PersistentFlags().BoolVarP(&drifts.FromRelease, "from-release", "", false,
		"enable the flag to identify drifts from a release instead (disabled by default)")
	cmd.PersistentFlags().BoolVarP(&drifts.SkipValidation, "skip-validation", "", false,
		"enable the flag if prerequisite validation needs to be skipped")
	cmd.PersistentFlags().BoolVarP(&drifts.SkipClean, "skip-cleaning", "", false,
		"enable the flag to skip cleaning the manifests rendered on to disk")
}
