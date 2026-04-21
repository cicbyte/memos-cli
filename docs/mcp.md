# mcp

MCP (Model Context Protocol) Server，以 stdio 模式运行。

## 用法

```bash
memos-cli mcp
```

## 描述

启动 stdio 模式的 MCP Server，让 AI 客户端（Cherry Studio、Claude Desktop、Cursor 等）能够搜索和操作本地备忘录。

## 注册的 Tools

| Tool | 描述 | 参数 |
|------|------|------|
| `memo_search` | 按关键词、标签搜索本地备忘录 | `keywords`、`tags`、`limit`、`start_time`、`end_time` |
| `memo_semantic_search` | 语义搜索备忘录（需要配置 embedding 服务） | `query`、`start_time`、`end_time`、`limit` |
| `memo_get` | 根据 ID 获取单条备忘录 | `memo_id` |
| `memo_create` | 创建新备忘录 | `content`、`visibility` |
| `memo_stats` | 获取本地备忘录统计概览 | 无 |

## 客户端配置

### Cherry Studio

设置 → 模型服务 → MCP 服务器 → 添加：
- 名称：`memos`
- 命令：`memos-cli`
- 参数：`mcp`

### Claude Desktop

在 `claude_desktop_config.json` 中添加：

```json
{
  "mcpServers": {
    "memos": {
      "command": "memos-cli",
      "args": ["mcp"]
    }
  }
}
```

### Cursor

在 `.cursor/mcp.json` 中添加：

```json
{
  "mcpServers": {
    "memos": {
      "command": "memos-cli",
      "args": ["mcp"]
    }
  }
}
```

## 前置条件

- 需要 `memos-cli` 已配置好服务器并同步过数据
- 语义搜索需要配置 embedding 服务（`memo_search` 不需要）
- 创建备忘录需要已登录服务器

## 示例

```bash
# 启动 MCP Server
memos-cli mcp

# 手动测试（发送 MCP 初始化请求）
echo '{"jsonrpc":"2.0","method":"initialize","id":1,"params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | memos-cli mcp
```
