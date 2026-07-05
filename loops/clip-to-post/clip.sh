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

echo "Getting stream URLs..."
VIDEO_URL="$(yt-dlp -g -f "bv" "$URL")"
AUDIO_URL="$(yt-dlp -g -f "ba" "$URL")"

CLIP_OUT="$OUT_DIR/$ID/clip.mp4"
echo "Extracting clip $START → $END..."

if [[ -n "$AUDIO_URL" ]]; then
  # YouTube serves separate video + audio streams; mux both or clip is silent.
  ffmpeg -ss "$START" -to "$END" -i "$VIDEO_URL" \
    -ss "$START" -to "$END" -i "$AUDIO_URL" \
    -map 0:v:0 -map 1:a:0 -shortest \
    -c:v libx264 -c:a aac -b:a 192k -movflags +faststart \
    "$CLIP_OUT" -y
else
  # Fallback for muxed sources (single stream with audio baked in).
  ffmpeg -ss "$START" -to "$END" -i "$VIDEO_URL" \
    -c:v libx264 -c:a aac -movflags +faststart \
    "$CLIP_OUT" -y
fi

echo "Clip saved: $CLIP_OUT"
echo "$ID"