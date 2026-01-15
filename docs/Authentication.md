# Authentication Configuration

go2short 支持多种认证方式：

| 方式 | 用途 | 配置 |
|------|------|------|
| 邮箱+密码 | 普通用户注册/登录 | 无需配置，默认启用 |
| Google OAuth | 用户快速登录 | 需配置 |
| GitHub OAuth | 用户快速登录 | 需配置 |
| 超级管理员 | 系统管理 | 环境变量 |

---

## 环境变量

在 `.env` 文件或环境中设置：

```bash
# 必需
BASE_URL=https://your-domain.com    # OAuth 回调依赖此地址

# 超级管理员（默认 admin/admin123）
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your-secure-password
ADMIN_TOKEN_TTL=24h

# Google OAuth（可选）
GOOGLE_CLIENT_ID=xxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=xxx

# GitHub OAuth（可选）
GITHUB_CLIENT_ID=xxx
GITHUB_CLIENT_SECRET=xxx
```

---

## Google OAuth 配置

1. 访问 [Google Cloud Console](https://console.cloud.google.com/apis/credentials)

2. 创建项目（或选择已有项目）

3. 配置 OAuth 同意屏幕：
   - User Type: External
   - 填写应用名称、用户支持邮箱
   - Scopes 添加: `email`, `profile`

4. 创建凭据：
   - 类型: OAuth 2.0 Client ID
   - 应用类型: Web application
   - Authorized redirect URIs:
     ```
     https://your-domain.com/api/auth/google/callback
     ```

5. 复制 Client ID 和 Client Secret 到环境变量

---

## GitHub OAuth 配置

1. 访问 [GitHub Developer Settings](https://github.com/settings/developers)

2. OAuth Apps → New OAuth App

3. 填写信息：
   - Application name: `go2short`
   - Homepage URL: `https://your-domain.com`
   - Authorization callback URL:
     ```
     https://your-domain.com/api/auth/github/callback
     ```

4. 复制 Client ID，生成 Client Secret

5. 配置到环境变量

---

## 权限模型

| 角色 | 识别方式 | 数据范围 |
|------|---------|---------|
| 超级管理员 | `ADMIN_USERNAME/PASSWORD` 登录 | 所有数据 |
| 普通用户 | 邮箱密码或 OAuth 登录 | 仅自己的数据 |

### 数据隔离

- 每个用户只能看到/操作自己创建的短链接和 API Token
- 超级管理员可查看所有数据
- 现有数据（`user_id=NULL`）仅超管可见

---

## API 端点

### 用户认证

```
POST /api/auth/register    # 邮箱注册
POST /api/auth/login       # 邮箱登录
GET  /api/auth/providers   # 获取可用 OAuth 提供商
GET  /api/auth/google      # Google OAuth 跳转
GET  /api/auth/google/callback
GET  /api/auth/github      # GitHub OAuth 跳转
GET  /api/auth/github/callback
```

### 超级管理员

```
POST /api/admin/login      # 超管登录
POST /api/admin/logout     # 登出
```

---

## 本地开发

本地测试 OAuth 时，可使用 ngrok 或类似工具：

```bash
ngrok http 8080
```

然后将 ngrok 提供的 HTTPS URL 设置为 `BASE_URL`，并更新 OAuth 应用的回调地址。

---

## 不配置 OAuth

OAuth 是可选的。不配置时：
- 登录页面不显示 OAuth 按钮
- 用户只能通过邮箱+密码注册/登录
- 所有功能正常使用
