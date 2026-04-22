# 项目目录结构重构方案

> 目标：把当前 `backend / frontend / check-in / mihua-token-fetcher` 这种四个并列子项目的“多项目仓”形态，整理成一个**以 Go 后端为主、前端及附属系统按职责清晰挂载**的生产级单体仓 (mono-repo) 布局。

---

## 1. 当前结构存在的问题

当前根目录：

```text
crm-go-vue-shadcn/
├── backend/                  # Go 后端（Gin + GORM）
├── frontend/                 # Vue3 + Vite + shadcn-vue 主前端
├── check-in/                 # uni-app H5 签到系统
├── mihua-token-fetcher/      # Python 抓取秘话 token 的小工具
├── scripts/                  # 部署 / 打包脚本
├── release/                  # 打包产物
├── tmp/
└── README.md
```

主要问题：

1. **Go 后端被埋在 `backend/` 子目录**：不符合 Go 项目惯例（`go.mod` 通常在仓库根），导致 `go build ./...`、`go install`、IDE 跳转、CI 缓存路径都需要额外加 `cd backend`，很多 Go 工具链的默认行为也无法直接用。
2. **四个子项目并列**，从命名上看不出主次关系，新人接手时不容易第一眼看出"这是一个 Go 后端项目，附带一个主前端、一个 H5 签到系统、一个 Python 工具"。
3. **`check-in` / `mihua-token-fetcher` 命名不规范**：
   - `check-in` 太通用，不知道是签到的什么系统，给谁用
   - `mihua-token-fetcher` 是中文拼音 + 英文混拼，且作为目录名过长
4. **`frontend/` 命名过于泛化**：当仓库内同时存在多个"前端形态"（主后台、H5、未来可能的小程序）时，`frontend` 这个名字不够具体。
5. **`release/`、`tmp/` 这类产物目录直接散落在根目录**，且没有统一的 `build/` 或 `deploy/` 概念。

---

## 2. 目标结构（推荐方案）

按你的要求：**项目根目录就是 Go 后端**，主前端放在 `web/`，`check-in` 与 `mihua-token-fetcher` 重命名。

```text
crm-go-vue-shadcn/                         # 仓库根 = Go 后端模块根（go.mod 在这里）
│
├── go.mod
├── go.sum
├── main.go
├── README.md
├── LICENSE
├── .env.example                           # 后端环境变量样例（原 backend/production.env.sample）
├── .gitignore
├── .editorconfig
│
├── cmd/                                   # 各个可执行入口（原 backend/cmd/）
│   ├── server/                            # 主服务进程（也可以保留根 main.go，二选一）
│   ├── bootstrap/
│   ├── migrate/
│   ├── import-customers/
│   ├── sqlite2mysql/
│   ├── sync-spxxjj-telemarketing/
│   ├── google-search-check/
│   ├── mihua-call-record-check/
│   ├── hanghang-inspect-temp/
│   └── hanghang-verify-temp/
│
├── internal/                              # Go 内部包（原 backend/internal/）
│   ├── authctx/
│   ├── config/
│   ├── database/
│   ├── errmsg/
│   ├── external/
│   ├── handler/
│   ├── middleware/
│   ├── model/
│   ├── repository/
│   ├── router/
│   ├── scoring/
│   ├── service/
│   └── util/
│
├── docs/                                  # 后端 / 项目文档（原 backend/docs/ 合并 + 本方案）
│   ├── api/
│   ├── deploy/
│   └── RESTRUCTURE_PROPOSAL.md            # 即本文件
│
├── web/                                   # 主后台前端（原 frontend/）
│   ├── src/
│   ├── index.html
│   ├── package.json
│   ├── pnpm-lock.yaml
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   └── components.json
│
├── apps/                                  # 子应用集合（独立可部署的非主前端）
│   ├── checkin-h5/                        # 原 check-in/，uni-app H5 签到
│   │   ├── src/
│   │   ├── pages/
│   │   ├── static/
│   │   ├── api/
│   │   ├── utils/
│   │   ├── App.vue
│   │   ├── main.js
│   │   ├── pages.json
│   │   ├── manifest.json
│   │   └── package.json
│   │
│   └── mihua-token-service/               # 原 mihua-token-fetcher/，Python token 服务
│       ├── src/
│       ├── tests/
│       ├── pyproject.toml
│       └── README.md
│
├── scripts/                               # 构建 / 部署脚本（保留）
│   ├── build.sh                           # 由原 build-cn.sh 改名（或保留中国镜像版）
│   ├── package-release.sh
│   └── deploy-overseas.sh
│
├── deploy/                                # 部署相关声明式资源
│   ├── systemd/
│   │   └── crm-backend.service
│   ├── nginx/
│   │   └── crm.conf
│   └── docker/                            # 后续可加 Dockerfile / compose
│
├── build/                                 # 构建产物（取代根目录散乱的 release/ 和 tmp/）
│   └── release/                           # 打包脚本输出位置
│       ├── crm-backend                    # Go 二进制
│       ├── web/                           # 主前端 dist
│       └── checkin-h5/                    # H5 静态产物
│
└── .github/                               # 后续上 CI 时使用
    └── workflows/
```

