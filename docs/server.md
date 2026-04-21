# server

管理 Memos 服务器配置。支持添加、列出、删除服务器和设置默认服务器。

## 用法

```bash
memos-cli server [command]
```

## 子命令

| 命令 | 别名 | 说明 |
|------|------|------|
| [add](#server-add) | — | 添加服务器配置 |
| [list](#server-list) | `ls` | 列出所有服务器配置 |
| [default](#server-default) | — | 设置默认服务器 |
| [remove](#server-remove) | `rm`, `delete` | 删除服务器配置 |

## 配置文件

配置保存在 `~/.cicbyte/memos-cli/config/config.yaml`。也可以直接编辑此文件来管理配置。

---

## server add

添加一个新的 Memos 服务器配置。

```bash
memos-cli server add [flags]
```

### 选项

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--name` | `-n` | — | 服务器名称（必填） |
| `--url` | `-u` | — | 服务器 URL（必填），如 `https://memos.example.com` |
| `--token` | `-t` | — | 认证 Token（可选，也可后续通过 `auth login` 设置） |
| `--default` | `-d` | false | 设为默认服务器（第一个服务器自动设为默认） |

### 示例

```bash
# 交互式添加
memos-cli server add

# 通过参数添加
memos-cli server add --name=work --url=https://memos.example.com --token=your-token

# 添加并设为默认
memos-cli server add --name=personal --url=https://memo.example.com --default
```

---

## server list

列出所有已配置的服务器。

```bash
memos-cli server list
```

### 示例

```bash
memos-cli server list
```

---

## server default

设置默认服务器。默认服务器在执行 `sync`、`chat` 等命令时自动使用。

```bash
memos-cli server default <server-name>
```

### 示例

```bash
memos-cli server default personal
```

---

## server remove

删除服务器配置。如果删除的是默认服务器，会自动将第一个剩余服务器设为默认。

```bash
memos-cli server remove <server-name> [flags]
```

### 选项

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--force` | `-f` | false | 跳过确认直接删除 |

### 示例

```bash
memos-cli server remove work
memos-cli server remove personal --force
```
