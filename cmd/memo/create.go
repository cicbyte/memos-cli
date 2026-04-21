package memo

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cicbyte/memos-cli/internal/common"
	memologic "github.com/cicbyte/memos-cli/internal/logic/memo"
	"github.com/spf13/cobra"
)

var (
	memoCreateContent    string
	memoCreateFile      string
	memoCreateVisibility string
)

func getCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "创建备忘录",
		Aliases: []string{"new", "add"},
		Long: `创建新备忘录。

示例:
  memos-cli memo create --content="Hello, world!"
  memos-cli memo create --content="Secret note" --visibility=private
  memos-cli memo create --file=note.md
  echo "内容" | memos-cli memo create
  cat note.md | memos-cli memo create`,
		Run: runCreate,
	}

	cmd.Flags().StringVarP(&memoCreateContent, "content", "c", "", "备忘录内容")
	cmd.Flags().StringVarP(&memoCreateFile, "file", "f", "", "从文件读取内容")
	cmd.Flags().StringVarP(&memoCreateVisibility, "visibility", "v", "", "可见性 (public/private/protected)，默认: private")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string) {
	content := memoCreateContent

	if memoCreateFile != "" {
		data, err := os.ReadFile(memoCreateFile)
		if err != nil {
			fmt.Printf("❌ 读取文件失败: %v\n", err)
			os.Exit(1)
		}
		content = string(data)
	}

	if content == "" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Printf("❌ 读取管道输入失败: %v\n", err)
				os.Exit(1)
			}
			content = strings.TrimSpace(string(data))
		} else {
			fmt.Println("请输入备忘录内容（Ctrl+D 或空行结束）:")
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string
			for {
				if !scanner.Scan() {
					break
				}
				line := scanner.Text()
				if line == "" && len(lines) > 0 {
					break
				}
				lines = append(lines, line)
			}
			content = strings.Join(lines, "\n")
		}
	}

	if strings.TrimSpace(content) == "" {
		fmt.Println("❌ 内容不能为空")
		os.Exit(1)
	}

	if err := checkServerConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg := &memologic.CreateConfig{
		Content:    content,
		Visibility: memoCreateVisibility,
	}

	processor := memologic.NewCreateProcessor(cfg, common.GetAppConfig())
	result, err := processor.Execute(context.Background())
	if err != nil {
		fmt.Printf("❌ 创建备忘录失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 备忘录创建成功！\n")
	fmt.Printf("  ID: %s\n", result.MemoID)
	fmt.Printf("  可见性: %s\n", result.Visibility)
}
