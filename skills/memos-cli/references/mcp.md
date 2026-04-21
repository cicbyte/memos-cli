# MCP Server

以 stdio 模式运行，让 AI 客户端（Cherry Studio、Claude Desktop、Cursor 等）搜索和操作本地备忘录。

## 启动

```bash
memos-cli mcp
```

## 注册的 Tools

| Tool | 描述 | 参数 |
|------|------|------|
| `memo_search` | 按关键词、标签搜索 | `keywords`、`tags`、`limit`、`start_time`、`end_time` |
| `memo_semantic_search` | 语义搜索（需 embedding 服务） | `query`、`start_time`、`end_time`、`limit` |
| `memo_get` | 获取单条备忘录 | `memo_id` |
| `memo_create` | 创建备忘录 | `content`、`visibility` |
| `memo_stats` | 统计概览 | 无 |

## 客户端配置

### Cherry Studio

设置 -> 模型服务 -> MCP 服务器 -> 添加：
- 名称：`memos`
- 命令：`memos-cli`
- 参数：`mcp`

### Claude Desktop

`claude_desktop_config.json`：
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

`.cursor/mcp.json`：
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

- 已配置服务器并同步过数据
- 语义搜索需要 embedding 服务
- 创建备忘录需要已登录
