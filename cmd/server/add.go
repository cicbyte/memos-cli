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
	serverAddName    string
	serverAddURL     string
	serverAddToken   string
	serverAddDefault bool
)

func getAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "添加新的 Memos 服务器",
		Long: `添加新的 Memos 服务器配置。

需要提供服务器名称、URL 和认证 Token。

示例:
  memos-cli server add
  memos-cli server add --name=my-server --url=https://memos.example.com --token=your-token`,
		Run: runAdd,
	}

	cmd.Flags().StringVarP(&serverAddName, "name", "n", "", "服务器名称")
	cmd.Flags().StringVarP(&serverAddURL, "url", "u", "", "服务器 URL (如 https://memos.example.com)")
	cmd.Flags().StringVarP(&serverAddToken, "token", "t", "", "认证 Token")
	cmd.Flags().BoolVarP(&serverAddDefault, "default", "d", false, "设为默认服务器")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) {
	name := serverAddName
	url := serverAddURL
	token := serverAddToken

	if name == "" {
		fmt.Print("请输入服务器名称: ")
		reader := bufio.NewReader(os.Stdin)
		name, _ = reader.ReadString('\n')
		name = strings.TrimSpace(name)
	}

	if url == "" {
		fmt.Print("请输入服务器 URL (如 https://memos.example.com): ")
		reader := bufio.NewReader(os.Stdin)
		url, _ = reader.ReadString('\n')
		url = strings.TrimSpace(url)
	}

	if token == "" {
		fmt.Print("请输入认证 Token: ")
		reader := bufio.NewReader(os.Stdin)
		token, _ = reader.ReadString('\n')
		token = strings.TrimSpace(token)
	}

	if err := validateAddParams(name, url); err != nil {
		fmt.Printf("❌ %v\n", err)
		cmd.Help()
		return
	}

	cfg := &configlogic.AddConfig{
		Name:    name,
		URL:     url,
		Token:   token,
		Default: serverAddDefault,
	}

	processor := configlogic.NewAddProcessor(cfg, common.GetAppConfig())
	serverName, isDefault, err := processor.Execute()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 服务器 '%s' 添加成功\n", serverName)
	if isDefault {
		fmt.Printf("✓ 已设为默认服务器\n")
	}
}
