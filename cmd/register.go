package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/nikhilsbhat/helm-drift/pkg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	drifts    = pkg.Drift{}
	cliLogger *logrus.Logger
)

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
	command.commands = append(command.commands, getRunCommand())
	command.commands = append(command.commands, getAllCommand())
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

//nolint:goerr113
func validateAndSetArgs(cmd *cobra.Command, args []string) error {
	logger := logrus.New()
	logger.SetLevel(pkg.GetLoglevel(drifts.LogLevel))
	logger.WithField("helm-drift", true)
	logger.SetFormatter(&logrus.JSONFormatter{})
	cliLogger = logger

	minArgError := errors.New("[RELEASE] or [CHART] cannot be empty")
	oneOfThemError := errors.New("when '--from-release' is enabled, valid input is [RELEASE] and not both [RELEASE] [CHART]")
	cmd.SilenceUsage = true

	if drifts.Revision != 0 && !drifts.FromRelease {
		cliLogger.Fatalf("the '--revision' flag can only be used when retrieving images from a release, i.e., when the '--from-release' flag is set")
	}

	drifts.SetRelease(args[0])

	if !drifts.FromRelease {
		if len(args) != getArgumentCountLocal {
			log.Println(minArgError)

			return fmt.Errorf("%w", minArgError)
		}

		drifts.SetChart(args[1])

		return nil
	}

	if len(args) > getArgumentCountRelease {
		log.Fatalln(fmt.Errorf("%w", oneOfThemError))
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
