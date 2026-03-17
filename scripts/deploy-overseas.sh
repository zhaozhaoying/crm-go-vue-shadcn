#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# 海外服务器一键发布脚本
#
# 这份脚本默认适配当前项目的线上结构：
# 1. 前端静态文件发布到 /home/shipin/crm-go.zhaozhaoying.cn/dist/
# 2. 后端二进制发布为 /home/shipin/crm-go.zhaozhaoying.cn/overseas_linux
# 3. 远端 systemd 服务名为 crm-go
# 4. 通过 zhangyang 用户 + SSH 私钥连接服务器
# 5. 通过 sudo 密码提权重启 systemctl 服务
#
# 为了避免把 sudo 密码硬编码到仓库里：
# - 默认在执行时安全输入密码
# - 或者通过环境变量 DEPLOY_SUDO_PASSWORD 传入
#
# 常见用法：
#   bash ./scripts/deploy-overseas.sh
#   bash ./scripts/deploy-overseas.sh --clean-build
#   bash ./scripts/deploy-overseas.sh --with-env
#   bash ./scripts/deploy-overseas.sh --frontend-only
#   bash ./scripts/deploy-overseas.sh --backend-only
#   bash ./scripts/deploy-overseas.sh --no-restart
# ============================================================

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# ----------------------------
# 远端连接与部署目标默认值
# ----------------------------
DEPLOY_SSH_TARGET="${DEPLOY_SSH_TARGET:-zhangyang@192.155.80.209}"
DEPLOY_SSH_KEY="${DEPLOY_SSH_KEY:-/Users/zhangyang/dev/zhaozhaoying/jiaoben/KeyPairForZhangYang.pem}"
DEPLOY_REMOTE_DIR="${DEPLOY_REMOTE_DIR:-/home/shipin/crm-go.zhaozhaoying.cn}"
DEPLOY_REMOTE_SERVICE="${DEPLOY_REMOTE_SERVICE:-crm-go}"
DEPLOY_REMOTE_HEALTH_URL="${DEPLOY_REMOTE_HEALTH_URL:-http://127.0.0.1:8080/api/health}"

# ----------------------------
# 本地产物与构建参数默认值
# ----------------------------
DEPLOY_RELEASE_DIR="${DEPLOY_RELEASE_DIR:-release}"
DEPLOY_BACKEND_BIN_NAME="${DEPLOY_BACKEND_BIN_NAME:-overseas_linux}"
GOOS_TARGET="${GOOS_TARGET:-linux}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"
CGO_ENABLED_TARGET="${CGO_ENABLED_TARGET:-0}"

# ----------------------------
# 可选部署行为
# ----------------------------
DEPLOY_SYNC_ENV="${DEPLOY_SYNC_ENV:-0}"
DEPLOY_RESTART_SERVICE="${DEPLOY_RESTART_SERVICE:-1}"
DEPLOY_SUDO_PASSWORD="${DEPLOY_SUDO_PASSWORD:-}"
DEPLOY_CLEAN_BUILD="${DEPLOY_CLEAN_BUILD:-0}"

# ----------------------------
# 运行时开关
# ----------------------------
SKIP_BUILD=0
DEPLOY_FRONTEND=1
DEPLOY_BACKEND=1

