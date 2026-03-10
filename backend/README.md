# Backend (Go + Gin)

## Start

```bash
go mod tidy
go run ./cmd/migrate up
go run .
```

Default listen address: `:8080`

## Environment

Core auth variables:

- `JWT_SECRET`: JWT signing secret. In `APP_ENV=production|prod`, this must be explicitly configured and cannot use default placeholder.
- `JWT_EXPIRY_HOURS`: access token expiration hours (default `24`)
- `REFRESH_TOKEN_EXPIRY_HOURS`: refresh token expiration hours (default `168`)
- `BAIDU_MAP_AK`: 百度地图 Web API AK（资源池检索必填）
- `BAIDU_MAP_BASE_URL`: 百度地图 API 基础地址（默认 `https://api.map.baidu.com`）

External company search variables:

- `GOOGLE_API_KEY`: Google Custom Search API key
- `GOOGLE_CX`: Google Programmable Search Engine CX
- `GOOGLE_SEARCH_NUM`: 每页抓取数量，Google API 最大 `10`
- `GOOGLE_PROXY_URL`: 当服务所在机器无法直连 Google 时使用的代理地址，例如 `http://127.0.0.1:7890`

## Migration

```bash
# apply pending migrations
go run ./cmd/migrate up

# show migration status
go run ./cmd/migrate status
```

## Bootstrap (optional)

```bash
# initialize admin + dictionaries (idempotent)
go run ./cmd/bootstrap

# only initialize dictionaries
go run ./cmd/bootstrap --skip-admin

# only initialize admin
go run ./cmd/bootstrap --skip-dictionaries

# reset existing admin password
go run ./cmd/bootstrap --reset-admin-password --admin-password 'new_password'
```

## API

- `GET /api/health`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/auth/me` (Bearer token required)
- `POST /api/v1/auth/logout` (Bearer token required)
- `GET /api/v1/resource-pool` (Bearer token required)
- `POST /api/v1/resource-pool/search` (Bearer token required)
- `GET /api/v1/customers`

## Google Search Check

```bash
go run ./cmd/google-search-check --keyword "led light manufacturer"
```

如果当前机器无法直连 Google API，可先配置 `GOOGLE_PROXY_URL`，或者在启动进程前设置 `HTTPS_PROXY`。

## Current layered structure

- `internal/config`: environment configuration
- `internal/router`: route registration
- `internal/middleware`: middleware collection
- `internal/handler`: HTTP handlers + unified response
- `internal/service`: business logic layer
- `internal/repository`: data access layer
