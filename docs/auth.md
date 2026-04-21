# auth

认证管理，处理与 Memos 服务器的登录、登出和状态查询。

## 用法

```bash
memos-cli auth [command]
```

## 子命令

| 命令 | 说明 |
|------|------|
| [login](#auth-login) | 登录到 Memos 服务器 |
| [logout](#auth-logout) | 从服务器登出 |
| [status](#auth-status) | 查看当前认证状态 |

---

## auth login

登录到 Memos 服务器。支持用户名/密码和 Token 两种方式。

```bash
memos-cli auth login [flags]
```

### 选项

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--name` | `-n` | — | 服务器名称（保存配置时使用），默认自动生成 `用户名@域名` |
| `--url` | `-u` | — | 服务器 URL，默认使用已配置的服务器 |
| `--username` | `-U` | — | 用户名，不传则交互式输入 |
| `--password` | `-p` | — | 密码，不传则交互式输入（输入时隐藏） |
| `--token` | `-t` | — | 直接使用 Token 认证（跳过用户名/密码） |

### 登录流程

1. 如果提供了 `--token`，直接验证 Token 并保存配置
2. 如果提供了 `--username` 和 `--password`，通过 API 登录获取 Token
3. 如果未提供参数，进入交互式输入流程

### 服务器名称生成规则

- 指定了 `--name`：使用指定名称
- 未指定但有用户名：自动生成为 `用户名@服务器域名`
- 都没有：使用服务器域名

### 示例

```bash
# 交互式登录（使用默认服务器）
memos-cli auth login

# 指定服务器登录
memos-cli auth login --url=https://memos.example.com --username=myuser

# 使用 Token 登录
memos-cli auth login --url=https://memos.example.com --token=my-token

# 登录并命名服务器
memos-cli auth login --name=work --url=https://work.example.com
```

---

## auth logout

从服务器登出，清除本地存储的认证 Token。

```bash
memos-cli auth logout [flags]
```

### 选项

| 标志 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--server` | `-s` | — | 指定要登出的服务器名称，默认使用默认服务器 |
| `--force` | `-f` | false | 跳过确认直接登出 |

### 示例

```bash
# 登出默认服务器（有确认提示）
memos-cli auth logout

# 登出指定服务器
memos-cli auth logout --server=work

# 强制登出
memos-cli auth logout --force
```

---

## auth status

显示当前认证状态，包括服务器信息、用户信息和连接状态。

```bash
memos-cli auth status
```

### 输出内容

- 服务器名称和 URL
- 认证状态（已认证 / 未认证 / 认证失败）
- 用户名、昵称、邮箱、角色（已认证时）

### 示例

```bash
memos-cli auth status
```

输出示例：

```
  Authentication Status

  Server: myuser@example.com
  URL: https://memos.example.com

  Status: Authenticated ✓

  Username: myuser
  Nickname: 我的昵称
  Email: myuser@example.com
  Role: HOST
```
