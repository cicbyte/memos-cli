package server

import (
	"fmt"
	"os"

	"github.com/cicbyte/memos-cli/internal/common"
	configlogic "github.com/cicbyte/memos-cli/internal/logic/config"
	"github.com/spf13/cobra"
)

func getDefaultCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "default [server-name]",
		Short: "设置默认服务器",
		Long: `设置默认的 Memos 服务器。

未指定服务器时使用默认服务器。

示例:
  memos-cli server default my-server`,
		Args: cobra.ExactArgs(1),
		Run:  runDefault,
	}
}

func runDefault(cmd *cobra.Command, args []string) {
	name := args[0]

	cfg := &configlogic.DefaultConfig{Name: name}
	processor := configlogic.NewDefaultProcessor(cfg, common.GetAppConfig())

	if err := processor.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("\n使用 'memos-cli server list' 查看可用服务器。")
		os.Exit(1)
	}

	fmt.Printf("✓ 默认服务器已设为 '%s'\n", name)
}
