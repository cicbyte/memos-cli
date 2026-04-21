package mcp

import (
	"github.com/spf13/cobra"
)

func GetMcpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "MCP Server",
		Long: `MCP (Model Context Protocol) Server，以 stdio 模式运行。

让 AI 客户端（Cherry Studio、Claude Desktop、Cursor 等）能够搜索和操作本地备忘录。

Cherry Studio 配置:
  设置 → 模型服务 → MCP 服务器 → 添加:
    名称: memos
    命令: memos-cli
    参数: mcp

Claude Desktop 配置:
  在 claude_desktop_config.json 中添加:
  {
    "mcpServers": {
      "memos": {
        "command": "memos-cli",
        "args": ["mcp"]
      }
    }
  }`,
		RunE: runMcp,
	}
}
