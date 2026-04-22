#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# release 产物准备脚本
#
# 当前脚本只负责把需要上线的内容纳入本地 build/release：
# 1. 后端二进制
# 2. 主前端 Web dist
# 3. attendance-h5 静态文件
# 4. mihua-token-fetcher 目录
#
# 不再执行任何远端上传、SSH、rsync、systemctl 重启动作。
# ============================================================

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEPLOY_RELEASE_DIR="${DEPLOY_RELEASE_DIR:-build/release}"
DEPLOY_BACKEND_BIN_NAME="${DEPLOY_BACKEND_BIN_NAME:-overseas_linux}"
GOOS_TARGET="${GOOS_TARGET:-linux}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"
CGO_ENABLED_TARGET="${CGO_ENABLED_TARGET:-0}"
DEPLOY_CLEAN_BUILD="${DEPLOY_CLEAN_BUILD:-0}"

SKIP_BUILD=0
INCLUDE_BACKEND=1
INCLUDE_WEB=1
INCLUDE_ATTENDANCE_H5=1
INCLUDE_MIHUA_TOKEN=1

usage() {
  cat <<'EOF'
用法：
  bash ./scripts/deploy.sh [选项]

说明：
  该脚本只负责准备本地 release 产物，不做上传。

选项：
  --clean-build    先清理 Go 构建缓存，再重新打包后端
  --skip-build     跳过本地打包，直接复用已有 build/release/ 产物
  --frontend-only  只保留主前端和 attendance-h5 到 release
  --backend-only   只保留后端二进制到 release
  -h, --help       查看帮助

可覆盖的环境变量：
  DEPLOY_RELEASE_DIR       默认: build/release
  DEPLOY_BACKEND_BIN_NAME  默认: overseas_linux
  GOOS_TARGET              默认: linux
  GOARCH_TARGET            默认: amd64
  CGO_ENABLED_TARGET       默认: 0
  DEPLOY_CLEAN_BUILD       默认: 0。为 1 时先执行 go clean -cache -testcache

示例：
  bash ./scripts/deploy.sh
  bash ./scripts/deploy.sh --clean-build
  bash ./scripts/deploy.sh --frontend-only
  bash ./scripts/deploy.sh --backend-only
  bash ./scripts/deploy.sh --skip-build
EOF
}

validate_file() {
  local path="$1"
  local label="$2"
  if [ ! -f "$path" ]; then
    echo "${label}不存在: $path" >&2
    exit 1
  fi
}

validate_dir() {
  local path="$1"
  local label="$2"
  if [ ! -d "$path" ]; then
    echo "${label}不存在: $path" >&2
    exit 1
  fi
}

while [ $# -gt 0 ]; do
  case "$1" in
    --clean-build)
      DEPLOY_CLEAN_BUILD=1
      ;;
    --skip-build)
      SKIP_BUILD=1
      ;;
    --frontend-only)
      INCLUDE_BACKEND=0
      INCLUDE_WEB=1
      INCLUDE_ATTENDANCE_H5=1
      INCLUDE_MIHUA_TOKEN=0
      ;;
    --backend-only)
      INCLUDE_BACKEND=1
      INCLUDE_WEB=0
      INCLUDE_ATTENDANCE_H5=0
      INCLUDE_MIHUA_TOKEN=0
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

if [ "$INCLUDE_BACKEND" != "1" ] && [ "$INCLUDE_WEB" != "1" ] && [ "$INCLUDE_ATTENDANCE_H5" != "1" ] && [ "$INCLUDE_MIHUA_TOKEN" != "1" ]; then
  echo "没有需要纳入 release 的产物" >&2
  exit 1
fi

LOCAL_RELEASE_DIR="${ROOT_DIR}/${DEPLOY_RELEASE_DIR}"
LOCAL_WEB_DIR="${LOCAL_RELEASE_DIR}/dist"
LOCAL_CHECKIN_DIR="${LOCAL_RELEASE_DIR}/check-in"
LOCAL_BIN_PATH="${LOCAL_RELEASE_DIR}/${DEPLOY_BACKEND_BIN_NAME}"
LOCAL_MIHUA_DIR="${LOCAL_RELEASE_DIR}/mihua-token-fetcher"

echo "==> release 配置"
echo "release_dir=${LOCAL_RELEASE_DIR}"
echo "backend_bin=${DEPLOY_BACKEND_BIN_NAME}"
echo "include_backend=${INCLUDE_BACKEND}"
echo "include_web=${INCLUDE_WEB}"
echo "include_attendance_h5=${INCLUDE_ATTENDANCE_H5}"
echo "include_mihua_token=${INCLUDE_MIHUA_TOKEN}"
echo "clean_build=${DEPLOY_CLEAN_BUILD}"
echo "skip_build=${SKIP_BUILD}"

if [ "$SKIP_BUILD" -eq 0 ]; then
  echo "==> 开始准备本地 release 产物"
  RELEASE_DIR="$DEPLOY_RELEASE_DIR" \
  BACKEND_BIN_NAME="$DEPLOY_BACKEND_BIN_NAME" \
  GOOS_TARGET="$GOOS_TARGET" \
  GOARCH_TARGET="$GOARCH_TARGET" \
  CGO_ENABLED_TARGET="$CGO_ENABLED_TARGET" \
  CLEAN_BUILD_TARGET="$DEPLOY_CLEAN_BUILD" \
  INCLUDE_BACKEND="$INCLUDE_BACKEND" \
  INCLUDE_WEB="$INCLUDE_WEB" \
  INCLUDE_ATTENDANCE_H5="$INCLUDE_ATTENDANCE_H5" \
  INCLUDE_MIHUA_TOKEN="$INCLUDE_MIHUA_TOKEN" \
    bash "$ROOT_DIR/scripts/package-release.sh"
fi

if [ "$INCLUDE_BACKEND" = "1" ]; then
  validate_file "$LOCAL_BIN_PATH" "后端二进制"
fi

if [ "$INCLUDE_WEB" = "1" ]; then
  validate_dir "$LOCAL_WEB_DIR" "主前端产物目录"
  validate_file "${LOCAL_WEB_DIR}/index.html" "主前端入口文件"
fi

if [ "$INCLUDE_ATTENDANCE_H5" = "1" ]; then
  validate_dir "$LOCAL_CHECKIN_DIR" "attendance-h5 产物目录"
  validate_file "${LOCAL_CHECKIN_DIR}/index.html" "attendance-h5 入口文件"
fi

if [ "$INCLUDE_MIHUA_TOKEN" = "1" ]; then
  validate_dir "$LOCAL_MIHUA_DIR" "mihua-token-fetcher 目录"
  validate_file "${LOCAL_MIHUA_DIR}/pyproject.toml" "mihua-token-fetcher 配置文件"
fi

echo "==> release 产物已准备完成"
