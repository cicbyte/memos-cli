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

var memoDeleteForce bool

func getDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "delete <memo-id>",
		Short:  "删除备忘录",
		Aliases: []string{"rm", "remove"},
		Long: `删除备忘录。

此操作不可撤销！

示例:
  memos-cli memo delete 123
  memos-cli memo delete 123 --force`,
		Args: cobra.ExactArgs(1),
		Run:  runDelete,
	}

	cmd.Flags().BoolVarP(&memoDeleteForce, "force", "f", false, "跳过确认直接删除")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) {
	memoID := args[0]

	if err := checkServerConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !memoDeleteForce {
		fmt.Printf("确认删除备忘录 #%s？[y/N]: ", memoID)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			fmt.Println("已取消。")
			return
		}
	}

	cfg := &memologic.DeleteConfig{MemoID: memoID}
	processor := memologic.NewDeleteProcessor(cfg, common.GetAppConfig())

	if err := processor.Execute(context.Background()); err != nil {
		fmt.Printf("❌ 删除备忘录失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 备忘录 #%s 已删除\n", memoID)
}
