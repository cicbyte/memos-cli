# 配置参考

配置文件路径：`~/.cicbyte/memos-cli/config/config.yaml`

## config 命令

```bash
memos-cli config list              # 列出所有配置
memos-cli config get <key>         # 查看配置值
memos-cli config get <key> --show  # 查看敏感字段明文
memos-cli config set <key> [value] # 设置配置（敏感字段不传 value 进入交互式输入）
```

## 配置项

### AI（LLM 对话）

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `ai.provider` | `ollama` | 提供商：`ollama`/`openai`/`zhipu` |
| `ai.base_url` | `http://localhost:11434/v1` | API 地址 |
| `ai.model` | `gemma4:e4b` | 模型名称 |
| `ai.api_key` | — | API 密钥 |
| `ai.max_tokens` | `2048` | 最大 token 数 |
| `ai.temperature` | `0.8` | 温度 (0.0-2.0) |
| `ai.timeout` | `60` | 超时秒数 |

### Embedding（向量嵌入）

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `embedding.provider` | `ollama` | 提供商 |
| `embedding.base_url` | `http://localhost:11434/v1` | API 地址 |
| `embedding.model` | `nomic-embed-text` | 模型名称 |
| `embedding.api_key` | — | API 密钥 |
| `embedding.timeout` | `60` | 超时秒数 |

### Log（日志）

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `log.level` | `info` | 级别 |
| `log.max_size` | `10` | 单文件最大 MB |
| `log.max_backups` | `30` | 备份数量 |
| `log.max_age` | `30` | 保留天数 |

## 常用配置示例

```bash
# Ollama 本地模型
memos-cli config set ai.provider ollama
memos-cli config set ai.base_url http://localhost:11434/v1
memos-cli config set ai.model gemma4:e4b

# OpenAI
memos-cli config set ai.provider openai
memos-cli config set ai.base_url https://api.openai.com/v1
memos-cli config set ai.model gpt-4o
memos-cli config set ai.api_key sk-xxx

# 智谱
memos-cli config set ai.provider zhipu
memos-cli config set ai.base_url https://open.bigmodel.cn/api/paas/v4
memos-cli config set ai.model glm-4
memos-cli config set ai.api_key your-key
```
