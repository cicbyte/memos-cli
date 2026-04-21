# memos-cli

Memos 个人笔记应用的命令行工具，支持交互式 TUI 和子命令两种使用方式。

## 功能特性

- **TUI 交互界面** — 基于 Bubbletea 的终端 UI，支持浏览、创建、编辑、搜索 memo
- **CLI 子命令** — 通过 Cobra 命令直接操作，适合脚本和自动化
- **服务器管理** — 添加、切换、删除多个 Memos 服务器配置
- **智能同步** — 增量同步远程备忘录到本地 SQLite，自动 MD5 去重
- **AI 对话** — Agent 模式，LLM 自动选择检索方式（数据库查询 / 向量搜索），流式输出，Markdown 渲染，支持单轮和多轮对话
- **MCP Server** — stdio 模式的 MCP Server，让 Cherry Studio、Claude Desktop 等 AI 客户端直接搜索和操作本地备忘录
- **配置管理** — 统一管理 AI、Embedding、日志等应用配置
- **多提供商支持** — AI 支持 OpenAI、Ollama、智谱等 OpenAI 兼容 API

## 安装

### 从 Release 下载

前往 [Releases](https://github.com/cicbyte/memos-cli/releases) 下载对应平台的预编译二进制文件。

### 从源码构建

```bash
git clone https://github.com/cicbyte/memos-cli.git
cd memos-cli
go build -o memos-cli .
```

交叉编译：

```bash
python scripts/build.py --local    # 当前平台
python scripts/build.py             # 全平台（Windows/Linux/macOS）
```

## 快速开始

```bash
# 添加 Memos 服务器配置
memos-cli server add

# 登录服务器
memos-cli auth login

# 同步远程 memo 到本地
memos-cli sync

# 与备忘录对话
memos-cli chat "我上周有哪些工作计划？"

# 启动 TUI 交互界面
memos-cli
```

## 命令一览

| 命令 | 说明 |
|------|------|
| `memos-cli` | 启动 TUI 交互界面 |
| `memos-cli mcp` | 启动 MCP Server |
| `memos-cli chat [问题]` | 与备忘录 AI 对话（默认单轮，`-i` 多轮） |
| `memos-cli memo list` | 列出备忘录列表 |
| `memos-cli memo stats` | 备忘录统计概览 |
| `memos-cli memo get <id>` | 查看 memo 详情 |
| `memos-cli memo create` | 创建 memo（支持管道输入） |
| `memos-cli memo update <id>` | 更新 memo |
| `memos-cli memo delete <id>` | 删除 memo |
| `memos-cli sync` | 同步远程备忘录到本地 |
| `memos-cli sync status` | 查看同步状态 |
| `memos-cli server add` | 添加服务器配置 |
| `memos-cli server list` | 列出服务器配置 |
| `memos-cli server default <name>` | 设置默认服务器 |
| `memos-cli server remove <name>` | 删除服务器配置 |
| `memos-cli config list` | 查看应用配置项 |
| `memos-cli config get <key>` | 查看配置值 |
| `memos-cli config set <key> <value>` | 设置配置项 |
| `memos-cli auth login` | 登录服务器 |
| `memos-cli auth logout` | 登出服务器 |
| `memos-cli auth status` | 查看认证状态 |

## MCP Server

`memos-cli mcp` 以 stdio 模式运行 MCP Server，让 AI 客户端能直接搜索和操作本地备忘录。

注册的 Tools：`memo_search`、`memo_semantic_search`、`memo_get`、`memo_create`、`memo_stats`

**Cherry Studio 配置：**

设置 → 模型服务 → MCP 服务器 → 添加：
- 名称：`memos`
- 命令：`memos-cli`
- 参数：`mcp`

**Claude Desktop 配置：**

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

## AI 配置

通过 `memos-cli config set` 或直接编辑 `~/.cicbyte/memos-cli/config/config.yaml` 配置。

**OpenAI：**

```yaml
ai:
  provider: openai
  base_url: https://api.openai.com/v1
  api_key: sk-xxx
  model: gpt-4o
```

**Ollama（本地，默认）：**

```yaml
ai:
  provider: ollama
  base_url: http://localhost:11434/v1
  model: gemma4:e4b
```

**智谱：**

```yaml
ai:
  provider: zhipu
  base_url: https://open.bigmodel.cn/api/paas/v4
  api_key: xxx
  model: glm-4
```

Embedding 配置结构相同，使用 `embedding` 字段。

## 检索模式

`chat` 命令支持三种检索模式，通过 `-m` 指定：

| 模式 | 说明 | 依赖 |
|------|------|------|
| `auto`（默认） | LLM 自动选择数据库查询或向量搜索 | AI 服务 |
| `db` | 仅数据库查询（关键词/标签/时间） | 仅数据库 |
| `vector` | 仅向量语义搜索 | AI + Embedding 服务 |

## 管道输入

`memo create` 支持管道输入，优先级：`--content` / `--file` > 管道输入 > 交互式输入。

```bash
echo "内容" | memos-cli memo create
cat note.md | memos-cli memo create
memos-cli memo create --content="直接指定内容"
```

## 数据存储

所有数据存储在 `~/.cicbyte/memos-cli/` 目录下：

```
~/.cicbyte/memos-cli/
├── config/
│   └── config.yaml    # 应用配置
├── db/
│   └── app.db         # SQLite 数据库
└── logs/
    └── app.log        # 日志文件（自动轮转）
```

## 技术栈

- Go 1.24
- [Bubbletea](https://github.com/charmbracelet/bubbletea) / [Lipgloss](https://github.com/charmbracelet/lipgloss) / [Glamour](https://github.com/charmbracelet/glamour) — TUI 框架 + Markdown 渲染
- [Cobra](https://github.com/spf13/cobra) — CLI 框架
- [mcp-go](https://github.com/mark3labs/mcp-go) — MCP Server
- [GORM](https://gorm.io/) + SQLite — 数据持久化
- [go-openai](https://github.com/sashabaranov/go-openai) — OpenAI 兼容 API 客户端（function calling）
- [Resty](https://github.com/go-resty/resty) — HTTP 客户端（Embedding）
- [Zap](https://github.com/uber-go/zap) — 结构化日志

## AI Skill

项目附带 `memos-cli` skill，指导 AI 正确使用 memos-cli 命令。skill 文件位于 `skills/memos-cli/`，可通过 junction 链接到 `.claude/skills/` 供 Claude Code 使用：

```bash
powershell -Command "New-Item -ItemType Junction -Path '.claude\skills\memos-cli' -Target 'skills\memos-cli'"
```

Skill 会自动约束 AI 的行为：
- 禁止启动交互式 TUI（`memos-cli` 无参数）
- 禁止执行需要交互式输入的命令（如无参数的 `auth login`、`server add`）
- 禁止修改配置（`config set`）
- 只允许执行非交互式的只读/写入命令（`memo list`、`memo get`、`memo create -c "..."`、`sync`、`chat "问题"` 等）

也可在 Cherry Studio 等 AI 客户端中作为 MCP 上下文使用。

## 许可证

[MIT](LICENSE)
