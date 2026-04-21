---
name: memos-cli
description: memos-cli 使用指南。Memos 个人笔记 CLI 工具，支持备忘录 CRUD、AI 语义搜索问答、MCP Server、服务器同步。在以下场景触发：(1) 用户想通过 CLI 管理 Memos 备忘录（创建/查看/搜索/更新/删除），(2) 用户想用 AI 对话检索备忘录，(3) 用户想同步远程 Memos 服务器数据，(4) 用户想配置 MCP Server 集成到 AI 客户端，(5) 用户想管理服务器连接或应用配置
---

# memos-cli 使用指南

memos-cli 是 Memos 个人笔记应用的命令行工具。

## AI 使用约束

- **禁止启动交互式模式**：`memos-cli`（无参数）会启动 Bubbletea TUI，这是给人类使用的，AI 绝不能执行此命令
- **禁止使用交互式输入命令**：不带参数的 `memos-cli memo create`、`memos-cli auth login`（无 flags）等会进入交互式等待输入，AI 不能执行。必须通过 flags 提供所有参数（如 `--content`、`--token`）
- **禁止修改配置**：`memos-cli config set`、`memos-cli auth login`、`memos-cli server add` 等配置操作必须由人类完成，AI 不能代为执行。如果前置条件未满足，应提示用户自行配置
- **AI 只能执行非交互式只读/写入命令**：如 `memo list`、`memo get`、`memo stats`、`memo create -c "..."`、`memo update <id> -c "..."`、`sync`、`chat "问题"`

## 前置条件（需人类提前配置）

1. 配置服务器并登录：`memos-cli auth login` 或 `memos-cli server add`
2. 同步数据：`memos-cli sync`
3. AI 功能需要额外配置 embedding 和 LLM 服务（见 [配置参考](references/config.md)）

## 命令速查

### AI 可用

| 命令 | 说明 |
|------|------|
| `memos-cli memo list` | 列出备忘录 |
| `memos-cli memo stats` | 备忘录统计概览 |
| `memos-cli memo get <id>` | 查看详情 |
| `memos-cli memo create -c "内容"` | 创建备忘录 |
| `memos-cli memo update <id> -c "新内容"` | 更新备忘录 |
| `memos-cli memo delete <id> -f` | 删除备忘录 |
| `memos-cli chat "问题"` | AI 对话（单轮） |
| `memos-cli sync` | 同步远程数据 |

### AI 禁止使用（需人类操作）

| 命令 | 原因 |
|------|------|
| `memos-cli` | 启动交互式 TUI |
| `memos-cli memo create`（无参数） | 交互式输入 |
| `memos-cli chat -i` | 多轮交互式对话 |
| `memos-cli auth login`（无 flags） | 交互式登录 |
| `memos-cli server add`（无 flags） | 交互式配置 |
| `memos-cli config set` | 修改配置 |

## 典型工作流

### 首次使用

```bash
memos-cli auth login                          # 登录服务器
memos-cli sync                                # 同步数据
memos-cli memo list                           # 查看备忘录
```

### AI 检索备忘录

```bash
memos-cli chat "我上周的工作计划是什么？"        # 单轮对话
memos-cli chat -i "帮我总结最近的工作"          # 多轮对话
memos-cli chat "关于微服务架构的笔记" -m vector # 指定语义搜索
```

### 创建备忘录

```bash
memos-cli memo create -c "内容"                # 参数创建
memos-cli memo create -f note.md               # 从文件创建
```

### MCP Server 集成

在 Cherry Studio / Claude Desktop / Cursor 等客户端中配置，让 AI 直接搜索和操作备忘录。详见 [MCP 参考](references/mcp.md)。

## 模块详细文档

- [备忘录管理](references/memo.md) — CRUD、列表过滤、统计
- [AI 对话](references/chat.md) — 检索模式、多轮对话、流式输出
- [MCP Server](references/mcp.md) — 工具列表、客户端配置
- [同步与认证](references/sync-auth.md) — 数据同步、登录登出
- [配置参考](references/config.md) — 服务器、AI/Embedding、日志配置