> ✅ 这个结构同时满足：
> - 根目录是 Go 后端（`go.mod` 在根，`go build ./...` 可直接用）
> - `web/` 是主前端
> - 原 `check-in` → `apps/checkin-h5/`
> - 原 `mihua-token-fetcher` → `apps/mihua-token-service/`
> - 部署脚本、构建产物、文档都各归其位

---

## 3. 重命名建议详解

### 3.1 `frontend/` → `web/`

**理由**：

- `web` 是 Go 生态里对"主 Web 前端"很通用的命名（如 Buffalo、Echo 模板、众多开源 CRM 都用 `web/`）。
- 与 `apps/checkin-h5/` 形成清晰的层级关系：**主后台**叫 `web`，**附属应用**放在 `apps/`。
- 比 `frontend` 更短，且与 `internal`、`cmd` 同为单字目录，整体观感整齐。

### 3.2 `check-in/` → `apps/checkin-h5/`

**理由**：

- `check-in` 太泛，作为顶层目录会让人误以为是"签到 SDK"或"签到模块"
- 加 `-h5` 后缀显式表明这是 **H5 形态的签到端**（区别于未来可能的小程序版、原生 App 版）
- 放在 `apps/` 下表明它是**独立可部署的子应用**，不属于主前端 `web/`

可选命名：

| 候选名 | 说明 |
|---|---|
| `apps/checkin-h5/` ✅ 推荐 | 直观、形态清晰 |
| `apps/attendance-h5/` | 如果是员工考勤场景更准确 |
| `apps/visitor-checkin/` | 如果是访客 / 客户签到场景 |

> 👉 请根据实际业务（员工考勤 / 客户签到 / 活动签到）二选一，本方案默认用 `checkin-h5`。

### 3.3 `mihua-token-fetcher/` → `apps/mihua-token-service/`

**理由**：

- "fetcher" 给人感觉是**一次性脚本**，但从 `pyproject.toml` + `src/` + `tests/` 的结构看，它实际上是一个**长期运行的 Python 服务**
- 改成 `-service` 更能体现它的长期角色
- 放在 `apps/` 下与 H5 平级，统一作为"非主前端的附属应用"

可选命名：

| 候选名 | 说明 |
|---|---|
| `apps/mihua-token-service/` ✅ 推荐 | 强调是后台服务 |
| `apps/mihua-token/` | 更简洁 |
| `apps/mihua-auth/` | 如果它的核心职责是鉴权 / 登录态维护 |

> 👉 如果它**只是定时拉一次 token 入库**，那保留 `-fetcher` 也合理；如果它**常驻且对外暴露接口**，建议 `-service`。

### 3.4 `release/` + `tmp/` → `build/release/`

把所有"构建产物"统一收到 `build/` 下，避免根目录被各种编译临时文件污染；同时 `build/` 加入 `.gitignore`。

---

## 4. 迁移步骤（Step-by-Step）

> 建议在一个独立的 `chore/restructure-layout` 分支里完成，分阶段提交，每阶段都跑一次 build 验证。

### Step 1：备份与分支

```bash
cd /Users/zhangyang/dev/zhaozhaoying/crm-go-vue-shadcn
git status                                # 确认无未提交改动
git checkout -b chore/restructure-layout
```

