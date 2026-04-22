# mihua-token-service

一个放在当前仓库里的小型 Python 工具，用米话接口直连登录或浏览器监听的方式获取 `token`，并可选回写到仓库根目录 `.env` 的 `MIHUA_CALL_RECORD_TOKEN`。

当前目录名已调整为 `apps/mihua-token-service/`，CLI 命令名仍然保持 `mihua-token-fetcher`，以兼容现有使用方式。

## 功能

- 打开米话页面后监听请求头里的 `token`
- 支持直接请求米话 `captcha` 和 `user/token` 接口获取 `token`
- 验证码优先使用 `ddddocr` 识别，缺失时自动降级到 `tesseract`
- 直连 API 失败后会自动回退到浏览器自动登录继续获取 `token`
- 尝试从登录响应 JSON、`localStorage`、`sessionStorage`、cookie 里补充提取
- 默认优先读取仓库根目录 `.env`
- 可选把最新 token 写回仓库根目录 `.env`
- 可选按仓库根目录 `.env` 的数据库配置，把 token 回写到 `system_settings.mihua_call_record_token`

## 目录

```text
apps/mihua-token-service
├── pyproject.toml
├── README.md
├── src/mihua_token_fetcher
│   ├── __init__.py
│   ├── __main__.py
│   ├── cli.py
│   └── helpers.py
└── tests
    └── test_helpers.py
```

## 快速开始

```bash
cd "apps/mihua-token-service"
python3 -m venv .venv
source ".venv/bin/activate"
pip install -e . --no-build-isolation
python3 -m playwright install chromium
```

如果你要启用本地 OCR，建议一并确认：

```bash
python3 -c "import ddddocr; print('ddddocr ok')"
```

直接运行：

```bash
cd "apps/mihua-token-service"
source ".venv/bin/activate"
mihua-token-fetcher
```

或者：

```bash
cd "apps/mihua-token-service"
source ".venv/bin/activate"
python3 -m mihua_token_fetcher
```

## 常用参数

```bash
mihua-token-fetcher --write-env
mihua-token-fetcher --origin "https://spxxjj.emicloudcc.com"
mihua-token-fetcher --open-url "https://spxxjj.emicloudcc.com"
mihua-token-fetcher --timeout-seconds 300
mihua-token-fetcher --headless
mihua-token-fetcher --username "你的账号" --password "你的密码" --write-env
mihua-token-fetcher --username "你的账号" --password "你的密码" --captcha "Ab12" --write-env
mihua-token-fetcher --write-db
mihua-token-fetcher --write-db --write-env --headless
```

## 使用流程

1. 如果传了 `--username` 和 `--password`，工具会先尝试直连 API 自动登录。
2. API 未拿到 `token` 时，会自动回退到浏览器模式，继续自动识别验证码并提交登录。
3. 工具监听到请求头、响应体或页面存储中的 `token` 后会立即输出。
4. 如果传了 `--write-env`，会同步更新仓库根目录 `.env` 里的 `MIHUA_CALL_RECORD_TOKEN`。
5. 如果传了 `--write-db`，会按仓库根目录 `.env` 的数据库配置更新 `system_settings.key = mihua_call_record_token`。

## 直连 API 模式

如果传入 `--username` 和 `--password`，工具默认优先走直连 API：

- 请求 `https://cmb.emicloudcc.com/captcha` 获取 `session_id` 和验证码图片
- 优先使用 `--captcha`，否则优先调用 `ddddocr`，缺失时再尝试本机 `tesseract`
- 根据米话前端签名规则请求 `https://cmb.emicloudcc.com/user/token`
- 成功后直接拿到 `token`
- 如果连续重试仍失败，会自动回退到浏览器模式继续闭环获取

示例：

```bash
cd "apps/mihua-token-service"
source ".venv/bin/activate"
mihua-token-fetcher \
  --origin "https://spxxjj.emicloudcc.com" \
  --username "你的账号" \
  --password "你的密码" \
  --captcha "Ab12" \
  --write-env
```

如果不传 `--captcha` 且 OCR 失败，工具会把验证码图片保存到临时文件，并在下一轮自动刷新验证码继续尝试。

## 浏览器模式

如果传入 `--username` 和 `--password`，工具会：

- 自动填写用户名和密码
- 读取同一页面里的验证码图片并保存到临时文件
- 优先使用 `ddddocr`，缺失时回退到 `tesseract`
- 自动提交登录并继续抓取 token
- API 失败时会自动进入此模式，无需手动重跑命令

示例：

```bash
cd "apps/mihua-token-service"
source ".venv/bin/activate"
mihua-token-fetcher \
  --force-browser \
  --open-url "https://spxxjj.emicloudcc.com/#/login" \
  --origin "https://spxxjj.emicloudcc.com" \
  --username "你的账号" \
  --password "你的密码" \
  --write-env
```

如果 OCR 失败，可以额外手动传：

```bash
mihua-token-fetcher --username "你的账号" --password "你的密码" --captcha "Ab12" --write-env
```

## 注意

- 默认情况下，工具会尝试从仓库根目录的 `.env` 读取：
  - `MIHUA_CALL_RECORD_SOURCE_ORIGIN`
  - `MIHUA_CALL_RECORD_LIST_URL`
  - `MIHUA_TELEMARKETING_RECORDING_LIST_URL`
  - `MIHUA_TELEMARKETING_RECORDING_DETAIL_URL`
- 这些值只用于定位米话站点和过滤目标请求，不会自动提交到 Git。
- 仓库根目录 `.env` 一般不会提交到 Git，但仍然建议不要把输出的 token 发到公共渠道。
- 如果你是在工具目录自己的 `.venv` 里安装了 `ddddocr`，脚本也会优先尝试从该 `.venv` 里加载它。
- 如果 `DB_DRIVER=mysql`，请确保虚拟环境里已安装 `PyMySQL`；使用 `pip install -e . --no-build-isolation` 即可一起装上。

## 宝塔定时刷新

如果你要在宝塔里每 3 小时刷新一次，推荐直接配置 Shell 任务：

```bash
cd "/www/wwwroot/crm-go-shadcn/apps/mihua-token-service" && .venv/bin/mihua-token-fetcher --env-file "/www/wwwroot/crm-go-shadcn/.env" --headless --write-db
```

如果你还想保留 `.env` 里的副本：

```bash
cd "/www/wwwroot/crm-go-shadcn/apps/mihua-token-service" && .venv/bin/mihua-token-fetcher --env-file "/www/wwwroot/crm-go-shadcn/.env" --headless --write-db --write-env
```

宝塔的 cron 表达式用：

```text
0 */3 * * *
```

前提：

- `/www/wwwroot/crm-go-shadcn/.env` 里已经配置好 `DB_DRIVER`、`MYSQL_*` 或 `DB_PATH`
- `.venv` 已经执行过 `pip install -e . --no-build-isolation`
- `playwright` 浏览器依赖已安装完成
