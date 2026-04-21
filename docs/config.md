# config

管理 memos-cli 应用配置（AI、Embedding、日志等参数）。

## 用法

```bash
memos-cli config [command]
```

## 子命令

| 命令 | 说明 |
|------|------|
| [list](#config-list) | 列出所有配置项及当前值 |
| [get](#config-get) | 查看单个配置项的值 |
| [set](#config-set) | 设置配置项 |

## 敏感字段保护

涉及 API 密钥等敏感字段（如 `ai.api_key`、`embedding.api_key`）：

- `config list` 中显示为 `******`
- `config get` 中默认显示为 `******`，使用 `--show` 查看明文
- `config set` 不传 value 时进入交互式输入（不回显）

---

## config list

以表格形式列出所有配置项、当前值和说明。

```bash
memos-cli config list
```

### 示例

```
╭─────────────────────────────────────────────────────╮
│  KEY                   VALUE                         DESCRIPTION            │
├─────────────────────────────────────────────────────┤
│  [AI]                                                        │
│  ai.provider           ollama                        LLM 提供商             │
│  ai.base_url           http://localhost:11434/v1     LLM API 地址           │
│  ai.model              llama3.2                      LLM 模型名称           │
│  ai.api_key            ******                        LLM API 密钥           │
│  ai.max_tokens         2048                          最大 token 数          │
│  ai.temperature        0.8                           温度参数 (0.0-2.0)     │
│  ai.timeout            60                            请求超时秒数           │
│  [Embedding]                                                  │
│  embedding.provider     ollama                        Embedding 提供商        │
│  embedding.base_url     http://localhost:11434/v1     Embedding API 地址      │
│  embedding.model        nomic-embed-text              Embedding 模型名称      │
│  embedding.api_key     ******                        Embedding API 密钥      │
│  embedding.timeout     60                            请求超时秒数           │
│  [Log]                                                      │
│  log.level             info                          日志级别               │
│  log.max_size          10                            单个日志文件最大 MB      │
│  log.max_backups       30                            日志备份数量           │
│  log.max_age           30                            日志保留天数           │
│  log.compress          true                          是否压缩日志           │
╰─────────────────────────────────────────────────────────╯
```

---

## config get

查看指定配置项的值。

```bash
memos-cli config get <key> [flags]
```

### 选项

| 标志 | 默认值 | 说明 |
|------|--------|------|
| `--show` | false | 显示敏感字段的明文值 |

### 示例

```bash
memos-cli config get ai.model
# 输出: llama3.2

memos-cli config get ai.api_key
# 输出: ******

memos-cli config get ai.api_key --show
# 输出: sk-xxx（明文）
```

---

## config set

设置指定配置项的值。

```bash
memos-cli config set <key> [value] [flags]
```

### 敏感字段

敏感字段（如 `ai.api_key`）不传 value 时进入交互式输入，密码不回显。

### 示例

```bash
memos-cli config set ai.model qwen2.5
memos-cli config set ai.temperature 0.7
memos-cli config set log.level debug
memos-cli config set ai.api_key sk-xxx
memos-cli config set ai.api_key        # 交互式输入（不回显）
```
