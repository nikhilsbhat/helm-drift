package pkg

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/nikhilsbhat/helm-drift/pkg/deviation"
	"github.com/olekukonko/tablewriter"
)

func (drift *Drift) render(drifts []*deviation.DriftedRelease) error {
	drift.write(addNewLine(""))

	release := deviation.DriftedReleases(drifts)

	if drift.json || drift.yaml {
		return drift.renderer.Render(drifts)
	}

	if drift.table {
		drift.toTABLE(drifts)

		return nil
	}

	drift.print(drifts)

	if release.Drifted() && !drift.DisableExitWithError {
		os.Exit(1)
	}

	return nil
}

func (drift *Drift) toTABLE(drifts []*deviation.DriftedRelease) {
	drift.log.Debug("rendering the drifts in table format since --summary is enabled")
	table := drift.tableSchema()

	switch drift.All {
	case true:
		drift.allTable(table, drifts)
	default:
		drift.runTable(table, drifts)
	}

	table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold})

	table.Render()
	drift.write(addNewLine(fmt.Sprintf("Time spent in identifying drift: '%v'\n", drift.timeSpent)))
}

func (drift *Drift) runTable(table *tablewriter.Table, deviations []*deviation.DriftedRelease) bool {
	drifts := deviations[0]

	table.SetHeader([]string{"kind", "name", "drift"})

	for _, dft := range drifts.Deviations {
		tableRow := []string{dft.Kind, dft.Resource, dft.Drifted()}

		if dft.HasDrift {
			switch !drift.NoColor {
			case true:
				table.Rich(tableRow, []tablewriter.Colors{{}, {}, {tablewriter.FgRedColor}})
			default:
				table.Append(tableRow)
			}
		} else {
			switch !drift.NoColor {
			case true:
				table.Rich(tableRow, []tablewriter.Colors{{}, {}, {tablewriter.FgGreenColor}})
			default:
				table.Append(tableRow)
			}
		}
	}

	dvn := deviation.Deviations(drifts.Deviations)
	hasDrift := dvn.Status()
	table.SetFooter([]string{"", "Status", hasDrift})
	table.SetCaption(true, drift.getCaption())

	if !drift.NoColor {
		if dvn.Status() == deviation.Failed {
			table.SetFooterColor(tablewriter.Colors{}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.FgRedColor})
		} else {
			table.SetFooterColor(tablewriter.Colors{}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.FgGreenColor})
		}
	}

	return hasDrift == deviation.Failed
}

func (drift *Drift) allTable(table *tablewriter.Table, deviations []*deviation.DriftedRelease) bool {
	table.SetHeader([]string{"release", "namespace", "drifted"})

	for _, dvn := range deviations {
		tableRow := []string{dvn.Release, dvn.Namespace, dvn.Drifted()}

		if dvn.HasDrift {
			switch !drift.NoColor {
			case true:
				table.Rich(tableRow, []tablewriter.Colors{{}, {}, {tablewriter.FgRedColor}})
			default:
				table.Append(tableRow)
			}
		} else {
			switch !drift.NoColor {
			case true:
				table.Rich(tableRow, []tablewriter.Colors{{}, {}, {tablewriter.FgGreenColor}})
			default:
				table.Append(tableRow)
			}
		}
	}

	dvn := deviation.DriftedReleases(deviations)
	dvnStatus := dvn.Status()

	table.SetFooter([]string{"", "Status", dvnStatus})

	if !drift.NoColor {
		if dvnStatus == deviation.Failed {
			table.SetFooterColor(tablewriter.Colors{}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.FgRedColor})
		} else {
			table.SetFooterColor(tablewriter.Colors{}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.FgGreenColor})
		}
	}

	return dvnStatus == deviation.Failed
}

func (drift *Drift) print(drifts []*deviation.DriftedRelease) {
	if len(drifts) == 0 {
		os.Exit(0)
	}

	drft := drifts[0]
	deviations := deviation.Deviations(drft.Deviations)
	release := deviation.DriftedReleases(drifts)

	for _, dft := range drifts {
		if !dft.HasDrift {
			continue
		}

		drift.write(addNewLine("------------------------------------------------------------------------------------"))
		drift.write(addNewLine(fmt.Sprintf("Release                                : %s", dft.Release)))

		if len(dft.Chart) != 0 {
			drift.write(addNewLine(fmt.Sprintf("Chart                                  : %s", dft.Chart)))
		}

		for _, dvn := range dft.Deviations {
			if dvn.HasDrift {
				drift.write(addNewLine("------------------------------------------------------------------------------------"))
				drift.write(addNewLine(fmt.Sprintf("Identified drifts in: '%s' '%s'", dvn.Kind, dvn.Resource)))
				drift.write(addNewLine("-----------"))
				drift.write(addNewLine(""))
				drift.write(dvn.Deviations)
				drift.write(addNewLine(addNewLine("-----------")))
			}
		}

		drift.write(addNewLine("------------------------------------------------------------------------------------"))
	}

	switch !release.Drifted() {
	case true:
		drift.write(addNewLine("YAY...! NO DRIFTS FOUND"))
	default:
		drift.write(addNewLine("OOPS...! DRIFTS FOUND"))
	}

	drift.write(addNewLine("------------------------------------------------------------------------------------"))
	drift.write(addNewLine(fmt.Sprintf("Total time spent on identifying drifts : %v", drift.timeSpent)))

	if drift.All {
		drift.write(addNewLine(fmt.Sprintf("Total number of drifts found           : %v", deviations.Count())))
		drift.write(addNewLine(fmt.Sprintf("Status                                 : %s", deviations.Status())))
	} else {
		drift.write(addNewLine(fmt.Sprintf("Total number of drifts found           : %v", release.Count())))
		drift.write(addNewLine(fmt.Sprintf("Status                                 : %s", release.Status())))
	}

	drift.write(addNewLine("------------------------------------------------------------------------------------"))
}

func (drift *Drift) write(data string) {
	_, err := drift.writer.Write([]byte(data))
	if err != nil {
		drift.log.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			drift.log.Fatalln(err)
		}
	}(drift.writer)
}

func addNewLine(message string) string {
	return fmt.Sprintf("%s\n", message)
}

//nolint:nosnakecase
func (drift *Drift) tableSchema() *tablewriter.Table {
	table := tablewriter.NewWriter(drift.writer)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	table.SetColMinWidth(1, 1)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetHeaderLine(true)
	table.SetNoWhiteSpace(false)
	table.SetTablePadding("\t")
	table.SetAutoWrapText(true)
	table.SetCenterSeparator("|")
	table.SetRowSeparator("-")
	table.SetAutoMergeCells(false)

	return table
}

func (drift *Drift) getCaption() string {
	return fmt.Sprintf("Namespace: '%s'\nRelease: '%s'", drift.namespace, drift.release)
}

func (drift *Drift) SetOutputFormats() {
	switch strings.ToLower(drift.OutputFormat) {
	case "yaml", "y":
		drift.yaml = true
	case "json", "j":
		drift.json = true
	case "table", "t":
		drift.table = true
	default:
		if len(drift.OutputFormat) != 0 {
			drift.log.Fatalf("helm drift does not support format '%s', switching to default", drift.OutputFormat)
		}
	}
}
