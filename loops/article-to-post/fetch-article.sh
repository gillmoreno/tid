#!/bin/bash
set -euo pipefail

# fetch-article.sh — extract clean text from an article URL into drafts/{id}/
#
# Usage: ./fetch-article.sh --url URL --id ID [--out drafts]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

URL=""
OUT_DIR="drafts"
ID=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --url) URL="$2"; shift 2 ;;
    --out) OUT_DIR="$2"; shift 2 ;;
    --id) ID="$2"; shift 2 ;;
    *) echo "Unknown option $1" >&2; exit 1 ;;
  esac
done

[[ -n "$URL" ]] || { echo "Usage: $0 --url URL --id ID [--out DIR]" >&2; exit 1; }
[[ -n "$ID" ]] || { echo "Missing --id" >&2; exit 1; }

WORK_DIR="$OUT_DIR/$ID"
mkdir -p "$WORK_DIR"

python3 "$SCRIPT_DIR/extract_article.py" --url "$URL" --out "$WORK_DIR"
