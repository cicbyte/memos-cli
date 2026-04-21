package config

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/cicbyte/memos-cli/internal/common"
	configlogic "github.com/cicbyte/memos-cli/internal/logic/config"
	"github.com/spf13/cobra"
)

var (
	getShowFlag bool
	getKeyStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	getValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	getMaskStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	getDescStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	getTypeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	getHintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
)

func getGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "查看配置项的值",
		Long: `查看指定配置项的当前值。

敏感字段（如 api_key）默认显示为 ******，使用 --show 查看明文。

示例:
  memos-cli config get ai.model
  memos-cli config get ai.api_key
  memos-cli config get ai.api_key --show`,
		Args: cobra.ExactArgs(1),
		Run:  runGet,
	}
}

func init() {
	getGetCommand().Flags().BoolVar(&getShowFlag, "show", false, "显示敏感字段的明文值")
}

func runGet(cmd *cobra.Command, args []string) {
	appConfig := common.GetAppConfig()
	key := args[0]

	processor := configlogic.NewGetProcessor(appConfig)
	result, err := processor.Execute(key)
	if err != nil {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("  %v", err)))
		os.Exit(1)
	}

	item := result.Item
	value := result.Value

	fmt.Println()
	fmt.Printf("  %s %s\n", getKeyStyle.Render(item.Key), getTypeStyle.Render(fmt.Sprintf("<%s>", item.Type)))
	fmt.Printf("  %s\n", getDescStyle.Render(item.Desc))
	fmt.Println()

	if item.Sensitive && !getShowFlag {
		if value != "" {
			fmt.Printf("  值: %s\n", getMaskStyle.Render("******"))
			fmt.Println(getHintStyle.Render("  使用 --show 查看明文"))
		} else {
			fmt.Printf("  值: %s\n", getHintStyle.Render("(未设置)"))
		}
	} else {
		if value != "" {
			fmt.Printf("  值: %s\n", getValueStyle.Render(value))
		} else {
			fmt.Printf("  值: %s\n", getHintStyle.Render("(未设置)"))
		}
	}
	fmt.Println()
}
