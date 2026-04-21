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
	syncFull        bool
	syncForce       bool
	syncNoVectorize bool
	syncDryRun      bool
	syncVerbose     bool
)

var (
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	mutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true).Width(12)
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	loadingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	statBoxStyle = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8"))
)

func runSync(cmd *cobra.Command, args []string) {
	if syncFull && syncForce {
		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("9")).
			Padding(0, 1).
			Render(warnStyle.Render("  警告: 全量同步将删除所有本地数据!"))
		fmt.Println(box)
		fmt.Print("  确认继续? [y/N] ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println(mutedStyle.Render("  已取消"))
			return
		}
	}

	appConfig := common.GetAppConfig()
	if appConfig == nil || len(appConfig.Servers) == 0 {
		fmt.Println(errorStyle.Render("  未配置服务器，请先运行 memos-cli server add"))
		os.Exit(1)
	}

	cfg := &synclogic.SyncConfig{
		FullSync:    syncFull,
		Force:       syncForce,
		NoVectorize: syncNoVectorize,
		Verbose:     syncVerbose,
	}

	processor := synclogic.NewSyncProcessor(cfg, appConfig)
	serverConfig := appConfig.GetDefaultServer()

	fmt.Println(loadingStyle.Render(fmt.Sprintf("  正在同步备忘录... (%s)", serverConfig.URL)))

	if syncFull {
		fmt.Println(warnStyle.Render("  全量同步模式"))
	} else {
		lastSync, _ := processor.GetLastSyncTime()
		if lastSync > 0 {
			fmt.Println(mutedStyle.Render(fmt.Sprintf("  上次同步: %s", time.Unix(lastSync, 0).Format("2006-01-02 15:04:05"))))
		}
	}
	fmt.Println()

	result, err := processor.Execute()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("  同步失败: %v", err)))
		os.Exit(1)
	}

	total := result.Added + result.Updated + result.Skipped

	fmt.Println(titleStyle.Render("  同步完成"))
	fmt.Println(mutedStyle.Render("───────────────────────────────────────"))
	fmt.Println()

	stats := lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render("  新增")+fmt.Sprintf("  %d 条", result.Added),
		labelStyle.Render("  更新")+fmt.Sprintf("  %d 条", result.Updated),
		labelStyle.Render("  删除")+fmt.Sprintf("  %d 条", result.Deleted),
		labelStyle.Render("  跳过")+fmt.Sprintf("  %d 条", result.Skipped),
	)
	if result.Vectorized > 0 {
		stats = lipgloss.JoinVertical(lipgloss.Left,
			stats,
			labelStyle.Render("  向量化")+fmt.Sprintf("  %d 条", result.Vectorized),
		)
	}
	fmt.Println(statBoxStyle.Render(stats))
	fmt.Println()
	fmt.Println(successStyle.Render(fmt.Sprintf("  本地总计 %d 条备忘录 · 耗时 %.2fs", total, result.Duration.Seconds())))

	if syncVerbose {
		fmt.Println(mutedStyle.Render("  向量统计: (详情待实现)"))
	}
}
