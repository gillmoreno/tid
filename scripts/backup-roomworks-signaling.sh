#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: backup-roomworks-signaling.sh --db ABSOLUTE_PATH --output-dir ABSOLUTE_PATH [--retain-days N]

Creates a transactionally consistent SQLite backup and verifies it with PRAGMA quick_check.
EOF
}

DB_PATH=""
OUTPUT_DIR=""
RETAIN_DAYS=14

while [[ $# -gt 0 ]]; do
  case "$1" in
    --db) DB_PATH="${2:-}"; shift 2 ;;
    --output-dir) OUTPUT_DIR="${2:-}"; shift 2 ;;
    --retain-days) RETAIN_DAYS="${2:-}"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) printf 'Unknown option: %s\n' "$1" >&2; usage >&2; exit 1 ;;
  esac
done

[[ "$DB_PATH" == /* && -f "$DB_PATH" ]] || { printf 'A readable absolute --db path is required\n' >&2; exit 1; }
[[ "$OUTPUT_DIR" == /* ]] || { printf 'An absolute --output-dir path is required\n' >&2; exit 1; }
[[ "$RETAIN_DAYS" =~ ^[1-9][0-9]*$ ]] || { printf '--retain-days must be a positive integer\n' >&2; exit 1; }
command -v sqlite3 >/dev/null 2>&1 || { printf 'sqlite3 is required\n' >&2; exit 1; }

umask 077
mkdir -p "$OUTPUT_DIR"
chmod 700 "$OUTPUT_DIR"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
target="${OUTPUT_DIR}/roomworks-signaling-${timestamp}.db"
escaped_target="${target//\'/\'\'}"
sqlite3 "$DB_PATH" ".backup '${escaped_target}'"
[[ "$(sqlite3 "$target" 'PRAGMA quick_check;')" == "ok" ]] || {
  printf 'Backup verification failed: %s\n' "$target" >&2
  exit 1
}
chmod 600 "$target"
find "$OUTPUT_DIR" -type f -name 'roomworks-signaling-*.db' -mtime "+${RETAIN_DAYS}" -delete
printf '%s\n' "$target"
