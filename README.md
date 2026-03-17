# CRM Go + Vue Starter

This repository is initialized as a frontend-backend separated project:

- `backend`: Go + Gin
- `frontend`: Vue 3 + Vite + Tailwind + shadcn-vue compatible structure
- Layered backend + router/pinia/api frontend are scaffolded

## Directory

```text
.
├── backend
│   ├── internal
│   │   ├── config
│   │   ├── handler
│   │   ├── middleware
│   │   ├── model
│   │   ├── repository
│   │   ├── router
│   │   └── service
│   ├── .env.example
│   ├── go.mod
│   └── main.go
└── frontend
    ├── src
    │   ├── api
    │   ├── router
    │   ├── stores
    │   └── views
    ├── components.json
    ├── package.json
    ├── postcss.config.js
    ├── tailwind.config.ts
    ├── tsconfig.json
    ├── vite.config.ts
    └── src
```

## Run Backend

```bash
cd backend
go mod tidy
go run ./cmd/migrate up
# 可选：初始化管理员与字典数据
go run ./cmd/bootstrap
go run .
```

Backend URL: `http://localhost:8080`  
Health check: `http://localhost:8080/api/health`
Customers: `http://localhost:8080/api/v1/customers`

## Run Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend URL: `http://localhost:5173`

If your frontend is served by nginx and `/api` is not proxied, set `frontend/.env` with:
`VITE_API_BASE_URL=http://localhost:8080/api`

## Implemented demo pages

- `/login`: split-screen login page
- `/dashboard`: backend health dashboard
- `/customers`: customers list page (via axios + API module)

## Deployment Guide

下面给出一套适合当前仓库的生产部署方式：

- 前端：Vite build 后由 Nginx 托管静态文件
- 后端：Go 二进制 + `systemd` 常驻
- 数据库：推荐 MySQL，不建议生产继续使用 SQLite
- 域名：同域部署，前端走 `/`，后端 API 走 `/api`

### 1. Recommended Server Stack

建议至少准备以下运行环境：

- Ubuntu 22.04 / Debian 12
- Go `1.24.x`
- Node.js `20.x`
- pnpm `10.x`
- Nginx
- MySQL `8.x`

示例目录：

```bash
/srv/crm-go-vue-shadcn
├── backend
├── frontend
└── logs
```

### 2. Upload Project Code

推荐直接在服务器拉代码：

```bash
cd /srv
git clone <your-repo-url> crm-go-vue-shadcn
cd crm-go-vue-shadcn
```

如果你是本地打包后上传，也建议仍然保留完整仓库，方便后续更新和回滚。

### 3. Backend Production Config

复制后端环境变量模板：

```bash
cd /srv/crm-go-vue-shadcn/backend
cp .env.example .env
```

生产环境至少要改这些：

```env
APP_ENV=production
APP_PORT=8080
FRONTEND_ORIGIN=https://crm.example.com

DB_DRIVER=mysql
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=crm_user
MYSQL_PASSWORD=replace_me
MYSQL_DB=crm_db

JWT_SECRET=replace_with_a_long_random_secret
JWT_EXPIRY_HOURS=24
REFRESH_TOKEN_EXPIRY_HOURS=168

OSS_ENDPOINT=
OSS_ACCESS_KEY_ID=
OSS_ACCESS_KEY_SECRET=
OSS_BUCKET_NAME=
OSS_BASE_PATH=avatars/
```

说明：

- `JWT_SECRET` 在生产环境必须显式配置，否则后端启动会失败
- `FRONTEND_ORIGIN` 填你的正式前端域名，多个域名用英文逗号分隔
- 图片上传依赖 OSS，`OSS_*` 不完整时上传功能会出问题
- 资源池相关功能如果要用，需要继续补 `BAIDU_MAP_AK`、`GOOGLE_API_KEY`、`GOOGLE_CX` 等变量

### 4. Frontend Production Config

复制前端环境变量模板：

```bash
cd /srv/crm-go-vue-shadcn/frontend
cp .env.example .env.production
```

如果前端和后端走同一个域名，建议这样配：

```env
VITE_API_BASE_URL=/api
VITE_BAIDU_MAP_AK=
VITE_BAIDU_MAP_REVERSE_GEO_URL=https://api.map.baidu.com/reverse_geocoding/v3/
```

如果前后端分域，再改成完整地址，例如：

```env
VITE_API_BASE_URL=https://api.example.com/api
```

### 5. Prepare Database

先创建 MySQL 数据库和账号：

```sql
CREATE DATABASE crm_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'crm_user'@'127.0.0.1' IDENTIFIED BY 'replace_me';
GRANT ALL PRIVILEGES ON crm_db.* TO 'crm_user'@'127.0.0.1';
FLUSH PRIVILEGES;
```

### 6. Build And Initialize

先构建前后端：

```bash
cd /srv/crm-go-vue-shadcn
bash ./scripts/build-cn.sh
```

如果你的服务器在国内，推荐直接用仓库里的 `scripts/build-cn.sh`，它做了这几件事：

