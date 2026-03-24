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
HANGHANG_CRM_CLOUD_TOKEN=

OSS_ENDPOINT=
OSS_ACCESS_KEY_ID=
OSS_ACCESS_KEY_SECRET=
OSS_BUCKET_NAME=
OSS_BASE_PATH=avatars/
```

说明：

- `JWT_SECRET` 在生产环境必须显式配置，否则后端启动会失败
- 航航 CRM 一键同步依赖后端 `.env` 中的 `HANGHANG_CRM_CLOUD_TOKEN`
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

如果你想把前端 `dist`、`check-in` 的 H5 Web 产物和 Linux 后端程序一起打到 `release/` 目录，仓库里已经带了脚本：

```bash
cd /srv/crm-go-vue-shadcn
GOOS_TARGET=linux GOARCH_TARGET=amd64 CGO_ENABLED_TARGET=1 bash ./scripts/package-release.sh
```

打包完成后会生成：

```text
release/
├── crm-backend
├── dist/
└── check-in/
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

注意：

- `ExecStart` 必须指向你每次发布后真正会被覆盖更新的二进制
- 如果你用的是 `scripts/build-cn.sh`，默认会更新 `/srv/crm-go-vue-shadcn/backend/crm-backend`
- 如果你用的是 `scripts/package-release.sh`，默认产物在 `release/crm-backend`，要么手动复制过去，要么把 `ExecStart` 改成 `release` 目录里的实际路径

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

    location /check-in/ {
        alias /srv/crm-go-vue-shadcn/check-in/;
        try_files $uri $uri/ /check-in/index.html;
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
curl -s http://127.0.0.1:8080/api/health
```

说明：

- `scripts/build-cn.sh` 和 `scripts/package-release.sh` 会把 `version`、`git_commit`、`build_time` 写进后端二进制
- 重启后访问 `/api/health`，应该能看到当前运行中的构建版本和服务启动时间

### 9.1 One-Click Deploy For Current Overseas Server

如果你线上目录结构是下面这样：

```text
/home/shipin/crm-go.zhaozhaoying.cn
├── .env
├── dist/
├── check-in/
└── overseas_linux
```

并且 `systemd` 服务名是 `crm-go`，可以直接使用仓库里的脚本：

```bash
cd /Users/zhangyang/dev/zhaozhaoying/crm-go-vue-shadcn
bash ./scripts/deploy-overseas.sh
```

这个脚本会自动完成：

- 使用 `scripts/package-release.sh` 打包
- 生成 `release/dist`、`release/check-in` 和 `release/overseas_linux`
- 默认不会把本地 `.env` 打进 `release/`
- 上传前端 `dist`
- 上传 `check-in` H5 Web 静态资源
- 使用 `rsync` 上传后端二进制到远端唯一临时文件
- 在服务器上先停止并清理 `crm-go` 的失败状态，再原子替换 `overseas_linux`
- 启动 `crm-go`
- 如果启动失败，自动打印 `systemctl status` 和 `journalctl` 诊断信息
- 打印 `stat` 和 `/api/health` 结果，确认线上是否已经切到新版本

常用参数：

```bash
# 只有明确需要时，才把指定 env 文件上传到线上 .env
DEPLOY_ENV_FILE=./backend/.env.staging bash ./scripts/deploy-overseas.sh --with-env

# 先清理 Go 构建缓存，再重新打包 overseas_linux
bash ./scripts/deploy-overseas.sh --clean-build

# 只发前端
bash ./scripts/deploy-overseas.sh --frontend-only

# 只发后端
bash ./scripts/deploy-overseas.sh --backend-only

# 不重新打包，直接使用现有 release/ 产物
bash ./scripts/deploy-overseas.sh --skip-build

# 只上传，不重启远端 systemd 服务
bash ./scripts/deploy-overseas.sh --no-restart
```

默认值和覆盖方式：

- 默认 SSH 目标：`zhangyang@192.155.80.209`
- 默认 SSH 私钥：`/Users/zhangyang/dev/zhaozhaoying/jiaoben/KeyPairForZhangYang.pem`
- 默认远端目录：`/home/shipin/crm-go.zhaozhaoying.cn`
- 默认服务名：`crm-go`
- 默认后端文件名：`overseas_linux`
- 默认不会覆盖线上 `.env`
- 如需显式上传环境变量文件：`DEPLOY_ENV_FILE=/path/to/file bash ./scripts/deploy-overseas.sh --with-env`
- 默认构建目标：`GOOS=linux GOARCH=amd64 CGO_ENABLED=0`
- 可选彻底重编译：`--clean-build`，会先执行 `go clean -cache -testcache`
- 默认服务切换方式：执行时提示输入远端 `sudo` 密码，先停止并清理失败状态，再启动 `crm-go`

如果要覆盖默认值，可以这样传环境变量：

```bash
DEPLOY_SSH_TARGET=root@your-server \
DEPLOY_SSH_KEY=~/.ssh/id_rsa \
DEPLOY_REMOTE_DIR=/home/shipin/your-app \
DEPLOY_REMOTE_SERVICE=crm-go \
DEPLOY_BACKEND_BIN_NAME=overseas_linux \
DEPLOY_CLEAN_BUILD=1 \
DEPLOY_SUDO_PASSWORD='你的sudo密码' \
bash ./scripts/deploy-overseas.sh
```

### 10. Deployment Checklist

上线前至少确认这些：

- 后端健康检查通过：`http://127.0.0.1:8080/api/health`
- 健康检查返回的 `gitCommit`、`buildTime`、`startedAt` 符合本次发布
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
- 服务看起来重启了，但接口还是旧逻辑
  - 执行 `systemctl show crm-backend -p ExecStart,WorkingDirectory`
  - 执行 `readlink -f /proc/$(pgrep -f crm-backend | head -n 1)/exe`
  - 执行 `curl -s http://127.0.0.1:8080/api/health`
  - 确认 `ExecStart` 指向的二进制、进程实际运行的二进制、以及这次构建输出的二进制是同一个文件
- 合同开始时间/结束时间没有生成
  - 确认后端服务已经重启到最新代码
  - 重新保存一次合同，让后端执行最新提交逻辑

### 12. Suggested Next Step

如果你准备正式长期运维，建议下一步补这几样：

- `Dockerfile` 和 `docker-compose.yml`
- 生产环境备份脚本
- CI/CD 自动构建和发布流程