usage() {
  cat <<'EOF'
用法：
  bash ./scripts/deploy-overseas.sh [选项]

选项：
  --with-env       连同 release/.env 一起上传到远端 .env
  --clean-build    先清理 Go 构建缓存，再重新打包 overseas_linux
  --skip-build     跳过本地打包，直接复用已有 release/ 产物
  --frontend-only  只发布前端 dist
  --backend-only   只发布后端二进制
  --no-restart     只上传文件，不重启远端 systemd 服务
  -h, --help       查看帮助

可覆盖的环境变量：
  DEPLOY_SSH_TARGET        默认: zhangyang@192.155.80.209
  DEPLOY_SSH_KEY           默认: /Users/zhangyang/dev/zhaozhaoying/jiaoben/KeyPairForZhangYang.pem
  DEPLOY_REMOTE_DIR        默认: /home/shipin/crm-go.zhaozhaoying.cn
  DEPLOY_REMOTE_SERVICE    默认: crm-go
  DEPLOY_REMOTE_HEALTH_URL 默认: http://127.0.0.1:8080/api/health
  DEPLOY_RELEASE_DIR       默认: release
  DEPLOY_BACKEND_BIN_NAME  默认: overseas_linux
  GOOS_TARGET              默认: linux
  GOARCH_TARGET            默认: amd64
  CGO_ENABLED_TARGET       默认: 0
  DEPLOY_SYNC_ENV          默认: 0
  DEPLOY_RESTART_SERVICE   默认: 1
  DEPLOY_SUDO_PASSWORD     默认: 空。若未提供且需要重启服务，会在执行时提示安全输入
  DEPLOY_CLEAN_BUILD       默认: 0。为 1 时先执行 go clean -cache -testcache

示例：
  bash ./scripts/deploy-overseas.sh
  bash ./scripts/deploy-overseas.sh --clean-build
  bash ./scripts/deploy-overseas.sh --with-env
  bash ./scripts/deploy-overseas.sh --frontend-only
  DEPLOY_SUDO_PASSWORD='你的sudo密码' bash ./scripts/deploy-overseas.sh
  DEPLOY_SSH_TARGET=root@your-server DEPLOY_SSH_KEY=~/.ssh/id_rsa bash ./scripts/deploy-overseas.sh
EOF
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "缺少必要命令: $1" >&2
    exit 1
  fi
}

need_service_restart() {
  if [ "$DEPLOY_RESTART_SERVICE" != "1" ]; then
    return 1
  fi
  if [ "$DEPLOY_BACKEND" = "1" ] || [ "$DEPLOY_SYNC_ENV" = "1" ]; then
    return 0
  fi
  return 1
}

prompt_sudo_password_if_needed() {
  if ! need_service_restart; then
    return 0
  fi

  if [ -n "$DEPLOY_SUDO_PASSWORD" ]; then
    return 0
  fi

  # 这里用静默输入，避免密码回显到终端。
  read -r -s -p "请输入远端 sudo 密码（不会回显）: " DEPLOY_SUDO_PASSWORD
  echo

  if [ -z "$DEPLOY_SUDO_PASSWORD" ]; then
    echo "sudo 密码不能为空" >&2
    exit 1
  fi
}

while [ $# -gt 0 ]; do
  case "$1" in
    --with-env)
      DEPLOY_SYNC_ENV=1
      ;;
    --clean-build)
      DEPLOY_CLEAN_BUILD=1
      ;;
    --skip-build)
      SKIP_BUILD=1
      ;;
    --frontend-only)
      DEPLOY_BACKEND=0
      ;;
    --backend-only)
      DEPLOY_FRONTEND=0
      ;;
    --no-restart)
      DEPLOY_RESTART_SERVICE=0
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "未知参数: $1" >&2
      usage
      exit 1
      ;;
  esac
  shift
done

if [ "$DEPLOY_FRONTEND" -eq 0 ] && [ "$DEPLOY_BACKEND" -eq 0 ]; then
  echo "没有可发布的内容：前端和后端都被关闭了" >&2
  exit 1
fi

require_command bash
require_command ssh
require_command scp
require_command rsync
require_command stat
require_command base64

if [ ! -f "$DEPLOY_SSH_KEY" ]; then
  echo "SSH 私钥不存在: $DEPLOY_SSH_KEY" >&2
  exit 1
fi

prompt_sudo_password_if_needed

LOCAL_RELEASE_DIR="${ROOT_DIR}/${DEPLOY_RELEASE_DIR}"
LOCAL_DIST_DIR="${LOCAL_RELEASE_DIR}/dist/"
LOCAL_BIN_PATH="${LOCAL_RELEASE_DIR}/${DEPLOY_BACKEND_BIN_NAME}"
LOCAL_ENV_PATH="${LOCAL_RELEASE_DIR}/.env"
REMOTE_BIN_PATH="${DEPLOY_REMOTE_DIR}/${DEPLOY_BACKEND_BIN_NAME}"
REMOTE_BIN_TMP_PATH="${REMOTE_BIN_PATH}.new"

