#!/usr/bin/env bash
# Local dev: Go API + Vite (Post Factory needs host scripts — not Docker)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_PORT="${GO_API_PORT:-8080}"
FRONTEND_PORT="${VITE_PORT:-5180}"
export TID_REPO_ROOT="$ROOT"
export DATABASE_PATH="${DATABASE_PATH:-$ROOT/data/factory/tid.db}"

mkdir -p "$(dirname "$DATABASE_PATH")" "$ROOT/data/dev"

kill_port() {
  local port="$1"
  lsof -ti ":$port" | xargs kill -9 2>/dev/null || true
}

kill_port "$BACKEND_PORT"
kill_port "$FRONTEND_PORT"

echo "→ Starting Go API on :$BACKEND_PORT (db: $DATABASE_PATH)"
cd "$ROOT/go-backend"
go run ./cmd/server >"$ROOT/data/dev/go-api.log" 2>&1 &
GO_PID=$!
echo "$GO_PID" >"$ROOT/data/dev/go-api.pid"

for _ in $(seq 1 30); do
  if curl -sf "http://localhost:$BACKEND_PORT/health" >/dev/null 2>&1; then
    break
  fi
  sleep 0.2
done

if ! curl -sf "http://localhost:$BACKEND_PORT/api/factory/biases" >/dev/null 2>&1; then
  echo "✘ Go API started but factory routes unavailable. Check data/dev/go-api.log"
  exit 1
fi

echo "→ Starting Vite on :$FRONTEND_PORT (factory enabled)"
cd "$ROOT/frontend"
export VITE_API_URL=/api
export VITE_API_PROXY_TARGET="http://localhost:$BACKEND_PORT"
export VITE_FACTORY_ENABLED=true
exec npm run dev -- --port "$FRONTEND_PORT"