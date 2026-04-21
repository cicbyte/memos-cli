package config

import "github.com/spf13/cobra"

func GetConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "管理应用配置",
		Long: `管理 memos-cli 应用配置（AI、Embedding、日志等参数）。

示例:
  memos-cli config list
  memos-cli config get ai.model
  memos-cli config set ai.model qwen2.5`,
	}
	cmd.AddCommand(getListCommand())
	cmd.AddCommand(getGetCommand())
	cmd.AddCommand(getSetCommand())
	return cmd
}
