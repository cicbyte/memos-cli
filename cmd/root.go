/*
Copyright © 2025 cicbyte
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/cicbyte/memos-cli/cmd/auth"
	"github.com/cicbyte/memos-cli/cmd/chat"
	"github.com/cicbyte/memos-cli/cmd/config"
	"github.com/cicbyte/memos-cli/cmd/mcp"
	"github.com/cicbyte/memos-cli/cmd/memo"
	"github.com/cicbyte/memos-cli/cmd/server"
	"github.com/cicbyte/memos-cli/cmd/sync"
	"github.com/cicbyte/memos-cli/internal/common"
	"github.com/cicbyte/memos-cli/internal/log"
	"github.com/cicbyte/memos-cli/internal/tui"
	"github.com/cicbyte/memos-cli/internal/utils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "memos-cli",
	Short: "Memos 个人笔记命令行工具",
	Long: `memos-cli - Memos 个人笔记应用的命令行工具。

不带参数运行时启动交互式 TUI 界面，也可以使用子命令直接操作。

示例:
  memos-cli              # 启动 TUI 界面
  memos-cli memo list    # 列出备忘录
  memos-cli auth login   # 登录服务器
  memos-cli server add   # 添加服务器配置`,
	Run: func(cmd *cobra.Command, args []string) {
		// 无参数时启动 TUI
		if err := tui.StartTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 初始化应用目录
	if err := utils.InitAppDirs(); err != nil {
		fmt.Printf("初始化目录失败: %v\n", err)
		os.Exit(1)
	}
	// 加载配置(会自动创建默认配置)
	common.SetAppConfig(utils.ConfigInstance.LoadConfig())
	// 初始化日志
	if err := log.Init(utils.ConfigInstance.GetLogPath()); err != nil {
		fmt.Printf("日志初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化数据库连接
	if _, err := utils.GetGormDB(); err != nil {
		log.Error("数据库连接失败",
			zap.String("operation", "db init"),
			zap.Error(err))
		os.Exit(1)
	}
	log.Info("数据库连接成功")

	// 注册命令模块
	rootCmd.AddCommand(auth.GetAuthCommand())
	rootCmd.AddCommand(chat.GetChatCommand())
	rootCmd.AddCommand(config.GetConfigCommand())
	rootCmd.AddCommand(server.GetServerCommand())
	rootCmd.AddCommand(mcp.GetMcpCommand())
	rootCmd.AddCommand(memo.GetMemoCommand())
	rootCmd.AddCommand(sync.GetSyncCommand())
}
