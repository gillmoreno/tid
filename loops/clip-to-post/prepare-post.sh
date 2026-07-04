#!/bin/bash
set -euo pipefail

# prepare-post.sh — semi-automated posting prep (reads meta.json only)
#
# 1. Copy post_text to clipboard
# 2. Open Chrome Default profile on X compose
# 3. Open Finder on the draft folder (drag clip.mp4 into compose)
#
# Usage:
#   ./prepare-post.sh --draft 20260704-naval-taiwan-competition
#   ./prepare-post.sh 20260704-naval-taiwan-competition

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

DRAFT_ID=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --draft) DRAFT_ID="$2"; shift 2 ;;
    -h|--help)
      cat <<'EOF'
Usage: ./prepare-post.sh --draft <id>
       ./prepare-post.sh <id>

Reads drafts/<id>/meta.json and:
  - copies post_text to clipboard
  - opens Chrome (Default profile) on https://x.com/compose/post
  - opens Finder on the draft folder containing clip.mp4

You finish manually: Cmd+V paste text, drag clip.mp4, click Post.
EOF
      exit 0
      ;;
    -*)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
    *)
      [[ -z "$DRAFT_ID" ]] || { echo "Unexpected argument: $1" >&2; exit 1; }
      DRAFT_ID="$1"
      shift
      ;;
  esac
done

[[ -n "$DRAFT_ID" ]] || {
  echo "Usage: ./prepare-post.sh --draft <id>" >&2
  exit 1
}

DRAFT_DIR="$(require_draft "$DRAFT_ID")"
META="$DRAFT_DIR/meta.json"
POST_TEXT="$(meta_field "$META" "post_text")"
CLIP_PATH="$(resolve_clip_path "$DRAFT_DIR")"
SPEAKER="$(meta_field "$META" "speaker")"
PODCAST="$(meta_field "$META" "podcast")"
STATUS="$(meta_field "$META" "status")"

[[ -n "$POST_TEXT" ]] || {
  echo "ERROR: post_text missing in $META" >&2
  exit 1
}

printf '%s' "$POST_TEXT" | pbcopy

if [[ -d "/Applications/Google Chrome.app" ]]; then
  open -na "Google Chrome" --args --profile-directory="Default" "https://x.com/compose/post"
else
  open "https://x.com/compose/post"
fi

open "$DRAFT_DIR"

cat <<EOF
✓ Draft: $DRAFT_ID ($STATUS)
✓ Speaker: $SPEAKER · $PODCAST
✓ Copied post text to clipboard
✓ Opened X compose in Chrome (Default profile)
✓ Opened Finder: $DRAFT_DIR

Next (manual):
  1. Cmd+V to paste text in compose
  2. Drag $(basename "$CLIP_PATH") into the compose window
  3. Review and click Post
EOF