# 这里统一封装 SSH / SCP / RSYNC 的连接参数，避免多处重复。
RSYNC_SSH_COMMAND="$(printf 'ssh -i %q -o StrictHostKeyChecking=accept-new' "$DEPLOY_SSH_KEY")"
SSH_CMD=(ssh -i "$DEPLOY_SSH_KEY" -o StrictHostKeyChecking=accept-new)
SCP_CMD=(scp -i "$DEPLOY_SSH_KEY" -o StrictHostKeyChecking=accept-new)

# 如果需要远端 sudo，就把密码做一次 base64 编码再传过去。
# 原因：
# 1. 可以避免因为特殊字符导致远端 shell 引号转义很难写
# 2. 避免明文直接拼到 ssh 命令里
# 3. 不是加密，只是为了更稳地传输
DEPLOY_SUDO_PASSWORD_B64=""
if need_service_restart; then
  DEPLOY_SUDO_PASSWORD_B64="$(printf '%s' "$DEPLOY_SUDO_PASSWORD" | base64 | tr -d '\n')"
fi

echo "==> 发布配置"
echo "ssh_target=${DEPLOY_SSH_TARGET}"
echo "ssh_key=${DEPLOY_SSH_KEY}"
echo "remote_dir=${DEPLOY_REMOTE_DIR}"
echo "remote_service=${DEPLOY_REMOTE_SERVICE}"
echo "backend_bin=${DEPLOY_BACKEND_BIN_NAME}"
echo "deploy_frontend=${DEPLOY_FRONTEND}"
echo "deploy_backend=${DEPLOY_BACKEND}"
echo "deploy_env=${DEPLOY_SYNC_ENV}"
echo "clean_build=${DEPLOY_CLEAN_BUILD}"
echo "restart_service=${DEPLOY_RESTART_SERVICE}"
echo "release_dir=${LOCAL_RELEASE_DIR}"

if [ "$SKIP_BUILD" -eq 0 ]; then
  echo "==> 第一步：本地打包前后端产物"
  RELEASE_DIR="$DEPLOY_RELEASE_DIR" \
  BACKEND_BIN_NAME="$DEPLOY_BACKEND_BIN_NAME" \
  GOOS_TARGET="$GOOS_TARGET" \
  GOARCH_TARGET="$GOARCH_TARGET" \
  CGO_ENABLED_TARGET="$CGO_ENABLED_TARGET" \
  CLEAN_BUILD_TARGET="$DEPLOY_CLEAN_BUILD" \
    bash "$ROOT_DIR/scripts/package-release.sh"
fi

# 打包完成后，先确认本地产物是否真的存在。
if [ "$DEPLOY_FRONTEND" -eq 1 ] && [ ! -d "$LOCAL_DIST_DIR" ]; then
  echo "前端产物不存在: $LOCAL_DIST_DIR" >&2
  exit 1
fi

if [ "$DEPLOY_BACKEND" -eq 1 ] && [ ! -f "$LOCAL_BIN_PATH" ]; then
  echo "后端产物不存在: $LOCAL_BIN_PATH" >&2
  exit 1
fi

if [ "$DEPLOY_SYNC_ENV" = "1" ] && [ ! -f "$LOCAL_ENV_PATH" ]; then
  echo "环境变量文件不存在: $LOCAL_ENV_PATH" >&2
  echo "提示：如果要上传 .env，请先确认本地 backend/.env 存在，然后不要加 --skip-build" >&2
  exit 1
fi

echo "==> 第二步：预创建远端目录"
"${SSH_CMD[@]}" "$DEPLOY_SSH_TARGET" "mkdir -p '$DEPLOY_REMOTE_DIR' '$DEPLOY_REMOTE_DIR/dist'"

