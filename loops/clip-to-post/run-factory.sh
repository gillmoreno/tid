#!/bin/bash
# run-factory.sh — agent entry point: YouTube URL → SQLite candidates → /factory UI
#
# Usage:
#   ./run-factory.sh --url "https://youtube.com/watch?v=..." [--title "Guest"] [--podcast "Show"]
#
# Requires: just dev (or Go API on :8080), jq, curl

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
API="${FACTORY_API_URL:-http://localhost:8080/api/factory}"
FRONTEND="${FACTORY_UI_URL:-http://localhost:5180/factory}"

URL=""
TITLE=""
PODCAST=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --url) URL="$2"; shift 2 ;;
    --title) TITLE="$2"; shift 2 ;;
    --podcast) PODCAST="$2"; shift 2 ;;
    -h|--help)
      sed -n '2,6p' "$0"
      exit 0
      ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

[[ -n "$URL" ]] || { echo "Missing --url" >&2; exit 1; }

require_api() {
  if ! curl -sf "$API/biases" >/dev/null 2>&1; then
    echo "✘ Post Factory API not reachable at $API" >&2
    echo "  Start dev from repo root: just dev" >&2
    exit 1
  fi
}

require_api

echo "→ Ingesting source"
INGEST_BODY="$(jq -n --arg u "$URL" --arg t "$TITLE" --arg p "$PODCAST" \
  '{youtube_url: $u, title: $t, podcast: $p}')"
SOURCE_JSON="$(curl -sf -X POST "$API/sources" -H "Content-Type: application/json" -d "$INGEST_BODY")"
SOURCE_ID="$(echo "$SOURCE_JSON" | jq -r '.id')"
echo "  source: $SOURCE_ID"

echo "→ Analyzing (transcript + biases + prompt) — may take 1–2 min"
ANALYZE_JSON="$(curl -sf -X POST "$API/sources/$SOURCE_ID/analyze")"
COUNT="$(echo "$ANALYZE_JSON" | jq '.candidates | length')"
echo "  candidates: $COUNT"

echo ""
echo "✓ Post Factory run complete"
echo "  UI:      $FRONTEND"
echo "  source:  $SOURCE_ID"
echo "  db:      ${DATABASE_PATH:-$ROOT/data/factory/tid.db}"
echo ""
echo "Gil: open the UI, edit takes, clip, schedule. At post time: Tick now or just factory-tick."

if [[ "$COUNT" -gt 0 ]]; then
  echo ""
  echo "Candidates:"
  echo "$ANALYZE_JSON" | jq -r '.candidates[] | "  [\(.rank)] \(.hook) (\(.start_time)–\(.end_time))"'
fi