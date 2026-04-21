package memo

import "github.com/spf13/cobra"

func GetMemoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memo",
		Short: "管理备忘录",
		Long: `管理备忘录 - 创建、列出、查看、编辑和删除备忘录。

示例:
  memos-cli memo list
  memos-cli memo create --content="Hello, world!"
  memos-cli memo get <memo-id>
  memos-cli memo update <memo-id> --content="Updated content"
  memos-cli memo delete <memo-id>`,
	}
	cmd.AddCommand(getListCommand())
	cmd.AddCommand(getGetCommand())
	cmd.AddCommand(getCreateCommand())
	cmd.AddCommand(getUpdateCommand())
	cmd.AddCommand(getDeleteCommand())
	cmd.AddCommand(getStatsCommand())
	return cmd
}