### Step 2：把 `backend/` 内容上提到仓库根

```bash
# 1) 把 backend 内的所有内容（含隐藏文件）移到根
git mv backend/cmd ./cmd
git mv backend/internal ./internal
git mv backend/docs ./docs                # 若根已存在 docs/，改用 rsync 合并后再 git add
git mv backend/main.go ./main.go
git mv backend/go.mod ./go.mod
git mv backend/go.sum ./go.sum
git mv backend/README.md ./docs/backend.md
git mv backend/production.env.sample ./.env.example

# 2) data.db 不应入库，先移走再补 .gitignore
mv backend/data.db ./build/local-data.db   # 或直接删除
echo "build/" >> .gitignore
echo "*.db"   >> .gitignore

# 3) 删除空的 backend 目录
rmdir backend
```

⚠️ Go 模块路径检查：

- 打开根 `go.mod`，确认 `module` 行（例如 `module crm-go-vue-shadcn`）。
- `internal/` 包之间的 import 路径**不需要改**（因为 module 名没变，相对路径 `crm-go-vue-shadcn/internal/...` 仍然有效）。

### Step 3：重命名前端目录

```bash
git mv frontend web
```

### Step 4：建立 `apps/` 并迁入两个附属应用

```bash
mkdir apps
git mv check-in              apps/checkin-h5
git mv mihua-token-fetcher   apps/mihua-token-service
```

### Step 5：整理 `release/` 与 `tmp/`

```bash
mkdir -p build
git mv release build/release
rm -rf tmp                                # tmp 通常是本地缓存，不应入库
```

### Step 6：把部署声明式文件抽到 `deploy/`

如果你之后想把 `systemd/crm-backend.service`、`nginx/crm.conf` 这类配置纳入仓库管理：

```bash
mkdir -p deploy/systemd deploy/nginx deploy/docker
# 把现有线上的 service / nginx 配置 copy 进来
```

### Step 7：批量修改受影响的脚本与文档

需要全局替换的字符串（用编辑器全局搜索一次即可）：

| 原路径 | 新路径 |
|---|---|
| `backend/`         | （删除前缀，直接根路径）|
| `frontend/`        | `web/` |
| `check-in/`        | `apps/checkin-h5/` |
| `mihua-token-fetcher/` | `apps/mihua-token-service/` |
| `release/`         | `build/release/` |

涉及的文件至少包括：

- `scripts/build-cn.sh`（如果存在）
- `scripts/package-release.sh`
  - L7  `CHECKIN_WEB_SRC_DIR="${ROOT_DIR}/check-in/unpackage/dist/build/web"` → `${ROOT_DIR}/apps/checkin-h5/unpackage/dist/build/web`
  - L8  `CHECKIN_WEB_OUT_DIR="${OUT_DIR}/check-in"` → `${OUT_DIR}/checkin-h5`
  - L50 `cd "$ROOT_DIR/backend"` → `cd "$ROOT_DIR"`
  - L63/65 `cd "$ROOT_DIR/frontend"` → `cd "$ROOT_DIR/web"`
  - L73/76 `frontend/dist` → `web/dist`
- `scripts/deploy-overseas.sh`
  - L32 `${ROOT_DIR}/backend/.env` → `${ROOT_DIR}/.env`
  - L197 `${LOCAL_RELEASE_DIR}/check-in/` → `${LOCAL_RELEASE_DIR}/checkin-h5/`
  - 其他 `check-in` 字面量按上表替换
- `README.md` 中所有 `cd backend` / `cd frontend` 和示例目录树

### Step 8：构建验证

```bash
# 1. Go 后端：直接在根目录跑
go mod tidy
go build ./...
go run ./cmd/migrate up
go run ./cmd/bootstrap

# 2. 主前端
cd web
pnpm install && pnpm build
cd ..

# 3. H5 签到
cd apps/checkin-h5
npm install && npm run build:h5         # 或现有命令
cd ../..

# 4. Python token 服务
cd apps/mihua-token-service
uv sync                                 # 或 pip install -e .
pytest
cd ../..

# 5. 整体打包
bash ./scripts/package-release.sh
ls build/release/                       # 应当看到 crm-backend / web/dist / checkin-h5/
```

