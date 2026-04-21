package config

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/cicbyte/memos-cli/internal/common"
	configlogic "github.com/cicbyte/memos-cli/internal/logic/config"
	"github.com/spf13/cobra"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Padding(0, 1)
	keyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Padding(0, 1)
	valueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Padding(0, 1)
	descStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(0, 1)
	maskStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(0, 1)
	emptyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(0, 1)

	tableBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8")).Padding(0, 1)
)

func getListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有配置项及当前值",
		Run:   runList,
	}
}

func runList(cmd *cobra.Command, args []string) {
	appConfig := common.GetAppConfig()
	items := configlogic.AllConfigItems()

	keyCol := 22
	valueCol := 30
	descCol := 36

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		headerStyle.Width(keyCol).Render("KEY"),
		headerStyle.Width(valueCol).Render("VALUE"),
		headerStyle.Render("DESCRIPTION"),
	)

	totalWidth := keyCol + valueCol + descCol + 3
	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Width(totalWidth).Render(strings.Repeat("─", totalWidth))

	rows := make([]string, 0, len(items))
	currentSection := ""
	for _, item := range items {
		if item.Section != currentSection {
			currentSection = item.Section
			sectionRow := lipgloss.NewStyle().
				Foreground(lipgloss.Color("12")).Bold(true).
				Width(totalWidth).
				Render(fmt.Sprintf(" [%s]", currentSection))
			rows = append(rows, sectionRow)
		}

		raw := configlogic.GetConfigValue(appConfig, item.Key)
		var displayVal string
		switch {
		case raw == "":
			displayVal = emptyStyle.Width(valueCol).Render("(未设置)")
		case item.Sensitive:
			displayVal = maskStyle.Width(valueCol).Render("******")
		default:
			displayVal = valueStyle.Width(valueCol).Render(raw)
		}

		desc := descStyle.Width(descCol).Render(item.Desc)

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Width(keyCol).Render(item.Key),
			displayVal,
			desc,
		)
		rows = append(rows, row)
	}

	table := lipgloss.JoinVertical(lipgloss.Left,
		header,
		sep,
		lipgloss.JoinVertical(lipgloss.Left, rows...),
	)

	fmt.Println()
	fmt.Println(tableBorder.Render(table))
	fmt.Println()
}
