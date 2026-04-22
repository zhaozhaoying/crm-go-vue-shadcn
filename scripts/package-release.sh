#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${ROOT_DIR}/${RELEASE_DIR:-build/release}"
BACKEND_BIN_NAME="${BACKEND_BIN_NAME:-crm-backend}"
WEB_DIST_SRC_DIR="${ROOT_DIR}/web/dist"
WEB_DIST_OUT_DIR="${OUT_DIR}/dist"
ATTENDANCE_H5_SRC_DIR="${ROOT_DIR}/apps/attendance-h5/unpackage/dist/build/web"
ATTENDANCE_H5_OUT_DIR="${OUT_DIR}/check-in"
MIHUA_TOKEN_SRC_DIR="${ROOT_DIR}/apps/mihua-token-service"
MIHUA_TOKEN_OUT_DIR="${OUT_DIR}/mihua-token-fetcher"
INCLUDE_BACKEND="${INCLUDE_BACKEND:-1}"
INCLUDE_WEB="${INCLUDE_WEB:-1}"
INCLUDE_ATTENDANCE_H5="${INCLUDE_ATTENDANCE_H5:-1}"
INCLUDE_MIHUA_TOKEN="${INCLUDE_MIHUA_TOKEN:-1}"
GOOS_TARGET="${GOOS_TARGET:-linux}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"
CGO_ENABLED_TARGET="${CGO_ENABLED_TARGET:-1}"
CLEAN_BUILD_TARGET="${CLEAN_BUILD_TARGET:-0}"
APP_VERSION="${APP_VERSION:-$(git -C "$ROOT_DIR" describe --tags --always --dirty 2>/dev/null || echo dev)}"
GIT_COMMIT="${GIT_COMMIT:-$(git -C "$ROOT_DIR" rev-parse --short HEAD 2>/dev/null || echo unknown)}"
BUILD_TIME="${BUILD_TIME:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"
LDFLAGS="-X main.version=$APP_VERSION -X main.gitCommit=$GIT_COMMIT -X main.buildTime=$BUILD_TIME"

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "缺少必要命令: $1" >&2
    exit 1
  fi
}

sync_dir() {
  local src_dir="$1"
  local dst_dir="$2"

  mkdir -p "$dst_dir"
  if command -v rsync >/dev/null 2>&1; then
    rsync -a --delete "$src_dir/" "$dst_dir/"
  else
    rm -rf "$dst_dir"
    mkdir -p "$dst_dir"
    cp -R "$src_dir"/. "$dst_dir"/
  fi
}

cleanup_release_dir() {
  mkdir -p "$OUT_DIR"
  find "$OUT_DIR" -mindepth 1 -maxdepth 1 ! -name ".git" -exec rm -rf {} +
}

cleanup_python_caches() {
  if [ ! -d "$OUT_DIR" ]; then
    return 0
  fi

  find "$OUT_DIR" -type d \( -name "__pycache__" -o -name ".pytest_cache" -o -name ".mypy_cache" \) -prune -exec rm -rf {} +
  find "$OUT_DIR" -type f \( -name "*.pyc" -o -name ".DS_Store" \) -delete
}

build_backend() {
  echo "==> 打包后端二进制"
  if [ "$CLEAN_BUILD_TARGET" = "1" ]; then
    echo "==> 清理 Go 构建缓存"
    go clean -cache -testcache
  fi

  GOOS="$GOOS_TARGET" GOARCH="$GOARCH_TARGET" CGO_ENABLED="$CGO_ENABLED_TARGET" \
    go build -ldflags "$LDFLAGS" -o "$OUT_DIR/$BACKEND_BIN_NAME" ./
}

build_web() {
  echo "==> 构建主前端 Web"
  if command -v pnpm >/dev/null 2>&1; then
    (cd "$ROOT_DIR/web" && pnpm build)
  elif command -v npm >/dev/null 2>&1; then
    (cd "$ROOT_DIR/web" && npm run build)
  else
    echo "缺少前端构建命令: pnpm 或 npm" >&2
    exit 1
  fi

  if [ ! -d "$WEB_DIST_SRC_DIR" ]; then
    echo "主前端构建产物不存在: $WEB_DIST_SRC_DIR" >&2
    exit 1
  fi

  sync_dir "$WEB_DIST_SRC_DIR" "$WEB_DIST_OUT_DIR"
}

copy_attendance_h5() {
  echo "==> 收敛 attendance-h5 静态文件"
  if [ ! -d "$ATTENDANCE_H5_SRC_DIR" ]; then
    echo "attendance-h5 H5 静态产物不存在: $ATTENDANCE_H5_SRC_DIR" >&2
    echo "请先生成 uni-app H5 产物后再执行发布脚本。" >&2
    exit 1
  fi

  sync_dir "$ATTENDANCE_H5_SRC_DIR" "$ATTENDANCE_H5_OUT_DIR"
}

copy_mihua_token_service() {
  echo "==> 收敛 mihua-token-service"
  if [ ! -d "$MIHUA_TOKEN_SRC_DIR" ]; then
    echo "mihua-token-service 目录不存在: $MIHUA_TOKEN_SRC_DIR" >&2
    exit 1
  fi

  sync_dir "$MIHUA_TOKEN_SRC_DIR" "$MIHUA_TOKEN_OUT_DIR"
}

require_command bash

if [ "$INCLUDE_BACKEND" != "1" ] && [ "$INCLUDE_WEB" != "1" ] && [ "$INCLUDE_ATTENDANCE_H5" != "1" ] && [ "$INCLUDE_MIHUA_TOKEN" != "1" ]; then
  echo "没有需要纳入 release 的产物" >&2
  exit 1
fi

if [ "$INCLUDE_BACKEND" = "1" ]; then
  require_command go
fi

cleanup_release_dir

if [ "$INCLUDE_BACKEND" = "1" ]; then
  build_backend
fi

if [ "$INCLUDE_WEB" = "1" ]; then
  build_web
fi

if [ "$INCLUDE_ATTENDANCE_H5" = "1" ]; then
  copy_attendance_h5
fi

if [ "$INCLUDE_MIHUA_TOKEN" = "1" ]; then
  copy_mihua_token_service
fi

cleanup_python_caches

echo "version=$APP_VERSION"
echo "git_commit=$GIT_COMMIT"
echo "build_time=$BUILD_TIME"
echo "clean_build=$CLEAN_BUILD_TARGET"
echo "include_backend=$INCLUDE_BACKEND"
echo "include_web=$INCLUDE_WEB"
echo "include_attendance_h5=$INCLUDE_ATTENDANCE_H5"
echo "include_mihua_token=$INCLUDE_MIHUA_TOKEN"
echo "release_dir=$OUT_DIR"
