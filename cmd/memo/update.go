package memo

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cicbyte/memos-cli/internal/common"
	memologic "github.com/cicbyte/memos-cli/internal/logic/memo"
	"github.com/spf13/cobra"
)

var (
	memoUpdateContent    string
	memoUpdateVisibility string
	memoUpdateArchive    bool
	memoUpdateRestore    bool
	memoUpdatePin        bool
	memoUpdateUnpin      bool
)

func getUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <memo-id>",
		Short: "更新备忘录",
		Long: `更新备忘录。

示例:
  memos-cli memo update 123 --content="New content"
  memos-cli memo update 123 --visibility=private
  memos-cli memo update 123 --archive`,
		Args: cobra.ExactArgs(1),
		Run:  runUpdate,
	}

	cmd.Flags().StringVarP(&memoUpdateContent, "content", "c", "", "新内容（使用 '-' 进入交互输入）")
	cmd.Flags().StringVarP(&memoUpdateVisibility, "visibility", "v", "", "修改可见性 (public/private/protected)")
	cmd.Flags().BoolVar(&memoUpdateArchive, "archive", false, "归档备忘录")
	cmd.Flags().BoolVar(&memoUpdateRestore, "restore", false, "恢复已归档备忘录")
	cmd.Flags().BoolVar(&memoUpdatePin, "pin", false, "置顶备忘录")
	cmd.Flags().BoolVar(&memoUpdateUnpin, "unpin", false, "取消置顶")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) {
	memoID := args[0]
	content := memoUpdateContent

	if content == "" && memoUpdateVisibility == "" && !memoUpdateArchive && !memoUpdateRestore && !memoUpdatePin && !memoUpdateUnpin {
		fmt.Println("❌ 未指定更新内容")
		fmt.Println("\n请使用以下参数:")
		fmt.Println("  --content       更新内容")
		fmt.Println("  --visibility    修改可见性")
		fmt.Println("  --archive       归档备忘录")
		fmt.Println("  --restore       恢复已归档备忘录")
		fmt.Println("  --pin           置顶备忘录")
		fmt.Println("  --unpin         取消置顶")
		os.Exit(1)
	}

	if err := checkServerConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if content == "-" {
		fmt.Println("请输入新内容（Ctrl+D 结束）:")
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		content = strings.Join(lines, "\n")
	}

	cfg := &memologic.UpdateConfig{
		MemoID:     memoID,
		Content:    content,
		Visibility: memoUpdateVisibility,
		Archive:    memoUpdateArchive,
		Restore:    memoUpdateRestore,
		Pin:        memoUpdatePin,
		Unpin:      memoUpdateUnpin,
	}

	processor := memologic.NewUpdateProcessor(cfg, common.GetAppConfig())
	memo, err := processor.Execute(context.Background())
	if err != nil {
		fmt.Printf("❌ 更新备忘录失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 备忘录 #%s 更新成功！\n", memoID)
	if memo.Visibility != "" {
		fmt.Printf("  可见性: %s\n", memo.Visibility)
	}
	if memo.Pinned {
		fmt.Println("  已置顶")
	}
	if memo.RowStatus == "ARCHIVED" {
		fmt.Println("  状态: 已归档")
	}
}
