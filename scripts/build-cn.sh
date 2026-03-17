#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"
APP_VERSION="${APP_VERSION:-$(git -C "$ROOT_DIR" describe --tags --always --dirty 2>/dev/null || echo dev)}"
GIT_COMMIT="${GIT_COMMIT:-$(git -C "$ROOT_DIR" rev-parse --short HEAD 2>/dev/null || echo unknown)}"
BUILD_TIME="${BUILD_TIME:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"
LDFLAGS="-X main.version=$APP_VERSION -X main.gitCommit=$GIT_COMMIT -X main.buildTime=$BUILD_TIME"

export GOPROXY="${GOPROXY:-https://goproxy.cn,https://goproxy.io,direct}"
export GOSUMDB="${GOSUMDB:-sum.golang.google.cn}"
export GOTOOLCHAIN="${GOTOOLCHAIN:-local}"
export GOCACHE="${GOCACHE:-$BACKEND_DIR/.cache/go-build}"
export GOMODCACHE="${GOMODCACHE:-$BACKEND_DIR/.cache/go-mod}"

mkdir -p "$GOCACHE" "$GOMODCACHE"

echo "==> Building backend"
(
  cd "$BACKEND_DIR"
  go mod download
  go build -ldflags "$LDFLAGS" -o crm-backend .
)

echo "==> Building frontend"
(
  cd "$FRONTEND_DIR"
  npm ci
  npm run build
)

echo "==> Build metadata"
echo "version=$APP_VERSION"
echo "git_commit=$GIT_COMMIT"
echo "build_time=$BUILD_TIME"

echo "==> Build completed"
