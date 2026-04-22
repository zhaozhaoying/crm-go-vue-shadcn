# Backend (Go + Gin)

仓库根目录现在就是 Go 模块根，后端相关命令都直接在仓库根执行，不再需要 `cd backend`。

## Start

```bash
go mod tidy
go run ./cmd/migrate up
go run ./cmd/bootstrap
go run .
```

默认监听地址：`:8080`

## Hot Reload

后端本地热更新推荐使用 `air`：

```bash
go install github.com/air-verse/air@latest
air
```

项目已内置 [`/.air.toml`](/Users/zhangyang/dev/zhaozhaoying/crm-go-vue-shadcn/.air.toml) 配置，修改 `.go` 文件后会自动重新编译并重启服务。

主前端本身已支持热更新：

```bash
cd web
pnpm dev
```

## Environment

推荐把本地开发环境变量保存为根目录 `.env`。仓库中保留了两份示例：

- [`/.env.local.example`](/Users/zhangyang/dev/zhaozhaoying/crm-go-vue-shadcn/.env.local.example)：本地开发示例
- [`/.env.example`](/Users/zhangyang/dev/zhaozhaoying/crm-go-vue-shadcn/.env.example)：生产部署示例

核心鉴权变量：

- `JWT_SECRET`: JWT signing secret，生产环境必须显式配置
- `JWT_EXPIRY_HOURS`: access token 过期小时数，默认 `24`
- `REFRESH_TOKEN_EXPIRY_HOURS`: refresh token 过期小时数，默认 `168`
- `HANGHANG_CRM_CLOUD_TOKEN`: 航航 CRM 通话统计同步使用的 cloud-token
- `BAIDU_MAP_AK`: 百度地图 Web API AK
- `BAIDU_MAP_BASE_URL`: 百度地图 API 基础地址，默认 `https://api.map.baidu.com`

企业搜索相关变量：

- `GOOGLE_API_KEY`
- `GOOGLE_CX`
- `GOOGLE_SEARCH_NUM`
- `GOOGLE_PROXY_URL`

## Migration

```bash
# 执行未完成的迁移
go run ./cmd/migrate up

# 查看迁移状态
go run ./cmd/migrate status
```

## Bootstrap

```bash
# 初始化管理员和字典
go run ./cmd/bootstrap

# 只初始化字典
go run ./cmd/bootstrap --skip-admin

# 只初始化管理员
go run ./cmd/bootstrap --skip-dictionaries

# 重置管理员密码
go run ./cmd/bootstrap --reset-admin-password --admin-password 'new_password'
```

## API

- `GET /api/health`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/logout`
- `GET /api/v1/resource-pool`
- `POST /api/v1/resource-pool/search`
- `GET /api/v1/customers`

## Google Search Check

```bash
go run ./cmd/google-search-check --keyword "led light manufacturer"
```

如果当前机器无法直连 Google API，可先配置 `GOOGLE_PROXY_URL`，或者在启动进程前设置 `HTTPS_PROXY`。

## Current Layered Structure

- `cmd`: 可执行入口
- `internal/config`: 环境配置
- `internal/router`: 路由注册
- `internal/middleware`: 中间件集合
- `internal/handler`: HTTP handlers + unified response
- `internal/service`: 业务逻辑层
- `internal/repository`: 数据访问层
