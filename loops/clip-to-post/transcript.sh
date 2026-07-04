#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

URL=""
OUT_DIR="drafts"
DRAFT_ID=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --url) URL="$2"; shift 2 ;;
    --out) OUT_DIR="$2"; shift 2 ;;
    --id) DRAFT_ID="$2"; shift 2 ;;
    --slug) SLUG="$2"; shift 2 ;;
    *) echo "Unknown option $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$URL" ]]; then
  echo "Usage: $0 --url URL [--id ID | --slug slug] [--out DIR]" >&2
  exit 1
fi

ID="${DRAFT_ID:-$(default_draft_id "$URL" "${SLUG:-}")}"
mkdir -p "$OUT_DIR/$ID"

VIDEO_ID="$(echo "$URL" | sed -n 's/.*[?&]v=\([^&]*\).*/\1/p')"
[[ -n "$VIDEO_ID" ]] || VIDEO_ID="$(echo "$URL" | sed 's|.*youtu.be/||' | cut -c1-11)"

echo "Downloading auto-captions..."
yt-dlp --write-auto-sub --sub-lang en --skip-download \
  -o "$OUT_DIR/$ID/%(id)s" "$URL" 2>/dev/null || true

VTT="$OUT_DIR/$ID/${VIDEO_ID}.en.vtt"
if [[ ! -f "$VTT" ]]; then
  VTT="$(find "$OUT_DIR/$ID" -name '*.en.vtt' | head -1)"
fi

if [[ -n "$VTT" && -f "$VTT" ]]; then
  awk '
    /-->/ { next }
    { gsub(/<[^>]*>/, "") }
    length($0) > 3 && $0 !~ /^[0-9]/ && $0 !~ /^WEBVTT/ && $0 !~ /^Kind:/ && $0 !~ /^Language:/ { print }
  ' "$VTT" | awk '!seen[$0]++' > "$OUT_DIR/$ID/transcript.txt"
  echo "Transcript saved: $OUT_DIR/$ID/transcript.txt"
else
  echo "No captions found. Add transcript manually or use whisper."
  touch "$OUT_DIR/$ID/transcript.txt"
fi

echo "$ID"