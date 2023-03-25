package pkg

import (
	"bufio"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

func (drift *Drift) render(drifts []Deviation) {
	if drift.Summary {
		drift.toTABLE(drifts)

		return
	}
	drift.print(drifts)
}

func (drift *Drift) toTABLE(drifts []Deviation) {
	drift.log.Debug("rendering the drifts in table format since --summary is enabled")
	table := drift.tableSchema()

	table.SetHeader([]string{"kind", "name", "drift"})
	table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold})

	for _, dft := range drifts {
		tableRow := []string{dft.Kind, dft.Resource, dft.hasDrift()}
		if dft.HasDrift {
			if !drift.NoColor {
				table.Rich(tableRow, []tablewriter.Colors{{}, {}, {tablewriter.FgRedColor}})
			}
		} else {
			if !drift.NoColor {
				table.Rich(tableRow, []tablewriter.Colors{{}, {}, {tablewriter.FgGreenColor}})
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
}

func (drift *Drift) print(drifts []Deviation) {
	for _, dft := range drifts {
		if dft.HasDrift {
			drift.write(addNewLine("------------------------------------------------------------------------------------"))
			drift.write(addNewLine(addNewLine(fmt.Sprintf("Identified drifts in: '%s' '%s'", dft.Kind, dft.Resource))))
			drift.write(addNewLine("-----------"))
			drift.write(dft.Deviations)
			drift.write(addNewLine(addNewLine("-----------")))
		}
	}
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
