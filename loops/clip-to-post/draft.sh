#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

URL=""
START=""
END=""
SPEAKER=""
PODCAST=""
OUT_DIR="drafts"
DRAFT_ID=""
SLUG=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --url) URL="$2"; shift 2 ;;
    --start) START="$2"; shift 2 ;;
    --end) END="$2"; shift 2 ;;
    --speaker) SPEAKER="$2"; shift 2 ;;
    --podcast) PODCAST="$2"; shift 2 ;;
    --out) OUT_DIR="$2"; shift 2 ;;
    --id) DRAFT_ID="$2"; shift 2 ;;
    --slug) SLUG="$2"; shift 2 ;;
    *) echo "Unknown option $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$URL" || -z "$START" || -z "$END" || -z "$SPEAKER" || -z "$PODCAST" ]]; then
  echo "Usage: $0 --url URL --start START --end END --speaker NAME --podcast NAME [--slug slug] [--id ID]" >&2
  exit 1
fi

ID="${DRAFT_ID:-$(default_draft_id "$URL" "$SLUG")}"
mkdir -p "$OUT_DIR/$ID"

echo "=== Draft ID: $ID ==="

echo "=== clip.sh ==="
"$SCRIPT_DIR/clip.sh" --url "$URL" --start "$START" --end "$END" --out "$OUT_DIR" --id "$ID"

echo "=== transcript.sh ==="
"$SCRIPT_DIR/transcript.sh" --url "$URL" --out "$OUT_DIR" --id "$ID"

echo "=== draft-copy.sh ==="
"$SCRIPT_DIR/draft-copy.sh" \
  --transcript "$OUT_DIR/$ID/transcript.txt" \
  --speaker "$SPEAKER" \
  --podcast "$PODCAST" \
  --out "$OUT_DIR" \
  --id "$ID"

jq -n \
  --arg id "$ID" \
  --arg source_url "$URL" \
  --arg speaker "$SPEAKER" \
  --arg podcast "$PODCAST" \
  --arg start "$START" \
  --arg end "$END" \
  --arg clip_path "$OUT_DIR/$ID/clip.mp4" \
  --arg created_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --rawfile post_text "$OUT_DIR/$ID/post.txt" \
  '{
    id: $id,
    source_url: $source_url,
    speaker: $speaker,
    podcast: $podcast,
    start: $start,
    end: $end,
    status: "draft",
    post_text: $post_text,
    clip_path: $clip_path,
    created_at: $created_at
  }' > "$OUT_DIR/$ID/meta.json"

echo "=== Draft complete ==="
echo "Folder: $OUT_DIR/$ID"
echo "Review post.txt, edit meta.json, set status to approved, then:"
echo "  ./prepare-post.sh $ID"
ls -lh "$OUT_DIR/$ID/"