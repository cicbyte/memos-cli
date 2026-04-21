package auth

import (
	"context"
	"fmt"

	"github.com/cicbyte/memos-cli/internal/common"
	authlogic "github.com/cicbyte/memos-cli/internal/logic/auth"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func getStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "查看认证状态",
		Long: `查看当前认证状态。

显示当前用户信息和服务器连接状态。

示例:
  memos-cli auth status`,
			Run: runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) {
	processor := authlogic.NewStatusProcessor(common.GetAppConfig())
	result, err := processor.Execute(context.Background())
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Render("认证状态"))
	fmt.Println()

	if result.ServerName == "" {
		fmt.Println("  未配置服务器。")
		fmt.Println("\n  请先运行 'memos-cli auth login' 登录。")
		return
	}

	fmt.Printf("  服务器: %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(result.ServerName))
	fmt.Printf("  地址: %s\n", result.ServerURL)

	if !result.Authenticated {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("  状态: 未认证"))
		fmt.Println("\n  请运行 'memos-cli auth login' 进行认证。")
		return
	}

	if result.AuthError != nil {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("  状态: 认证失败"))
		fmt.Printf("  错误: %v\n", result.AuthError)
		fmt.Println("\n  请运行 'memos-cli auth login' 重新认证。")
		return
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("  状态: 已认证 ✓"))
	fmt.Println()
	if result.User != nil {
		fmt.Printf("  用户名: %s\n", result.User.Username)
		if result.User.Nickname != "" {
			fmt.Printf("  昵称: %s\n", result.User.Nickname)
		}
		if result.User.Email != "" {
			fmt.Printf("  邮箱: %s\n", result.User.Email)
		}
		fmt.Printf("  角色: %s\n", result.User.Role)
	}
}
