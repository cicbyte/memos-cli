package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/cicbyte/memos-cli/internal/common"
	authlogic "github.com/cicbyte/memos-cli/internal/logic/auth"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	authName     string
	authURL      string
	authUsername string
	authPassword string
	authToken    string
)

func getLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "登录 Memos 服务器",
		Long: `登录 Memos 服务器。

支持用户名/密码或 Token 两种方式登录。

示例:
  memos-cli auth login
  memos-cli auth login --url=https://memos.example.com --username=myuser
  memos-cli auth login --url=https://memos.example.com --token=my-token
  memos-cli auth login --name=my-server --url=https://memos.example.com`,
		Run: runLogin,
	}

	cmd.Flags().StringVarP(&authName, "name", "n", "", "服务器名称（保存配置用）")
	cmd.Flags().StringVarP(&authURL, "url", "u", "", "服务器 URL")
	cmd.Flags().StringVarP(&authUsername, "username", "U", "", "用户名")
	cmd.Flags().StringVarP(&authPassword, "password", "p", "", "密码")
	cmd.Flags().StringVarP(&authToken, "token", "t", "", "认证 Token")

	return cmd
}

func runLogin(cmd *cobra.Command, args []string) {
	url := authURL

	if url == "" {
		if server := common.GetAppConfig().GetDefaultServer(); server != nil {
			url = server.URL
			fmt.Printf("使用默认服务器: %s (%s)\n", server.Name, server.URL)
		} else {
			fmt.Print("请输入服务器 URL: ")
			reader := bufio.NewReader(os.Stdin)
			url, _ = reader.ReadString('\n')
			url = strings.TrimSpace(url)
		}
	}

	if err := validateLoginParams(url, authToken, authUsername, authPassword); err != nil {
		fmt.Printf("❌ %v\n", err)
		cmd.Help()
		return
	}

	username := authUsername
	password := authPassword

	if username == "" {
		fmt.Print("请输入用户名: ")
		reader := bufio.NewReader(os.Stdin)
		username, _ = reader.ReadString('\n')
		username = strings.TrimSpace(username)
	}

	if password == "" {
		fmt.Print("请输入密码: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			fmt.Printf("读取密码失败: %v\n", err)
			os.Exit(1)
		}
		password = string(passwordBytes)
	}

	config := &authlogic.LoginConfig{
		Name:     authName,
		URL:      url,
		Username: username,
		Password: password,
		Token:    authToken,
	}

	processor := authlogic.NewLoginProcessor(config, common.GetAppConfig())
	result, err := processor.Execute(context.Background())
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✓ 登录成功!")
	if result.User != nil {
		fmt.Printf("  登录用户: %s\n", result.User.Username)
		if result.User.Nickname != "" {
			fmt.Printf("  昵称: %s\n", result.User.Nickname)
		}
	}
	fmt.Printf("  服务器: %s\n", url)
	fmt.Printf("  配置已保存为: %s\n", result.ServerName)
}
