# chat

基于 AI Agent 模式与备忘录对话。LLM 自动选择合适的检索方式（数据库查询或向量搜索），找到相关备忘录后生成回答。支持流式输出，实时显示工具调用状态。

## 用法

```bash
memos-cli chat [问题] [flags]
```

## 描述

chat 命令内置 RAG Agent，通过 function calling 让 LLM 自主决策如何检索备忘录：

1. **memo_stats** — 获取备忘录统计概览（总数、标签分布、最近备忘录）
2. **search_memos** — 按关键词、标签、时间范围在数据库中精确查询
3. **semantic_search** — 基于向量相似度的语义搜索
4. **get_memo** — 获取单条备忘录详情

Agent 会根据问题自动选择最合适的工具，无需手动指定检索方式。

### 流式输出 + Markdown 渲染

回答内容逐 token 实时输出，完成后自动清除原始文本并渲染为 Markdown 格式（加粗、列表、代码块等）。工具调用时显示进度状态：

```
  > search_memos: {"keywords":"部署方案"}
  找到 2 条备忘录

  根据你的备忘录，找到以下相关内容：(逐字流式输出)
  ... (完成后自动替换为渲染后的 Markdown)

  参考来源 (2条)
  ─────────────────────────────────────────────────
  1. aBcDeFgHiJkLmNoPqRsTuV  2025-12-18
     Docker Compose 部署配置
  2. xYzAbCdEfGhIjKlMnOpQrSt  2025-07-17
     Nginx 反向代理

  Token: 1850 + 320 = 2170 · 2.8s
```

## 前置条件

- 需要先同步数据：`memos-cli sync`
- 需要配置 AI 服务（`memos-cli config set ai.provider/model/base_url/api_key`）

## 参数

| 参数 | 说明 |
|------|------|
| `问题` | 要查询的内容（单轮模式下必填） |

## 选项

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--mode` | `-m` | `auto` | 检索模式：`auto`（LLM 自动选择）、`db`（仅数据库查询）、`vector`（仅向量搜索） |
| `--tag` | — | — | 按标签过滤 |
| `--visibility` | — | — | 按可见性过滤：`PUBLIC` / `PRIVATE` / `PROTECTED` |
| `--limit` | `-l` | 10 | 返回的相关备忘录数量上限 |
| `--show-memos` | — | false | 显示参考来源完整内容（默认只显示首行预览） |
| `--output` | `-o` | — | 将对话记录保存到文件 |
| `--interactive` | `-i` | false | 进入多轮对话模式 |

## 检索模式

| 模式 | 说明 | 依赖 |
|------|------|------|
| `auto`（默认） | LLM 自动选择数据库查询或向量搜索 | AI 服务 |
| `db` | 仅数据库查询（关键词/标签/时间），速度快，不依赖 Embedding | 仅数据库 |
| `vector` | 仅向量语义搜索，适合模糊/概念性查询 | AI + Embedding 服务 |

## 多轮对话

使用 `-i` 进入多轮交互模式：

```bash
memos-cli chat -i
memos-cli chat -i "先总结最近的工作"
```

多轮模式下：
- 对话历史自动传入 LLM，支持上下文连续提问
- 输入 `/quit`、`/exit` 或 `q` 退出
- 输入 `/clear` 清除对话上下文

## 示例

```bash
# 单轮对话
memos-cli chat "我上周有哪些工作计划？"

# 查看备忘录概览
memos-cli chat "一共有多少条备忘录"

# 指定检索模式
memos-cli chat "关于React的笔记" -m db
memos-cli chat "微服务架构思考" -m vector

# 多轮对话
memos-cli chat -i

# 多轮对话并携带初始问题
memos-cli chat -i "帮我总结最近的工作"

# 显示参考来源完整内容
memos-cli chat "会议记录" --show-memos

# 保存对话记录
memos-cli chat "技术方案" -o summary.md
```
