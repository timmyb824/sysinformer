package sysinformer

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

// PrintSectionHeader prints a section heading in green for better visibility in CLI output
func PrintSectionHeader(header string) {
	green := "\033[32m"
	reset := "\033[0m"
	fmt.Printf("%s%s%s\n", green, header, reset)
}

// RenderTable prints a formatted table with the given headers and rows
func RenderTable(headers []string, data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoFormatHeaders(false)
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
}
