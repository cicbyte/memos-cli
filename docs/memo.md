# memo

管理备忘录，支持创建、查看、更新、删除操作。

## 用法

```bash
memos-cli memo [command]
```

## 子命令

| 命令 | 别名 | 说明 |
|------|------|------|
| [list](#memo-list) | `ls` | 列出备忘录列表 |
| [stats](#memo-stats) | — | 备忘录统计概览 |
| [get](#memo-get) | — | 查看备忘录详情 |
| [create](#memo-create) | `new`, `add` | 创建备忘录 |
| [update](#memo-update) | — | 更新备忘录 |
| [delete](#memo-delete) | `rm`, `remove` | 删除备忘录 |

---

## memo list

列出本地数据库中的备忘录。需要先执行 `memos-cli sync` 同步数据。

---

## memo stats

显示本地备忘录的统计概览，包括总数、可见性分布、热门标签和最近备忘录。

```bash
memos-cli memo stats
```

### 输出示例

```
  备忘录统计

  总计: 258 条
  可见性分布:
    PRIVATE: 240 条
    PUBLIC: 18 条
  热门标签 (Top 10):
    work: 45 条
    idea: 32 条
    ...
  最近 5 条备忘录

  1. 20260413120000-abc123 (2026-04-13) - Docker 部署配置...
  2. 20260412150000-def456 (2026-04-12) - 周会纪要...
```

---

```bash
memos-cli memo list [flags]
```

### 选项

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--limit` | `-l` | 20 | 每页显示数量 |
| `--visibility` | `-v` | — | 按可见性过滤：`public` / `private` / `protected` |
| `--tag` | `-t` | — | 按标签过滤 |
| `--page` | `-p` | — | 页码，支持 `2`、`1,3,5`（多页）、`all`（全部） |
| `--search` | `-s` | — | 文本搜索（模糊匹配内容） |
| `--from` | — | — | 起始日期（YYYY-MM-DD） |
| `--to` | — | — | 结束日期（YYYY-MM-DD，包含当天） |
| `--archived` | — | false | 显示已归档的备忘录 |

### 示例

```bash
# 列出最近 20 条备忘录
memos-cli memo list

# 查看第 2 页
memos-cli memo list --page=2

# 查看所有公开备忘录
memos-cli memo list --visibility=public

# 按标签过滤
memos-cli memo list --tag=work

# 文本搜索
memos-cli memo list --search="会议记录"

# 按时间范围查询
memos-cli memo list --from=2026-01-01
memos-cli memo list --from=2025-07-01 --to=2025-12-31

# 获取全部备忘录
memos-cli memo list --page=all
```

---

## memo get

查看备忘录的详细信息。

```bash
memos-cli memo get <memo-id> [flags]
```

### 参数

| 参数 | 说明 |
|------|------|
| `memo-id` | 备忘录 ID |

### 选项

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--raw` | `-r` | false | 仅输出原始内容 |

### 示例

```bash
# 查看详情
memos-cli memo get 123

# 仅获取原始内容（适合管道处理）
memos-cli memo get 123 --raw
```

---

## memo create

创建新备忘录。支持通过命令行参数、文件、管道或交互式输入提供内容。

```bash
memos-cli memo create [flags]
```

### 选项

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--content` | `-c` | — | 备忘录内容 |
| `--file` | `-f` | — | 从文件读取内容 |
| `--visibility` | `-v` | `private` | 可见性：`public` / `private` / `protected` |

输入优先级：`--content` / `--file` > 管道输入 > 交互式输入。

### 示例

```bash
# 通过参数创建
memos-cli memo create --content="Hello, world!"

# 创建公开备忘录
memos-cli memo create --content="公开笔记" --visibility=public

# 从文件创建
memos-cli memo create --file=note.md

# 管道输入
echo "内容" | memos-cli memo create
cat note.md | memos-cli memo create

# 交互式输入（不传参数时自动进入）
memos-cli memo create
```

---

## memo update

更新已有备忘录的内容、可见性、归档状态或置顶状态。

```bash
memos-cli memo update <memo-id> [flags]
```

### 参数

| 参数 | 说明 |
|------|------|
| `memo-id` | 备忘录 ID |

### 选项

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--content` | `-c` | — | 新内容，传 `-` 进入交互式编辑 |
| `--visibility` | `-v` | — | 修改可见性：`public` / `private` / `protected` |
| `--archive` | — | false | 归档备忘录 |
| `--restore` | — | false | 恢复已归档的备忘录 |
| `--pin` | — | false | 置顶 |
| `--unpin` | — | false | 取消置顶 |

### 示例

```bash
# 修改内容
memos-cli memo update 123 --content="新内容"

# 交互式编辑内容
memos-cli memo update 123 --content=-

# 改为公开
memos-cli memo update 123 --visibility=public

# 归档
memos-cli memo update 123 --archive

# 置顶
memos-cli memo update 123 --pin
```

---

## memo delete

删除备忘录。此操作不可撤销。

```bash
memos-cli memo delete <memo-id> [flags]
```

### 参数

| 参数 | 说明 |
|------|------|
| `memo-id` | 备忘录 ID |

### 选项

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--force` | `-f` | false | 跳过确认直接删除 |

### 示例

```bash
# 删除（有确认提示）
memos-cli memo delete 123

# 强制删除（无确认）
memos-cli memo delete 123 --force
```
