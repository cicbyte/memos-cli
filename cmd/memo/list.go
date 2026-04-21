package memo

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cicbyte/memos-cli/internal/common"
	memologic "github.com/cicbyte/memos-cli/internal/logic/memo"
	"github.com/cicbyte/memos-cli/internal/models"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	memoListLimit      int32
	memoListVisibility string
	memoListTag        string
	memoListArchived   bool
	memoListPage       string
	memoListSearch     string
)

var (
	memoListStyle = lipgloss.NewStyle().Padding(0, 1)

	memoTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))

	memoTimeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	memoContentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

	memoTagStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	memoPrivateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))

	memoPublicStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

func getListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "list",
		Short:  "列出备忘录列表",
		Aliases: []string{"ls"},
		Long: `列出本地数据库中的备忘录。

注意: 请先运行 'memos-cli sync' 同步远程备忘录。

示例:
  memos-cli memo list
  memos-cli memo list --page=2
  memos-cli memo list --page=all
  memos-cli memo list --visibility=public
  memos-cli memo list --tag=work
  memos-cli memo list --search="关键词"
  memos-cli memo list --archived`,
		Run: runList,
	}

	cmd.Flags().Int32VarP(&memoListLimit, "limit", "l", 20, "返回数量上限")
	cmd.Flags().StringVarP(&memoListVisibility, "visibility", "v", "", "按可见性过滤 (public/private/protected)")
	cmd.Flags().StringVarP(&memoListTag, "tag", "t", "", "按标签过滤")
	cmd.Flags().BoolVar(&memoListArchived, "archived", false, "显示已归档备忘录")
	cmd.Flags().StringVarP(&memoListPage, "page", "p", "", "页码 (如: 2, 或 '1,2,3', 或 'all' 获取全部)")
	cmd.Flags().StringVarP(&memoListSearch, "search", "s", "", "文本搜索（模糊匹配内容）")

	return cmd
}

func runList(cmd *cobra.Command, args []string) {
	cfg := &memologic.ListConfig{
		Limit:      memoListLimit,
		Visibility: memoListVisibility,
		Tag:        memoListTag,
		Archived:   memoListArchived,
		Page:       memoListPage,
		Search:     memoListSearch,
	}

	processor := memologic.NewListProcessor(cfg, common.GetAppConfig())
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

	if len(result.Memos) == 0 {
		fmt.Println()
		fmt.Println("未找到匹配的备忘录。")
		fmt.Println()
		return
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("备忘录 (%d/%d)", len(result.Memos), result.FilteredCount)))
	fmt.Println()

	for i := range result.Memos {
		printLocalMemoItem(&result.Memos[i])
	}

	if result.FilteredCount > int64(result.PageSize) {
		totalPages := (result.FilteredCount + int64(result.PageSize) - 1) / int64(result.PageSize)
		if memoListPage == "" {
			fmt.Println()
			fmt.Printf("💡 提示: 共 %d 条，当前第 1/%d 页\n", result.FilteredCount, totalPages)
			fmt.Printf("   使用 --page=2 查看下一页，或 --page=all 查看全部\n")
		} else {
			fmt.Println()
			fmt.Printf("💡 共 %d 条，第 %s/%d 页\n", result.FilteredCount, memoListPage, totalPages)
		}
	}
}

func printLocalMemoItem(memo *models.LocalMemo) {
	visibilityStyle := memoPublicStyle
	if memo.Visibility == "PRIVATE" {
		visibilityStyle = memoPrivateStyle
	}

	uid := memo.UID
	if strings.HasPrefix(uid, "memos/") {
		uid = strings.TrimPrefix(uid, "memos/")
	}

	title := memoTitleStyle.Render("#"+uid) + " " + visibilityStyle.Render(memo.Visibility)

	var timeStr string
	if memo.CreatedTime > 0 {
		timeStr = time.Unix(memo.CreatedTime, 0).Format("2006-01-02 15:04")
	}

	content := memo.Content
	if len(content) > 80 {
		content = content[:77] + "..."
	}
	content = truncateLines(content, 2)

	fmt.Printf("  %s\n", title)
	if timeStr != "" {
		fmt.Printf("  %s\n", memoTimeStyle.Render(timeStr))
	}
	fmt.Printf("  %s\n", memoContentStyle.Render(content))

	tags := extractTagsFromProperty(memo.Property)
	if len(tags) > 0 {
		tagStr := ""
		for _, tag := range tags {
			tagStr += "#" + tag + " "
		}
		fmt.Printf("  %s\n", memoTagStyle.Render(tagStr))
	}

	fmt.Println()
}

func extractTagsFromProperty(propertyJSON string) []string {
	var tags []string
	if strings.Contains(propertyJSON, "tags") {
		parts := strings.Split(propertyJSON, "\"")
		inTags := false
		for i, part := range parts {
			if part == "tags" && i+2 < len(parts) {
				inTags = true
				continue
			}
			if inTags && part != ":" && part != "[" && part != "]" && part != "," && part != "" {
				tags = append(tags, part)
			}
			if part == "]" && inTags {
				break
			}
		}
	}
	return tags
}

func truncateLines(s string, maxLines int) string {
	lines := 0
	result := ""
	for _, c := range s {
		if c == '\n' {
			lines++
			if lines >= maxLines {
				result += "..."
				break
			}
			result += " "
		} else {
			result += string(c)
		}
	}
	return result
}
