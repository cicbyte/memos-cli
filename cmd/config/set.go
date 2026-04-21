package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/cicbyte/memos-cli/internal/common"
	configlogic "github.com/cicbyte/memos-cli/internal/logic/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	setKeyStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	setValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	setErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

func getSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> [value]",
		Short: "设置配置项的值",
		Long: `设置指定配置项的值。

敏感字段（如 api_key）会以不回显方式交互式输入，也可通过 value 参数直接传入。

示例:
  memos-cli config set ai.model qwen2.5
  memos-cli config set ai.temperature 0.7
  memos-cli config set ai.api_key sk-xxx
  memos-cli config set ai.api_key        # 交互式输入（不回显）
  memos-cli config set log.compress true`,
		Args: cobra.RangeArgs(1, 2),
		Run:  runSet,
	}
}

func runSet(cmd *cobra.Command, args []string) {
	key := args[0]

	item := configlogic.FindConfigItem(key)
	if item == nil {
		fmt.Println(setErrorStyle.Render(fmt.Sprintf("  未知配置项: %s", key)))
		fmt.Println(setErrorStyle.Render("  使用 'memos-cli config list' 查看所有配置项"))
		os.Exit(1)
	}

	var value string

	if len(args) >= 2 {
		value = args[1]
	} else if item.Sensitive {
		fmt.Printf("  请输入 %s: ", setKeyStyle.Render(key))
		raw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			fmt.Println(setErrorStyle.Render("  读取输入失败"))
			os.Exit(1)
		}
		value = string(raw)
	} else {
		fmt.Printf("  请输入 %s: ", setKeyStyle.Render(key))
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		value = strings.TrimSpace(line)
	}

	if value == "" {
		fmt.Println(setErrorStyle.Render("  值不能为空"))
		os.Exit(1)
	}

	appConfig := common.GetAppConfig()
	processor := configlogic.NewSetProcessor(appConfig)

	if err := processor.Execute(key, value); err != nil {
		fmt.Println(setErrorStyle.Render(fmt.Sprintf("  %v", err)))
		os.Exit(1)
	}

	fmt.Println(setValueStyle.Render(fmt.Sprintf("  ✓ %s 已更新", key)))
}