### Step 9：更新 IDE / CI / 部署配置

| 项目 | 需要改的地方 |
|---|---|
| 本地 IDE  | Go 项目根改为仓库根；前端项目根改为 `web/` |
| systemd  | `WorkingDirectory=/srv/crm-go-vue-shadcn`（去掉 `/backend`）<br>`ExecStart=/srv/crm-go-vue-shadcn/crm-backend` |
| Nginx    | `root /srv/crm-go-vue-shadcn/web/dist;`<br>`/check-in/` 的 alias 改成 `apps/checkin-h5/unpackage/dist/build/web/`（或继续用 release 后的路径）|
| CI/CD    | 缓存 key、构建路径全部去掉 `backend/` 前缀 |

### Step 10：合并与发布

```bash
git add -A
git commit -m "chore: restructure repo layout — backend at root, web/, apps/checkin-h5, apps/mihua-token-service"
git push origin chore/restructure-layout
# PR review → merge → 在测试环境完整跑一次 deploy-overseas.sh 验证
```

---

## 5. 重构前后对照表

| 类别 | 重构前 | 重构后 |
|---|---|---|
| Go 后端模块根 | `backend/` | **仓库根** |
| Go 入口 | `backend/main.go` | `main.go` |
| Go 子命令 | `backend/cmd/*` | `cmd/*` |
| Go 内部包 | `backend/internal/*` | `internal/*` |
| 主前端 | `frontend/` | `web/` |
| H5 签到 | `check-in/` | `apps/checkin-h5/` |
| Python token 服务 | `mihua-token-fetcher/` | `apps/mihua-token-service/` |
| 后端环境变量样例 | `backend/production.env.sample` | `.env.example` |
| 后端文档 | `backend/docs/` | `docs/` |
| 部署声明式资源 | （散落在 README） | `deploy/{systemd,nginx,docker}/` |
| 构建产物 | `release/`、`tmp/` | `build/release/` |
| 部署脚本 | `scripts/` | `scripts/`（保留，仅改路径引用）|

---

## 6. 兼容性 / 风险提示

1. **systemd 与 Nginx 路径必须同步更新**，否则线上重启会找不到二进制 / 静态文件。建议先在测试机完整跑一遍再上生产。
2. **历史 git blame** 会因为 `git mv` 跨目录受到一定影响，但因为是整个目录搬迁，`--follow` 仍可追踪。
3. **CI 缓存路径**（如果有）需要重建一次，第一次构建会变慢。
4. **本地 `data.db` 不应该再放进仓库**，迁移时顺手清理；如果有人曾把它 commit 过，建议用 `git rm --cached data.db` + `.gitignore` 处理。
5. **Go module 名是否要改**？目前不建议改 `module` 名，避免所有 import 路径全部要改，迁移成本最低。
6. **uni-app `unpackage/`** 通常是构建产物，建议加入 `.gitignore`，不要随仓库带走。

---

## 7. 后续优化建议（非本次必须）

- 加 `Makefile` 或 `Taskfile.yml`，把 `make build / make dev / make release` 三个常用命令固化下来
- 加根级 `Dockerfile` + `docker-compose.yml`，本地一条命令拉起 MySQL + 后端 + 前端
- `deploy/` 里补 GitHub Actions workflow，PR 自动跑 `go vet / golangci-lint / pnpm build`
- 把 `apps/mihua-token-service` 的运行方式（systemd? cron? 容器?）也写进 `docs/deploy/`
- 给 `apps/checkin-h5` 单独写一份 `README.md`，说明它是"独立部署在 `/check-in/` 子路径下"

---

## 8. TL;DR

**改三件事**：

1. **`backend/` 内容全部上提到仓库根**，让根目录就是 Go 模块根。
2. **`frontend/` → `web/`**。
3. **新建 `apps/`**，把 `check-in/` → `apps/checkin-h5/`，`mihua-token-fetcher/` → `apps/mihua-token-service/`。

**附带收益**：

- `release/` + `tmp/` 收敛到 `build/release/`
- `deploy/` 集中存放 systemd / nginx / docker 配置
- 部署脚本只需做"前缀替换"，不需要改逻辑
