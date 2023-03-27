package pkg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

func (drift *Drift) render(drifts []Deviation) error {
	if drift.Summary {
		if drift.JSON {
			return drift.toJSON(drifts)
		}

		if drift.YAML {
			return drift.toYAML(drifts)
		}

		drift.toTABLE(drifts)

		return nil
	}
	drift.print(drifts)

	return nil
}

func (drift *Drift) toTABLE(drifts []Deviation) {
	drift.log.Debug("rendering the drifts in table format since --summary is enabled")
	table := drift.tableSchema()

	var hasDrift bool
	table.SetHeader([]string{"kind", "name", "drift"})
	table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold})

	for _, dft := range drifts {
		tableRow := []string{dft.Kind, dft.Resource, dft.hasDrift()}
		if dft.HasDrift {
			hasDrift = true
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

	table.SetFooter([]string{"", "Status", drift.status(drifts)})
	table.SetCaption(true, drift.getCaption())

	if !drift.NoColor {
		if drift.status(drifts) == Failed {
			table.SetFooterColor(tablewriter.Colors{}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.FgRedColor})
		} else {
			table.SetFooterColor(tablewriter.Colors{}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.FgGreenColor})
		}
	}

	table.Render()
	drift.write(addNewLine(fmt.Sprintf("Time spent in identifying drift: '%v'\n", drift.timeSpent)))

	if hasDrift && !drift.ExitWithError {
		os.Exit(1)
	}
}

func (drift *Drift) toYAML(drifts []Deviation) error {
	drift.log.Debug("rendering the images in yaml format since --yaml is enabled")

	driftMap := drift.getDriftMap(drifts)

	kindYAML, err := yaml.Marshal(driftMap)
	if err != nil {
		return err
	}

	yamlString := strings.Join([]string{"---", string(kindYAML)}, "\n")

	_, err = drift.writer.Write([]byte(yamlString))
	if err != nil {
		drift.log.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			drift.log.Fatalln(err)
		}
	}(drift.writer)

	return drift.generateReport(kindYAML, "yaml")
}

func (drift *Drift) toJSON(drifts []Deviation) error {
	drift.log.Debug("rendering the images in json format since --json is enabled")

	driftMap := drift.getDriftMap(drifts)

	kindJSON, err := json.MarshalIndent(driftMap, " ", " ")
	if err != nil {
		return err
	}

	kindJSON = append(kindJSON, []byte("\n")...)

	_, err = drift.writer.Write(kindJSON)
	if err != nil {
		drift.log.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			drift.log.Fatalln(err)
		}
	}(drift.writer)

	return drift.generateReport(kindJSON, "json")
}

func (drift *Drift) print(drifts []Deviation) {
	var hasDrift bool
	for _, dft := range drifts {
		if dft.HasDrift {
			hasDrift = true
			drift.write(addNewLine("------------------------------------------------------------------------------------"))
			drift.write(addNewLine(addNewLine(fmt.Sprintf("Identified drifts in: '%s' '%s'", dft.Kind, dft.Resource))))
			drift.write(addNewLine("-----------"))
			drift.write(dft.Deviations)
			drift.write(addNewLine(addNewLine("-----------")))
		}
	}
	if !hasDrift {
		drift.write(addNewLine("YAY...! NO DRIFTS FOUND"))
	}
	drift.write(addNewLine(fmt.Sprintf("Release                                : %s\nChart                                  : %s", drift.release, drift.chart)))
	drift.write(addNewLine(fmt.Sprintf("Total time spent on identifying drifts : %v", drift.timeSpent)))
	drift.write(addNewLine(fmt.Sprintf("Total number of drifts found           : %v", drift.driftCount(drifts))))
	drift.write(addNewLine(fmt.Sprintf("Status                                 : %s", drift.status(drifts))))
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
	table := tablewriter.NewWriter(os.Stdout)
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

func (drift *Drift) generateReport(data []byte, fileType string) error {
	if !drift.Report {
		drift.log.Debug("--report was not enabled, not generating summary report")

		return nil
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	reportName := filepath.Join(pwd, fmt.Sprintf("helm_drift_%s.%s", drift.release, fileType))

	drift.log.Debugf("generating summary report as '%s' since --report is enabled", reportName)

	return os.WriteFile(reportName, data, manifestFilePermission)
}
