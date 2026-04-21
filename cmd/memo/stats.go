package memo

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cicbyte/memos-cli/internal/common"
	memologic "github.com/cicbyte/memos-cli/internal/logic/memo"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func getStatsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "备忘录总览",
		Long: `查看本地备忘录的统计概览。

显示总数、可见性分布、热门标签和最近备忘录。

示例:
  memos-cli memo stats`,
		Run: runStats,
	}
}

func runStats(cmd *cobra.Command, args []string) {
	processor := memologic.NewStatsProcessor(common.GetAppConfig())
	result, err := processor.Execute()
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}

	if result.TotalCount == 0 {
		fmt.Println()
		fmt.Println("💡 本地数据库为空，请先同步:")
		fmt.Println("   memos-cli sync")
		fmt.Println()
		return
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	tagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("─────────────────────────────────")

	fmt.Println()
	fmt.Println(titleStyle.Render("备忘录总览"))
	fmt.Println(divider)
	fmt.Println()
	fmt.Printf("  总计: %s\n\n", titleStyle.Render(fmt.Sprintf("%d 条", result.TotalCount)))

	// 可见性分布
	if len(result.VisibilityCount) > 0 {
		fmt.Println(labelStyle.Render("  可见性分布:"))
		for vis, count := range result.VisibilityCount {
			fmt.Printf("    %s: %d 条\n", vis, count)
		}
		fmt.Println()
	}

	// 热门标签
	if len(result.TopTags) > 0 {
		fmt.Println(labelStyle.Render("  热门标签 (Top 10):"))
		for _, t := range result.TopTags {
			fmt.Printf("    #%s: %d 条\n", tagStyle.Render(t.Tag), t.Count)
		}
		fmt.Println()
	}

	// 最近备忘录
	if len(result.RecentMemos) > 0 {
		fmt.Println(labelStyle.Render(fmt.Sprintf("  最近 %d 条:", len(result.RecentMemos))))
		for i, m := range result.RecentMemos {
			uid := m.UID
			if strings.HasPrefix(uid, "memos/") {
				uid = strings.TrimPrefix(uid, "memos/")
			}
			preview := m.Content
			if len(preview) > 60 {
				preview = preview[:57] + "..."
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			created := time.Unix(m.CreatedTime, 0).Format("2006-01-02")
			fmt.Printf("    %d. %s (%s) %s\n", i+1, valueStyle.Render(uid), created, valueStyle.Render(preview))
		}
		fmt.Println()
	}
}
