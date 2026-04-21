package chat

import "github.com/spf13/cobra"

func GetChatCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat [问题]",
		Short: "与备忘录对话",
		Long: `基于 AI 与你的备忘录对话，自动检索相关内容并生成回答。

默认单轮对话，传入问题直接返回回答。
使用 --interactive 进入多轮对话模式，支持上下文连续提问。

示例:
  memos-cli chat "上周有哪些工作计划？"
  memos-cli chat "关于React的笔记" --tag "编程"
  memos-cli chat --interactive
  memos-cli chat --interactive "先帮我总结最近的工作"`,
		Run: runChat,
	}

	cmd.Flags().StringSliceVar(&chatTag, "tag", []string{}, "按标签过滤")
	cmd.Flags().StringVar(&chatVisibility, "visibility", "", "按可见性过滤 (PUBLIC/PRIVATE/PROTECTED)")
	cmd.Flags().IntVarP(&chatLimit, "limit", "l", 0, "返回结果数量 (默认: 10)")
	cmd.Flags().BoolVar(&chatShowMemos, "show-memos", false, "显示参考备忘录详情")
	cmd.Flags().StringVarP(&chatOutput, "output", "o", "", "输出到文件")
	cmd.Flags().BoolVarP(&chatInteractive, "interactive", "i", false, "多轮对话模式")
	cmd.Flags().StringVarP(&chatMode, "mode", "m", "auto", "检索模式: auto(自动), db(数据库查询), vector(向量搜索)")

	return cmd
}
