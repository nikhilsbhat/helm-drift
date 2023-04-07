package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nikhilsbhat/helm-drift/pkg/errors"
	"github.com/nikhilsbhat/helm-drift/version"
	"github.com/spf13/cobra"
)

func getRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "drift [command]",
		Short: "Utility that helps in identifying drifts in infrastructure",
		Long:  `Identifies drifts (mostly due to in place edits) in the kubernetes workloads provisioned via helm charts.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Usage(); err != nil {
				return err
			}

			return nil
		},
	}
	rootCommand.SetUsageTemplate(getUsageTemplate())

	return rootCommand
}

func getVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version [flags]",
		Short: "Command to fetch the version of helm-drift installed",
		Long:  `This will help user to find what version of helm-drift plugin he/she installed in her machine.`,
		RunE:  versionConfig,
	}
}

func getRunCommand() *cobra.Command {
	driftCommand := &cobra.Command{
		Use:   "run [RELEASE] [CHART] [flags]",
		Short: "Identifies drifts from a selected chart/release",
		Long:  "Lists all configuration drifts that are part of specified chart/release if exists.",
		Example: `  helm drift run prometheus-standalone path/to/chart/prometheus-standalone -f ~/path/to/override-config.yaml
  helm drift run prometheus-standalone --from-release`,
		Args: minimumArgError,
		RunE: func(cmd *cobra.Command, args []string) error {
			drifts.SetLogger(drifts.LogLevel)
			drifts.SetWriter(os.Stdout)
			cmd.SilenceUsage = true

			drifts.SetRelease(args[0])
			if !drifts.FromRelease {
				drifts.SetChart(args[1])
			}

			if !drifts.SkipValidation {
				if !drifts.ValidatePrerequisite() {
					return &errors.PreValidationError{Message: "validation failed, please install prerequisites to identify drifts"}
				}
			}

			return drifts.GetDrift()
		},
	}

	registerCommonFlags(driftCommand)
	registerDriftFlags(driftCommand)

	return driftCommand
}

func getAllCommand() *cobra.Command {
	driftCommand := &cobra.Command{
		Use:   "all",
		Short: "Identifies drifts from all release from the cluster",
		Long:  "Lists all configuration drifts that are part of various releases present in the cluster.",
		Example: `  helm drift all --kube-context k3d-sample
  helm drift all --kube-context k3d-sample -n sample`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			drifts.SetLogger(drifts.LogLevel)
			drifts.SetWriter(os.Stdout)
			cmd.SilenceUsage = true

			if !drifts.SkipValidation {
				if !drifts.ValidatePrerequisite() {
					return &errors.PreValidationError{Message: "validation failed, please install prerequisites to identify drifts"}
				}
			}

			drifts.All = true

			return drifts.GetAllDrift()
		},
	}

	registerCommonFlags(driftCommand)
	registerDriftAllFlags(driftCommand)

	return driftCommand
}

func versionConfig(cmd *cobra.Command, args []string) error {
	buildInfo, err := json.Marshal(version.GetBuildInfo())
	if err != nil {
		log.Fatalf("fetching version of helm-version failed with: %v", err)
	}

	writer := bufio.NewWriter(os.Stdout)
	versionInfo := fmt.Sprintf("%s \n", strings.Join([]string{"drift version", string(buildInfo)}, ": "))
	_, err = writer.Write([]byte(versionInfo))

	if err != nil {
		log.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			log.Fatalln(err)
		}
	}(writer)

	return nil
}
