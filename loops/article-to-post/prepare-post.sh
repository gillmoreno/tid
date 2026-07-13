#!/bin/bash
set -euo pipefail

# prepare-post.sh — semi-automated posting prep for an article post (text only)
#
# 1. Copy post_text to clipboard
# 2. Open Chrome Default profile on X compose
#
# Usage:
#   ./prepare-post.sh --draft <candidate-id>
#   ./prepare-post.sh <candidate-id>

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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

You finish manually: Cmd+V paste text, review, click Post.
EOF
      exit 0
      ;;
    -*) echo "Unknown option: $1" >&2; exit 1 ;;
    *)
      [[ -z "$DRAFT_ID" ]] || { echo "Unexpected argument: $1" >&2; exit 1; }
      DRAFT_ID="$1"; shift ;;
  esac
done

[[ -n "$DRAFT_ID" ]] || { echo "Usage: ./prepare-post.sh --draft <id>" >&2; exit 1; }

DRAFT_DIR="$SCRIPT_DIR/drafts/$DRAFT_ID"
META="$DRAFT_DIR/meta.json"
[[ -f "$META" ]] || { echo "ERROR: no draft at $DRAFT_DIR" >&2; exit 1; }

POST_TEXT="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("post_text",""))' "$META")"
PUBLICATION="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("publication",""))' "$META")"
TITLE="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("title",""))' "$META")"

[[ -n "$POST_TEXT" ]] || { echo "ERROR: post_text missing in $META" >&2; exit 1; }

printf '%s' "$POST_TEXT" | pbcopy

if [[ -d "/Applications/Google Chrome.app" ]]; then
  open -na "Google Chrome" --args --profile-directory="Default" "https://x.com/compose/post"
else
  open "https://x.com/compose/post"
fi

cat <<EOF
✓ Draft: $DRAFT_ID
✓ Publication: $PUBLICATION
✓ Article: $TITLE
✓ Copied post text to clipboard
✓ Opened X compose in Chrome (Default profile)

Next (manual):
  1. Cmd+V to paste text in compose
  2. Review and click Post
EOF
