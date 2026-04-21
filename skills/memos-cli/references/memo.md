# 备忘录管理

## memo list

```bash
memos-cli memo list [flags]
```

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--limit` | `-l` | 20 | 每页数量 |
| `--visibility` | `-v` | — | 过滤可见性：`public`/`private`/`protected` |
| `--tag` | `-t` | — | 按标签过滤 |
| `--page` | `-p` | — | 页码：`2`、`1,3,5`（多页）、`all`（全部） |
| `--search` | `-s` | — | 文本模糊搜索 |
| `--archived` | — | false | 显示已归档 |

```bash
memos-cli memo list                           # 最近 20 条
memos-cli memo list --page=2                  # 第 2 页
memos-cli memo list --tag=work                # 按标签过滤
memos-cli memo list --search="会议记录"        # 文本搜索
memos-cli memo list --page=all                # 全部
```

## memo stats

显示统计概览：总数、可见性分布、热门标签、最近备忘录。

```bash
memos-cli memo stats
```

## memo get

```bash
memos-cli memo get <id> [--raw|-r]
```

`--raw` 仅输出原始内容，适合管道处理。

## memo create

输入优先级：`--content`/`--file` > 管道输入 > 交互式输入。

```bash
memos-cli memo create -c "内容" -v private     # 参数创建
memos-cli memo create -f note.md               # 从文件
echo "内容" | memos-cli memo create            # 管道
memos-cli memo create                          # 交互式
```

## memo update

```bash
memos-cli memo update <id> [flags]
```

| 标志 | 说明 |
|------|------|
| `-c` | 新内容，传 `-` 进入交互式编辑 |
| `-v` | 修改可见性 |
| `--archive` | 归档 |
| `--restore` | 恢复归档 |
| `--pin` / `--unpin` | 置顶/取消置顶 |

## memo delete

```bash
memos-cli memo delete <id> [--force|-f]
```
