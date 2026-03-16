#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${ROOT_DIR}/${RELEASE_DIR:-release}"
BACKEND_BIN_NAME="${BACKEND_BIN_NAME:-crm-backend}"
GOOS_TARGET="${GOOS_TARGET:-linux}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"
CGO_ENABLED_TARGET="${CGO_ENABLED_TARGET:-1}"

# Keep existing .git if this is a standalone release repo
mkdir -p "$OUT_DIR"
find "$OUT_DIR" -mindepth 1 -maxdepth 1 ! -name ".git" -exec rm -rf {} +

# Build backend
(
  cd "$ROOT_DIR/backend"
  GOOS="$GOOS_TARGET" GOARCH="$GOARCH_TARGET" CGO_ENABLED="$CGO_ENABLED_TARGET" \
    go build -o "$OUT_DIR/$BACKEND_BIN_NAME" ./
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

# Copy backend .env (if exists)
if [ -f "$ROOT_DIR/backend/.env" ]; then
  cp "$ROOT_DIR/backend/.env" "$OUT_DIR/.env"
else
  echo "backend/.env not found, skipped" >&2
fi

echo "Release prepared at: $OUT_DIR"
