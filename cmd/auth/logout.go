package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cicbyte/memos-cli/internal/common"
	authlogic "github.com/cicbyte/memos-cli/internal/logic/auth"
	"github.com/spf13/cobra"
)

var (
	authLogoutServer string
	authLogoutForce  bool
)

func getLogoutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "登出 Memos 服务器",
		Long: `登出 Memos 服务器，清除已保存的认证令牌。

示例:
  memos-cli auth logout
  memos-cli auth logout --server=my-server`,
		Run: runLogout,
	}

	cmd.Flags().StringVarP(&authLogoutServer, "server", "s", "", "指定要登出的服务器名称")
	cmd.Flags().BoolVarP(&authLogoutForce, "force", "f", false, "跳过确认直接登出")

	return cmd
}

func runLogout(cmd *cobra.Command, args []string) {
	config := &authlogic.LogoutConfig{
		ServerName: authLogoutServer,
	}

	processor := authlogic.NewLogoutProcessor(config, common.GetAppConfig())

	if !authLogoutForce {
		var serverName string
		if authLogoutServer != "" {
			serverName = authLogoutServer
		} else if s := common.GetAppConfig().GetDefaultServer(); s != nil {
			serverName = s.Name
		}
		if serverName != "" {
			fmt.Printf("确认从 '%s' 登出？[y/N]: ", serverName)
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "y" && input != "yes" {
				fmt.Println("已取消。")
				return
			}
		}
	}

	serverName, err := processor.Execute(context.Background())
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		return
	}

	fmt.Printf("✓ 已从 '%s' 登出\n", serverName)
}
