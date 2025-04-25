package sysinformer

import (
	"os"
	"github.com/olekukonko/tablewriter"
)

// RenderTable prints a formatted table with the given headers and rows
func RenderTable(headers []string, data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoFormatHeaders(false)
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
}
