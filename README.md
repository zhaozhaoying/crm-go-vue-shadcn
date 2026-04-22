# CRM Go + Vue Monorepo

当前仓库已经整理为“仓库根即 Go 后端模块根”的单体仓布局：

- Go 后端直接位于仓库根
- 主后台前端位于 `web/`
- 子应用统一收敛到 `apps/`
- 本地构建产物统一收敛到 `build/`

## 目录结构

```text
.
├── .env.example
├── .env.local.example
├── .air.toml
├── go.mod
├── go.sum
├── main.go
├── cmd/
├── internal/
├── docs/
├── web/
├── apps/
│   ├── attendance-h5/
│   └── mihua-token-service/
├── scripts/
└── build/                 # 本地构建 / 临时产物，默认不入库
```

## 本地开发

### 1. 后端

推荐先从根目录环境示例复制一份本地配置：

```bash
cp .env.local.example .env
```

然后在仓库根执行：

```bash
go mod tidy
go run ./cmd/migrate up
go run ./cmd/bootstrap
go run .
```

后端地址：

- `http://localhost:8080`
- 健康检查：`http://localhost:8080/api/health`

### 2. 主前端

```bash
cd web
pnpm install
pnpm dev
```

主前端地址：`http://localhost:5173`

如果前端不是通过同域 Nginx 反代 `/api`，可以在 `web/.env` 中配置：

```env
VITE_API_BASE_URL=http://localhost:8080/api
```

### 3. H5 签到端

`apps/attendance-h5/` 是 uni-app H5 子应用，保留现有 uni-app 工作流。构建后的 H5 产物路径约定为：

```text
apps/attendance-h5/unpackage/dist/build/web/
```

### 4. 米话 Token 工具

`apps/mihua-token-service/` 是 Python 子应用，用于获取并回写米话 token。详情见：

- [`apps/mihua-token-service/README.md`](/Users/zhangyang/dev/zhaozhaoying/crm-go-vue-shadcn/apps/mihua-token-service/README.md)

## 常用命令

### 后端热更新

```bash
go install github.com/air-verse/air@latest
air
```

### 迁移状态

```bash
go run ./cmd/migrate status
```

### Google Search Check

```bash
go run ./cmd/google-search-check --keyword "led light manufacturer"
```

## 环境变量

根目录保留两份示例文件：

- [`/.env.local.example`](/Users/zhangyang/dev/zhaozhaoying/crm-go-vue-shadcn/.env.local.example)：本地开发示例
- [`/.env.example`](/Users/zhangyang/dev/zhaozhaoying/crm-go-vue-shadcn/.env.example)：生产部署示例

生产环境至少需要确认这些变量：

- `APP_ENV`
- `APP_PORT`
- `FRONTEND_ORIGIN`
- `DB_DRIVER`
- `MYSQL_*`
- `JWT_SECRET`
- `HANGHANG_CRM_CLOUD_TOKEN`
- `OSS_*`
- `BAIDU_MAP_AK`
- `GOOGLE_API_KEY`
- `GOOGLE_CX`

## 打包产物

使用根目录脚本统一打包：

```bash
bash ./scripts/package-release.sh
```

默认输出到：

```text
build/release/
├── crm-backend
├── dist/
├── check-in/
└── mihua-token-fetcher/
```

可覆盖的常用环境变量：

- `RELEASE_DIR`，默认 `build/release`
- `BACKEND_BIN_NAME`，默认 `crm-backend`
- `INCLUDE_BACKEND`，默认 `1`
- `INCLUDE_WEB`，默认 `1`
- `INCLUDE_ATTENDANCE_H5`，默认 `1`
- `INCLUDE_MIHUA_TOKEN`，默认 `1`
- `GOOS_TARGET`，默认 `linux`
- `GOARCH_TARGET`，默认 `amd64`
- `CGO_ENABLED_TARGET`，默认 `1`
- `CLEAN_BUILD_TARGET=1` 时会先清理 Go 构建缓存

示例：

```bash
GOOS_TARGET=linux GOARCH_TARGET=amd64 CGO_ENABLED_TARGET=0 bash ./scripts/package-release.sh
```

## 线上部署

### systemd

仓库根已经是后端工作目录，`systemd` 示例应改为：

```ini
[Unit]
Description=CRM Backend
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/srv/crm-go-vue-shadcn
ExecStart=/srv/crm-go-vue-shadcn/crm-backend
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Nginx

如果前后端同域部署，主前端静态目录可指向：

```nginx
root /srv/crm-go-vue-shadcn/web/dist;
index index.html;

location /api/ {
    proxy_pass http://127.0.0.1:8080/api/;
}

location / {
    try_files $uri $uri/ /index.html;
}

location /check-in/ {
    alias /srv/crm-go-vue-shadcn/apps/attendance-h5/unpackage/dist/build/web/;
    try_files $uri $uri/ /check-in/index.html;
}
```

如果你是通过打包目录上线，也可以把 `/check-in/` 指向发布后的静态目录。

## 本地 release 准备

使用：

```bash
bash ./scripts/deploy.sh
```

脚本默认行为：

- 本地先执行 `scripts/package-release.sh`
- 默认使用 `build/release/` 作为本地产物目录
- 默认后端二进制名为 `overseas_linux`
- 默认把主前端收敛到 `build/release/dist/`
- 默认把 `attendance-h5` 收敛到 `build/release/check-in/`
- 默认把 `apps/mihua-token-service/` 收敛到 `build/release/mihua-token-fetcher/`
- 不执行任何远端上传、SSH、服务重启动作

常用参数：

```bash
# 清理 Go 构建缓存后重新打包
bash ./scripts/deploy.sh --clean-build

# 只准备前端 release 产物
bash ./scripts/deploy.sh --frontend-only

# 只准备后端 release 产物
bash ./scripts/deploy.sh --backend-only

# 跳过本地构建，直接复用 build/release
bash ./scripts/deploy.sh --skip-build
```

## 上线检查清单

- `curl -s http://127.0.0.1:8080/api/health` 返回正常
- 前端页面可打开，`/api` 代理正常
- 登录、刷新 token、退出登录正常
- 图片上传正常
- 数据库已经切到 MySQL 或者明确知道当前仍在用 SQLite
- 管理员账号已初始化
- 合同、录音、资源池等核心业务页面已抽样验证
