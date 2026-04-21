# AI 对话

基于 RAG Agent 模式与备忘录对话，LLM 自动选择检索方式。

## 用法

```bash
memos-cli chat [问题] [flags]
```

## 选项

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--mode` | `-m` | `auto` | 检索模式：`auto`/`db`/`vector` |
| `--tag` | — | — | 按标签过滤 |
| `--visibility` | — | — | 按可见性过滤 |
| `--limit` | `-l` | 10 | 相关备忘录数量上限 |
| `--show-memos` | — | false | 显示参考来源完整内容 |
| `--output` | `-o` | — | 保存对话记录到文件 |
| `--interactive` | `-i` | false | 多轮对话模式 |

## 检索模式

| 模式 | 说明 | 依赖 |
|------|------|------|
| `auto` | LLM 自动选择查询方式 | AI 服务 |
| `db` | 仅数据库查询（关键词/标签/时间），速度快 | 仅数据库 |
| `vector` | 仅向量语义搜索，适合模糊/概念性查询 | AI + Embedding |

## AI Agent 工具

Agent 通过 function calling 自动选择：

1. **memo_stats** — 备忘录统计概览
2. **search_memos** — 关键词、标签、时间范围精确查询
3. **semantic_search** — 向量相似度语义搜索（支持时间过滤）
4. **get_memo** — 单条备忘录详情

## 多轮对话

```bash
memos-cli chat -i                           # 进入多轮模式
memos-cli chat -i "先总结最近的工作"         # 携带初始问题
```

- 输入 `/quit`、`/exit` 或 `q` 退出
- 输入 `/clear` 清除上下文

## 前置条件

- 需要 AI 服务配置（`memos-cli config set ai.provider/model/base_url/api_key`）
- `vector` 模式需要 Embedding 服务配置
- 需要先 `memos-cli sync` 同步数据
