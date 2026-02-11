package sysinformer

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
)

// PrintSectionHeader prints a section heading in green for better visibility in CLI output
func PrintSectionHeader(header string) {
	cyan := "\033[36m"
	bold := "\033[1m"
	reset := "\033[0m"
	fmt.Printf("%s%s%s%s\n", bold, cyan, header, reset)
}

func PrintPanel(title string, subtitle string) {
	bold := "\033[1m"
	reset := "\033[0m"
	gray := "\033[90m"

	content := title
	if subtitle != "" {
		content = title + "\n" + subtitle
	}

	lines := strings.Split(content, "\n")
	maxLen := 0
	for _, l := range lines {
		runeCount := len([]rune(l))
		if runeCount > maxLen {
			maxLen = runeCount
		}
	}

	innerWidth := maxLen + 2
	top := gray + "┌" + strings.Repeat("─", innerWidth) + "┐" + reset
	bot := gray + "└" + strings.Repeat("─", innerWidth) + "┘" + reset

	fmt.Println(top)
	for i, l := range lines {
		runeCount := len([]rune(l))
		pad := innerWidth - (runeCount + 2)
		if pad < 0 {
			pad = 0
		}
		style := ""
		if i == 0 {
			style = bold
		}
		line := gray + "│ " + reset + style + l + reset + strings.Repeat(" ", pad) + " " + gray + "│" + reset
		fmt.Println(line)
	}
	fmt.Println(bot)
}

// RenderTable prints a formatted table with the given headers and rows
func RenderTable(headers []string, data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoFormatHeaders(false)
	table.SetRowLine(false)
	table.SetAutoWrapText(true)
	table.SetReflowDuringAutoWrap(true)
	table.SetRowSeparator("─")
	if len(headers) > 0 {
		colors := make([]tablewriter.Colors, 0, len(headers))
		for range headers {
			colors = append(colors, tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor})
		}
		table.SetHeaderColor(colors...)
	}
	table.SetBorder(true)
	table.SetCenterSeparator("│")
	table.SetColumnSeparator("│")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	applyTableWidths(table, headers)
	table.AppendBulk(data)
	table.Render()
}

func RenderKeyValueTable(titleLeft string, titleRight string, rows [][]string) {
	headers := []string{titleLeft, titleRight}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoFormatHeaders(false)
	table.SetRowLine(false)
	table.SetAutoWrapText(true)
	table.SetReflowDuringAutoWrap(true)
	colors := []tablewriter.Colors{
		{tablewriter.Bold, tablewriter.FgCyanColor},
		{tablewriter.Bold, tablewriter.FgCyanColor},
	}
	table.SetHeaderColor(colors...)
	table.SetBorder(true)
	table.SetCenterSeparator("│")
	table.SetColumnSeparator("│")
	table.SetRowSeparator("─")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	applyKeyValueWidths(table)
	table.AppendBulk(rows)
	table.Render()
}

func applyTableWidths(table *tablewriter.Table, headers []string) {
	w := terminalWidth()
	if w <= 0 {
		return
	}
	cols := len(headers)
	if cols <= 0 {
		return
	}

	// Rough accounting for separators/padding.
	usable := w - (cols * 4) - 8
	if usable < 40 {
		usable = 40
	}
	perCol := usable / cols
	if perCol < 12 {
		perCol = 12
	}
	table.SetColWidth(perCol)
}

func applyKeyValueWidths(table *tablewriter.Table) {
	w := terminalWidth()
	if w <= 0 {
		return
	}
	left := 24
	// Extra margin helps prevent terminal soft-wrapping which makes the right border look dashed.
	right := w - left - 20
	if right < 30 {
		right = 30
	}
	table.SetColMinWidth(0, left)
	table.SetColMinWidth(1, right)
	table.SetColWidth(right)
}

func terminalWidth() int {
	fd := int(os.Stdout.Fd())
	if !term.IsTerminal(fd) {
		return 120
	}
	w, _, err := term.GetSize(fd)
	if err != nil || w <= 0 {
		return 120
	}
	return w
}
