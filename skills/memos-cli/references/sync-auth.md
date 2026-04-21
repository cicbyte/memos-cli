# 同步与认证

## sync — 同步远程数据

```bash
memos-cli sync [flags]
```

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--full` | `-f` | false | 全量同步（清除本地重新同步） |
| `--force` | `-F` | false | 全量同步需配合使用 |
| `--no-vectorize` | — | false | 跳过向量化 |
| `--verbose` | `-v` | false | 显示详细输出 |

使用 MD5 内容哈希增量同步。新备忘录自动生成向量嵌入（如 embedding 服务可用）。

```bash
memos-cli sync                # 增量同步
memos-cli sync --full --force # 全量同步
memos-cli sync status         # 查看同步状态
```

## auth — 认证管理

### auth login

```bash
memos-cli auth login [flags]
```

| 标志 | 简写 | 说明 |
|------|------|------|
| `--name` | `-n` | 服务器名称 |
| `--url` | `-u` | 服务器 URL |
| `--username` | `-U` | 用户名 |
| `--password` | `-p` | 密码 |
| `--token` | `-t` | 直接用 Token 认证 |

支持用户名/密码和 Token 两种方式，不传参数时进入交互式输入。

### auth logout

```bash
memos-cli auth logout [--server|-s 名称] [--force|-f]
```

### auth status

```bash
memos-cli auth status
```

显示服务器信息、认证状态、用户信息。

## server — 服务器管理

```bash
memos-cli server add [--name/-n] [--url/-u] [--token/-t] [--default/-d]
memos-cli server list     # 列出所有服务器
memos-cli server default <名称>  # 设置默认服务器
memos-cli server remove <名称> [--force/-f]
```
