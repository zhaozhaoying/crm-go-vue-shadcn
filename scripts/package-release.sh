#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${ROOT_DIR}/${RELEASE_DIR:-release}"
BACKEND_BIN_NAME="${BACKEND_BIN_NAME:-crm-backend}"
GOOS_TARGET="${GOOS_TARGET:-linux}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"
CGO_ENABLED_TARGET="${CGO_ENABLED_TARGET:-1}"
CLEAN_BUILD_TARGET="${CLEAN_BUILD_TARGET:-0}"
APP_VERSION="${APP_VERSION:-$(git -C "$ROOT_DIR" describe --tags --always --dirty 2>/dev/null || echo dev)}"
GIT_COMMIT="${GIT_COMMIT:-$(git -C "$ROOT_DIR" rev-parse --short HEAD 2>/dev/null || echo unknown)}"
BUILD_TIME="${BUILD_TIME:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"
LDFLAGS="-X main.version=$APP_VERSION -X main.gitCommit=$GIT_COMMIT -X main.buildTime=$BUILD_TIME"

cleanup_release_artifacts() {
  local target_dir="$1"
  if [ ! -d "$target_dir" ]; then
    return 0
  fi

  find "$target_dir" \
    -path "$target_dir/.git" -prune -o \
    -type f \( \
      -name "*.sh" -o \
      -name "*_test.go" -o \
      -name "*.test" -o \
      -name "go-test.out" -o \
      -name "coverage*.out" \
    \) -print -delete

  find "$target_dir" \
    -path "$target_dir/.git" -prune -o \
    -type d \( \
      -name "__tests__" -o \
      -name "tests" \
    \) -prune -print | while IFS= read -r path; do
      rm -rf "$path"
    done
}

# Keep existing .git if this is a standalone release repo
mkdir -p "$OUT_DIR"
find "$OUT_DIR" -mindepth 1 -maxdepth 1 ! -name ".git" -exec rm -rf {} +

# Build backend
(
  cd "$ROOT_DIR/backend"
  # clean-build 模式会先清理本机 Go 构建缓存，确保后端重新编译。
  # 这里不清 modcache，避免每次都重新下载依赖。
  if [ "$CLEAN_BUILD_TARGET" = "1" ]; then
    echo "==> Cleaning Go build cache"
    go clean -cache -testcache
  fi
  GOOS="$GOOS_TARGET" GOARCH="$GOARCH_TARGET" CGO_ENABLED="$CGO_ENABLED_TARGET" \
    go build -ldflags "$LDFLAGS" -o "$OUT_DIR/$BACKEND_BIN_NAME" ./
)

# Build frontend
if command -v pnpm >/dev/null 2>&1; then
  (cd "$ROOT_DIR/frontend" && pnpm build)
elif command -v npm >/dev/null 2>&1; then
  (cd "$ROOT_DIR/frontend" && npm run build)
else
  echo "pnpm or npm is required to build frontend" >&2
  exit 1
fi

# Copy frontend dist
if command -v rsync >/dev/null 2>&1; then
  rsync -a --delete "$ROOT_DIR/frontend/dist/" "$OUT_DIR/dist/"
else
  rm -rf "$OUT_DIR/dist"
  cp -R "$ROOT_DIR/frontend/dist" "$OUT_DIR/dist"
fi

# 清理发布目录里不需要上线的测试文件和脚本。
cleanup_release_artifacts "$OUT_DIR"

echo "version=$APP_VERSION"
echo "git_commit=$GIT_COMMIT"
echo "build_time=$BUILD_TIME"
echo "clean_build=$CLEAN_BUILD_TARGET"
echo "Release prepared at: $OUT_DIR"
