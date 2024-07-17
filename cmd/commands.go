package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nikhilsbhat/helm-drift/pkg/errors"
	"github.com/nikhilsbhat/helm-drift/version"
	"github.com/spf13/cobra"
)

var envSettings *EnvSettings

func getRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "drift [command]",
		Short: "A utility that helps in identifying drifts in infrastructure",
		Long:  `Identifies configuration drifts (mostly due to in-place edits) in the Kubernetes workloads provisioned via Helm charts.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}
	rootCommand.SetUsageTemplate(getUsageTemplate())

	envSettings = envSettings.New()

	return rootCommand
}

func getVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version [flags]",
		Short: "Command to fetch the version of helm-drift installed",
		Long:  `This will help the user find what version of the helm-drift plugin he or she installed in her machine.`,
		RunE:  versionConfig,
	}
}

func getRunCommand() *cobra.Command {
	driftRunCommand := &cobra.Command{
		Use:   "run [RELEASE] [CHART] [flags]",
		Short: "Identifies drifts from a selected chart or release.",
		Long:  "It lists all configuration drifts that are part of the specified chart or release, if one exists.",
		Example: `helm drift run prometheus-standalone path/to/chart/prometheus-standalone -f ~/path/to/override-config.yaml
helm drift run prometheus-standalone --from-release`,
		Args: validateAndSetArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			drifts.SetLogger(drifts.LogLevel)
			drifts.SetWriter(os.Stdout)
			drifts.SetOutputFormats()
			drifts.SetRenderer()
			cmd.SilenceUsage = true

			drifts.SetKubeConfig(envSettings.KubeConfig)
			drifts.SetKubeContext(envSettings.KubeContext)
			drifts.SetNamespace(envSettings.Namespace)

			if !drifts.SkipValidation {
				if !drifts.ValidatePrerequisite() {
					return &errors.PreValidationError{Message: "validation failed, please address the prerequisite errors to identify drifts"}
				}
			}

			drifts.GetDrift()

			return nil
		},
	}

	driftRunCommand.SilenceErrors = true
	registerCommonFlags(driftRunCommand)
	registerDriftFlags(driftRunCommand)

	return driftRunCommand
}

func getAllCommand() *cobra.Command {
	driftAllCommand := &cobra.Command{
		Use:   "all",
		Short: "Identifies drifts from all releases from the cluster.",
		Long: `It lists all configuration drifts that are part of various releases present in the cluster. 
Do note that this is expensive operation since multiple kubectl command would be executed in parallel.`,
		Example: `helm drift all --kube-context k3d-sample
helm drift all --kube-context k3d-sample -n sample`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			drifts.SetLogger(drifts.LogLevel)
			drifts.SetWriter(os.Stdout)
			drifts.SetOutputFormats()
			drifts.SetRenderer()

			if err := drifts.SetReleasesToSkips(); err != nil {
				return err
			}

			drifts.SetKubeConfig(envSettings.KubeConfig)
			drifts.SetKubeContext(envSettings.KubeContext)
			drifts.SetNamespace(envSettings.Namespace)

			if !drifts.SkipValidation {
				if !drifts.ValidatePrerequisite() {
					return &errors.PreValidationError{Message: "validation failed, please address the prerequisite errors to identify drifts"}
				}
			}

			drifts.All = true

			drifts.GetAllDrift()

			return nil
		},
	}

	driftAllCommand.SilenceErrors = true
	registerCommonFlags(driftAllCommand)
	registerDriftAllFlags(driftAllCommand)
	driftAllCommand.PersistentFlags().StringArrayVar(&drifts.SkipReleases, "skip-release", nil,
		"list of helm releases to be skipped for identifying helm drifts, ex: ReleaseName=Namespace | ReleaseName=Namespace")

	return driftAllCommand
}

func versionConfig(_ *cobra.Command, _ []string) error {
	buildInfo, err := json.Marshal(version.GetBuildInfo())
	if err != nil {
		cliLogger.Fatalf("fetching version of helm-version failed with: %v", err)
	}

	writer := bufio.NewWriter(os.Stdout)
	versionInfo := fmt.Sprintf("%s \n", strings.Join([]string{"drift version", string(buildInfo)}, ": "))

	if _, err = writer.WriteString(versionInfo); err != nil {
		cliLogger.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			cliLogger.Fatalln(err)
		}
	}(writer)

	return nil
}
