package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nikhilsbhat/helm-drift/pkg"
	"github.com/nikhilsbhat/helm-drift/version"
	"github.com/spf13/cobra"
)

var drifts = pkg.Drift{}

const (
	getArgumentCountLocal   = 2
	getArgumentCountRelease = 1
)

type driftCommands struct {
	commands []*cobra.Command
}

// SetDriftCommands helps in gathering all the subcommands so that it can be used while registering it with main command.
func SetDriftCommands() *cobra.Command {
	return getDriftCommands()
}

// Add an entry in below function to register new command.
func getDriftCommands() *cobra.Command {
	command := new(driftCommands)
	command.commands = append(command.commands, getDriftCommand())
	command.commands = append(command.commands, getVersionCommand())

	return command.prepareCommands()
}

func (c *driftCommands) prepareCommands() *cobra.Command {
	rootCmd := getRootCommand()
	for _, cmnd := range c.commands {
		rootCmd.AddCommand(cmnd)
	}
	registerFlags(rootCmd)

	return rootCmd
}

func getDriftCommand() *cobra.Command {
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

			return drifts.GetDrift()
		},
	}

	registerRunFlags(driftCommand)

	return driftCommand
}

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

func minimumArgError(cmd *cobra.Command, args []string) error {
	minArgError := errors.New("[RELEASE] or [CHART] cannot be empty")
	oneOfThemError := errors.New("when '--from-release' is enabled, valid input is [RELEASE] and not both [RELEASE] [CHART]")
	cmd.SilenceUsage = true

	if !drifts.FromRelease {
		if len(args) != getArgumentCountLocal {
			log.Println(minArgError)

			return minArgError
		}

		return nil
	}

	if len(args) > getArgumentCountRelease {
		log.Fatalln(oneOfThemError)
	}

	return nil
}

func getUsageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}{{printf "\n" }}
Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}{{printf "\n" }}
Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{printf "\n"}}
Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}{{printf "\n"}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}{{printf "\n"}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}{{printf "\n"}}
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}
{{if .HasAvailableSubCommands}}{{printf "\n"}}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
{{printf "\n"}}`
}
