package auth

import "github.com/spf13/cobra"

func GetAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "认证管理",
		Long: `认证管理 - 登录、登出和查看认证状态。

示例:
  memos-cli auth login --url=https://memos.example.com
  memos-cli auth status
  memos-cli auth logout`,
	}
	cmd.AddCommand(getLoginCommand())
	cmd.AddCommand(getLogoutCommand())
	cmd.AddCommand(getStatusCommand())
	return cmd
}
