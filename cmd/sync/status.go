package sync

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/cicbyte/memos-cli/internal/common"
	synclogic "github.com/cicbyte/memos-cli/internal/logic/sync"
	"github.com/spf13/cobra"
)

var (
	statusTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	statusLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true).Width(12)
	statusMutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	statusErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	statusBoxStyle   = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8"))
)

func getStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "显示同步状态",
		Run:   runStatus,
	}
}

func runStatus(cmd *cobra.Command, args []string) {
	appConfig := common.GetAppConfig()
	if appConfig == nil || len(appConfig.Servers) == 0 {
		fmt.Println(statusErrorStyle.Render("  未配置服务器，请先运行 memos-cli server add"))
		os.Exit(1)
	}

	processor := synclogic.NewStatusProcessor(appConfig)
	result, err := processor.Execute()
	if err != nil {
		fmt.Println(statusErrorStyle.Render(fmt.Sprintf("  获取同步状态失败: %v", err)))
		os.Exit(1)
	}

	var lastSync string
	if result.LastSyncTime > 0 {
		lastSync = time.Unix(result.LastSyncTime, 0).Format("2006-01-02 15:04:05")
	} else {
		lastSync = statusMutedStyle.Render("从未同步")
	}

	info := lipgloss.JoinVertical(lipgloss.Left,
		statusLabelStyle.Render("  服务器")+result.ServerName,
		statusLabelStyle.Render("  最后同步")+lastSync,
		statusLabelStyle.Render("  本地数量")+fmt.Sprintf("%d 条", result.MemoCount),
		statusLabelStyle.Render("  状态")+formatSyncStatus(result),
	)

	fmt.Println(statusTitleStyle.Render("  同步状态"))
	fmt.Println(statusMutedStyle.Render("───────────────────────────────────────"))
	fmt.Println()
	fmt.Println(statusBoxStyle.Render(info))
	fmt.Println()
}

func formatSyncStatus(result *synclogic.StatusResult) string {
	switch result.SyncStatus {
	case "idle":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("同步完成")
	case "syncing":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("同步中...")
	case "error":
		msg := "错误"
		if result.ErrorMsg != "" {
			msg += ": " + result.ErrorMsg
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(msg)
	default:
		return result.SyncStatus
	}
}
