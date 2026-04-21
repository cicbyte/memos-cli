# sync

将远程 Memos 服务器的备忘录同步到本地 SQLite 数据库。

## 用法

```bash
memos-cli sync [flags]
```

## 描述

同步服务使用 MD5 内容哈希进行增量同步，仅拉取新增或变更的备忘录。新同步的备忘录会自动生成向量嵌入（如果配置了 embedding 服务且服务可用），以支持 AI 语义搜索。

**Embedding 服务不可用时**：同步正常完成，自动跳过向量化步骤并提示。

## 前置条件

- 需要先配置服务器：`memos-cli server add` 或 `memos-cli auth login`

## 选项

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--full` | `-f` | false | 全量同步（清除本地数据重新同步） |
| `--force` | `-F` | false | 强制执行（全量同步时需要配合使用，会要求二次确认） |
| `--no-vectorize` | — | false | 同步时跳过向量化 |
| `dry-run` | — | false | 预览模式（不实际执行） |
| `--verbose` | `-v` | false | 显示详细输出（含向量统计） |

## 子命令

### sync status

显示当前同步状态信息。

```bash
memos-cli sync status
```

输出包含：服务器名称、最后同步时间、本地备忘录数量、同步状态（idle/syncing/error）。

## 示例

```bash
# 增量同步（默认）
memos-cli sync

# 全量同步（会删除所有本地数据）
memos-cli sync --full --force

# 同步但不生成向量
memos-cli sync --no-vectorize

# 查看同步状态
memos-cli sync status
```

## 输出示例

```
  正在同步备忘录...
  ✓ 已连接到服务器: https://memos.example.com
  上次同步: 2026-04-19 10:30:00

📊 同步统计:
  - 新增: 5 条
  - 更新: 2 条
  - 删除: 0 条
  - 跳过: 98 条

  💾 本地数据库总计: 105 条备忘录 · 耗时: 3.25 秒
  ✅ 同步完成!
```
