#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

URL=""
START=""
END=""
OUT_DIR="drafts"
DRAFT_ID=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --url) URL="$2"; shift 2 ;;
    --start) START="$2"; shift 2 ;;
    --end) END="$2"; shift 2 ;;
    --out) OUT_DIR="$2"; shift 2 ;;
    --id) DRAFT_ID="$2"; shift 2 ;;
    --slug) SLUG="$2"; shift 2 ;;
    *) echo "Unknown option $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$URL" || -z "$START" || -z "$END" ]]; then
  echo "Usage: $0 --url URL --start START --end END [--id ID | --slug slug] [--out DIR]" >&2
  exit 1
fi

ID="${DRAFT_ID:-$(default_draft_id "$URL" "${SLUG:-}")}"
mkdir -p "$OUT_DIR/$ID"

echo "Getting stream URL..."
STREAM="$(yt-dlp -g -f "bv*+ba/b" "$URL" | head -1)"

echo "Extracting clip $START → $END..."
ffmpeg -ss "$START" -to "$END" -i "$STREAM" \
  -c:v libx264 -c:a aac -movflags +faststart \
  "$OUT_DIR/$ID/clip.mp4" -y

echo "Clip saved: $OUT_DIR/$ID/clip.mp4"
echo "$ID"