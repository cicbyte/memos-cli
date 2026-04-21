package server

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cicbyte/memos-cli/internal/common"
	configlogic "github.com/cicbyte/memos-cli/internal/logic/config"
	"github.com/spf13/cobra"
)

var (
	serverRemoveForce bool
)

func getRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove [server-name]",
		Short:   "删除服务器配置",
		Aliases: []string{"rm", "delete"},
		Long: `删除一个已配置的 Memos 服务器。

示例:
  memos-cli server remove my-server
  memos-cli server remove my-server --force`,
		Args: cobra.ExactArgs(1),
		Run:  runRemove,
	}

	cmd.Flags().BoolVarP(&serverRemoveForce, "force", "f", false, "跳过确认直接删除")

	return cmd
}

func runRemove(cmd *cobra.Command, args []string) {
	name := args[0]

	if !serverRemoveForce {
		fmt.Printf("确认删除服务器 '%s'? [y/N]: ", name)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			fmt.Println("已取消。")
			return
		}
	}

	cfg := &configlogic.RemoveConfig{Name: name}
	processor := configlogic.NewRemoveProcessor(cfg, common.GetAppConfig())

	serverName, err := processor.Execute()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 服务器 '%s' 已删除\n", serverName)
}
