# memos-cli

Memos 个人笔记应用的命令行工具。

## 用法

```bash
memos-cli [flags]
memos-cli [command]
```

## 描述

直接运行 `memos-cli` 不带任何参数时，启动基于 Bubbletea 的交互式 TUI 界面。也可以使用子命令进行直接操作。

## 子命令

| 命令 | 说明 |
|------|------|
| [chat](./chat.md) | 与备忘录 AI 对话（单轮/多轮） |
| [memo](./memo.md) | 管理备忘录（增删改查） |
| [sync](./sync.md) | 同步远程备忘录到本地 |
| [server](./server.md) | 管理 Memos 服务器配置 |
| [config](./config.md) | 管理应用配置（AI/Embedding/日志） |
| [auth](./auth.md) | 认证管理（登录/登出/状态） |
| [mcp](./mcp.md) | 启动 MCP Server |

## 示例

```bash
# 启动 TUI 交互界面
memos-cli

# 列出备忘录
memos-cli memo list

# 同步远程数据
memos-cli sync

# AI 对话
memos-cli chat "我上周的工作计划是什么？"

# 多轮对话
memos-cli chat -i

# 管理服务器
memos-cli server add
memos-cli server list

# 查看配置
memos-cli config list

# 启动 MCP Server
memos-cli mcp
```

## 数据目录

所有数据存储在 `~/.cicbyte/memos-cli/` 下：

```
~/.cicbyte/memos-cli/
├── config/
│   └── config.yaml    # 应用配置
├── db/
│   └── app.db         # SQLite 数据库
└── logs/
    └── app.log        # 日志文件
```
