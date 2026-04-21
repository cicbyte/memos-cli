package sync

import "github.com/spf13/cobra"

func GetSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "同步远程备忘录到本地",
		Long: `同步远程 Memos 服务器的备忘录到本地数据库。

支持增量同步和全量同步，可选择自动向量化以启用语义搜索功能。

示例:
  memos-cli sync
  memos-cli sync --full
  memos-cli sync --dry-run
  memos-cli sync status`,
		Run: runSync,
	}

	cmd.Flags().BoolVarP(&syncFull, "full", "f", false, "全量同步（删除本地数据，重新同步）")
	cmd.Flags().BoolVarP(&syncForce, "force", "F", false, "强制执行（需要确认）")
	cmd.Flags().BoolVar(&syncNoVectorize, "no-vectorize", false, "不同步时自动向量化")
	cmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "预览模式（不实际执行）")
	cmd.Flags().BoolVarP(&syncVerbose, "verbose", "v", false, "详细输出")

	cmd.AddCommand(getStatusCommand())
	return cmd
}
