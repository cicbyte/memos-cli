package memo

import (
	"context"
	"fmt"
	"os"

	"github.com/cicbyte/memos-cli/internal/common"
	memologic "github.com/cicbyte/memos-cli/internal/logic/memo"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var memoGetRaw bool

func getGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <memo-id>",
		Short: "查看备忘录详情",
		Long: `查看备忘录详情。

示例:
  memos-cli memo get 123
  memos-cli memo get 123 --raw`,
		Args: cobra.ExactArgs(1),
		Run:  runGet,
	}
	cmd.Flags().BoolVarP(&memoGetRaw, "raw", "r", false, "仅输出原始内容")
	return cmd
}

func runGet(cmd *cobra.Command, args []string) {
	memoID := args[0]

	cfg := &memologic.GetConfig{Raw: memoGetRaw}
	processor := memologic.NewGetProcessor(cfg, common.GetAppConfig())

	memo, err := processor.Execute(context.Background(), memoID)
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		os.Exit(1)
	}

	if memoGetRaw {
		fmt.Println(memo.Content)
		return
	}

	fmt.Println()

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	fmt.Printf("%s\n", titleStyle.Render(fmt.Sprintf("备忘录 #%s", memoID)))

	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	if memo.CreateTime != nil {
		fmt.Printf("%s\n", metaStyle.Render("创建时间: "+memo.CreateTime.Format("2006-01-02 15:04:05")))
	}

	if memo.UpdateTime != nil && !memo.UpdateTime.Equal(*memo.CreateTime) {
		fmt.Printf("%s\n", metaStyle.Render("更新时间: "+memo.UpdateTime.Format("2006-01-02 15:04:05")))
	}

	visibilityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	if memo.Visibility == "PRIVATE" {
		visibilityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	}
	fmt.Printf("%s\n", metaStyle.Render("可见性: ")+visibilityStyle.Render(string(memo.Visibility)))

	if memo.Creator != "" {
		fmt.Printf("%s\n", metaStyle.Render("创建者: "+memo.Creator))
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Padding(0, 1).Render(memo.Content))

	if memo.Property != nil && len(memo.Property.Tags) > 0 {
		fmt.Println()
		tagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
		tags := ""
		for _, tag := range memo.Property.Tags {
			tags += "#" + tag + " "
		}
		fmt.Printf("%s\n", tagStyle.Render(tags))
	}

	if len(memo.Resources) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("附件:"))
		for _, res := range memo.Resources {
			fmt.Printf("  - %s (%s)\n", res.Filename, res.Type)
		}
	}

	if len(memo.Reactions) > 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Bold(true).Render("表情回应:"))
		reactions := ""
		for _, r := range memo.Reactions {
			reactions += r.ReactionType + " "
		}
		fmt.Printf("  %s\n", reactions)
	}
}
