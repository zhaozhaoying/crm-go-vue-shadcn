#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"

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
  go build -o crm-backend .
)

echo "==> Building frontend"
(
  cd "$FRONTEND_DIR"
  npm ci
  npm run build
)

echo "==> Build completed"
