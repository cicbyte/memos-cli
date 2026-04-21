package server

import (
	"fmt"

	"github.com/cicbyte/memos-cli/internal/common"
	configlogic "github.com/cicbyte/memos-cli/internal/logic/config"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Padding(0, 1)
	nameStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	urlStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	defaultStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	tokenStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func getListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "列出所有已配置的服务器",
		Aliases: []string{"ls"},
		Long: `列出所有已配置的 Memos 服务器。

显示服务器名称、URL 和默认标记。

示例:
  memos-cli server list`,
		Run: runList,
	}
}

func runList(cmd *cobra.Command, args []string) {
	processor := configlogic.NewListProcessor(common.GetAppConfig())
	result := processor.Execute()

	if len(result.Servers) == 0 {
		fmt.Println("暂无已配置的服务器。")
		fmt.Println("\n使用 'memos-cli server add' 添加服务器。")
		return
	}

	fmt.Println()
	fmt.Println(headerStyle.Render("已配置的服务器"))
	fmt.Println()

	for i, server := range result.Servers {
		name := nameStyle.Render(server.Name)
		if server.IsDefault {
			name = nameStyle.Render(server.Name) + " " + defaultStyle.Render("(默认)")
		}

		url := urlStyle.Render(server.URL)
		token := tokenStyle.Render("Token: " + server.TokenPreview)

		fmt.Printf("  %d. %s\n", i+1, name)
		fmt.Printf("     URL: %s\n", url)
		fmt.Printf("     %s\n", token)
		fmt.Println()
	}

	if result.LastUsed != "" {
		fmt.Printf("  上次使用: %s\n", nameStyle.Render(result.LastUsed))
	}
}