if [ "$DEPLOY_FRONTEND" -eq 1 ]; then
  echo "==> 第三步：上传前端 dist"
  # 这里使用 rsync --delete，保证远端 dist 和本地 release/dist 一致。
  rsync -az --delete -e "$RSYNC_SSH_COMMAND" "$LOCAL_DIST_DIR" "${DEPLOY_SSH_TARGET}:${DEPLOY_REMOTE_DIR}/dist/"
fi

if [ "$DEPLOY_BACKEND" -eq 1 ]; then
  echo "==> 第四步：上传后端二进制到临时文件"
  # 先上传成 overseas_linux.new，避免直接覆盖正在运行的文件。
  "${SCP_CMD[@]}" "$LOCAL_BIN_PATH" "${DEPLOY_SSH_TARGET}:${REMOTE_BIN_TMP_PATH}"
fi

if [ "$DEPLOY_SYNC_ENV" = "1" ]; then
  echo "==> 第五步：上传根目录 .env"
  "${SCP_CMD[@]}" "$LOCAL_ENV_PATH" "${DEPLOY_SSH_TARGET}:${DEPLOY_REMOTE_DIR}/.env"
fi

echo "==> 第六步：远端切换产物并按需重启服务"
"${SSH_CMD[@]}" "$DEPLOY_SSH_TARGET" \
  "DEPLOY_REMOTE_DIR='$DEPLOY_REMOTE_DIR' DEPLOY_REMOTE_SERVICE='$DEPLOY_REMOTE_SERVICE' DEPLOY_REMOTE_HEALTH_URL='$DEPLOY_REMOTE_HEALTH_URL' DEPLOY_RESTART_SERVICE='$DEPLOY_RESTART_SERVICE' DEPLOY_SUDO_PASSWORD_B64='$DEPLOY_SUDO_PASSWORD_B64' BACKEND_BIN_NAME='$DEPLOY_BACKEND_BIN_NAME' DEPLOY_BACKEND='$DEPLOY_BACKEND' DEPLOY_FRONTEND='$DEPLOY_FRONTEND' DEPLOY_SYNC_ENV='$DEPLOY_SYNC_ENV' bash -s" <<'EOF'
set -euo pipefail

remote_bin_path="${DEPLOY_REMOTE_DIR}/${BACKEND_BIN_NAME}"
remote_bin_tmp_path="${remote_bin_path}.new"

decode_sudo_password() {
  if [ -z "${DEPLOY_SUDO_PASSWORD_B64:-}" ]; then
    return 0
  fi
  printf '%s' "$DEPLOY_SUDO_PASSWORD_B64" | base64 --decode
}

run_sudo() {
  if [ -n "${DEPLOY_SUDO_PASSWORD_B64:-}" ]; then
    local password
    password="$(decode_sudo_password)"
    printf '%s\n' "$password" | sudo -S -p '' "$@"
  else
    sudo "$@"
  fi
}

if [ "$DEPLOY_BACKEND" = "1" ]; then
  # 先赋予执行权限，再原子替换正式二进制。
  chmod 755 "$remote_bin_tmp_path"
  mv "$remote_bin_tmp_path" "$remote_bin_path"
fi

if [ "$DEPLOY_RESTART_SERVICE" = "1" ] && { [ "$DEPLOY_BACKEND" = "1" ] || [ "$DEPLOY_SYNC_ENV" = "1" ]; }; then
  run_sudo systemctl restart "$DEPLOY_REMOTE_SERVICE"
  run_sudo systemctl status "$DEPLOY_REMOTE_SERVICE" --no-pager
fi

if [ "$DEPLOY_BACKEND" = "1" ]; then
  echo "==> 远端后端文件信息"
  stat "$remote_bin_path"
fi

if [ "$DEPLOY_FRONTEND" = "1" ]; then
  echo "==> 远端前端文件信息"
  stat "${DEPLOY_REMOTE_DIR}/dist/index.html"
fi

echo "==> 远端健康检查"
curl -fsS "$DEPLOY_REMOTE_HEALTH_URL"
EOF

echo "==> 发布完成"