- 前端固定使用 `npm ci`，避免 `corepack enable` 在线拉取 pnpm 时卡住
- 前端默认读取 `frontend/.npmrc`，走 `npmmirror`
- 后端默认设置 `GOPROXY=https://goproxy.cn,https://goproxy.io,direct`
- 后端默认设置 `GOSUMDB=sum.golang.google.cn`
- 后端默认设置 `GOTOOLCHAIN=local`，避免因为 `backend/go.mod` 里的 `go 1.24.0` 触发自动下载 Go 工具链

如果你的服务器本机 Go 版本低于 `1.24`，请先手动安装 Go `1.24+` 或 `1.25+`，再执行脚本。

首次部署建议手动执行一次迁移和初始化：

```bash
cd /srv/crm-go-vue-shadcn/backend
go run ./cmd/migrate up
go run ./cmd/bootstrap
```

如果你在 macOS 本地交叉编译 Linux 版本，推荐优先使用下面这条：

```bash
cd /srv/crm-go-vue-shadcn/backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o crm-backend ./
```

这条命令更适合当前项目，优点是：

- 不依赖 Linux 交叉编译工具链
- 在 macOS 上可直接打 Linux 包
- 生成的二进制更方便直接上传到 Linux 服务器运行

如果你是在 Linux 服务器本机编译，或者你已经装好了 Linux 交叉编译器，也可以使用：

```bash
cd /srv/crm-go-vue-shadcn/backend
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o crm-backend ./
```

说明：

- 你如果在 macOS 上直接执行 `CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build ...`，很容易因为本机使用的是 Apple clang 和 macOS SDK，无法编译 Linux 的 cgo 代码而报错
- 出现这类错误时，直接改用 `CGO_ENABLED=0` 即可
- 如果必须启用 `CGO_ENABLED=1`，建议改为在 Linux 环境中编译，或者使用专门的 Linux 交叉编译器

如果你想把前端 `dist` 和 Linux 后端程序一起打到 `release/` 目录，仓库里已经带了脚本：

```bash
cd /srv/crm-go-vue-shadcn
GOOS_TARGET=linux GOARCH_TARGET=amd64 CGO_ENABLED_TARGET=1 bash ./scripts/package-release.sh
```

打包完成后会生成：

```text
release/
├── crm-backend
├── .env
└── dist/
```

说明：

- `bootstrap` 会初始化管理员、角色、客户级别、客户来源、跟进方式
- 如果你已经初始化过，后续再次执行也是幂等的

### 7. systemd For Backend

创建服务文件：

```bash
sudo vim /etc/systemd/system/crm-backend.service
```

写入：

```ini
[Unit]
Description=CRM Backend
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/srv/crm-go-vue-shadcn/backend
ExecStart=/srv/crm-go-vue-shadcn/backend/crm-backend
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable crm-backend
sudo systemctl start crm-backend
sudo systemctl status crm-backend
```

查看日志：

```bash
journalctl -u crm-backend -f
```

### 8. Nginx For Frontend And API Proxy

创建站点配置：

```bash
sudo vim /etc/nginx/sites-available/crm.conf
```

示例配置：

```nginx
server {
    listen 80;
    server_name crm.example.com;

    root /srv/crm-go-vue-shadcn/frontend/dist;
    index index.html;

    client_max_body_size 20m;

    location /api/ {
        proxy_pass http://127.0.0.1:8080/api/;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

启用站点：

```bash
sudo ln -s /etc/nginx/sites-available/crm.conf /etc/nginx/sites-enabled/crm.conf
sudo nginx -t
sudo systemctl reload nginx
```

如果要上 HTTPS，建议再配一层 Certbot：

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d crm.example.com
```

### 9. Release Workflow

后续每次发布建议按这个顺序：

```bash
cd /srv/crm-go-vue-shadcn
git pull

bash ./scripts/build-cn.sh

cd backend
go run ./cmd/migrate up

sudo systemctl restart crm-backend
sudo nginx -t && sudo systemctl reload nginx
```

### 10. Deployment Checklist

上线前至少确认这些：

- 后端健康检查通过：`http://127.0.0.1:8080/api/health`
- 前端页面能正常打开
- 登录、刷新 token、退出登录正常
- 图片上传正常
- MySQL 已连接，不再使用 SQLite
- 管理员账号已初始化
- 合同模块实测通过
  - 销售编辑权限正常
  - 运营只能改站点与服务
  - 财务经理和管理员可见审核按钮
  - 开启“已上线”后，保存时后端能生成合同开始时间和结束时间

### 11. Troubleshooting

常见问题排查：

- 前端接口 404
  - 检查 `VITE_API_BASE_URL` 是否为 `/api`
  - 检查 Nginx 是否正确代理了 `/api/`
- 登录跨域失败
  - 检查后端 `.env` 中 `FRONTEND_ORIGIN`
- 图片上传失败
  - 检查 `OSS_*` 是否完整
- 服务启动失败
  - 先看 `journalctl -u crm-backend -f`
- 合同开始时间/结束时间没有生成
  - 确认后端服务已经重启到最新代码
  - 重新保存一次合同，让后端执行最新提交逻辑

### 12. Suggested Next Step

如果你准备正式长期运维，建议下一步补这几样：

- `Dockerfile` 和 `docker-compose.yml`
- 一键发布脚本 `deploy.sh`
- 生产环境备份脚本
- CI/CD 自动构建和发布流程
