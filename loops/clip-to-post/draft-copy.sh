#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TRANSCRIPT=""
SPEAKER=""
PODCAST=""
OUT_DIR="drafts"
DRAFT_ID=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --transcript) TRANSCRIPT="$2"; shift 2 ;;
    --speaker) SPEAKER="$2"; shift 2 ;;
    --podcast) PODCAST="$2"; shift 2 ;;
    --out) OUT_DIR="$2"; shift 2 ;;
    --id) DRAFT_ID="$2"; shift 2 ;;
    *) echo "Unknown option $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$TRANSCRIPT" || -z "$SPEAKER" || -z "$PODCAST" ]]; then
  echo "Usage: $0 --transcript FILE --speaker NAME --podcast NAME [--id ID] [--out DIR]" >&2
  exit 1
fi

ID="${DRAFT_ID:-$(basename "$(dirname "$TRANSCRIPT")")}"
mkdir -p "$OUT_DIR/$ID"

# Phase 1: simple excerpt. Phase 2: LLM + taste.md prompt.
HOOK="$(tr '\n' ' ' < "$TRANSCRIPT" | sed 's/  */ /g' | cut -c1-120 | sed 's/ *$//')"

{
  printf '%s: %s\n\n' "$SPEAKER" "$HOOK"
  tr '\n' ' ' < "$TRANSCRIPT" | sed 's/  */ /g' | fold -s -w 110 | head -3
  printf '\n\nSource: %s\n' "$PODCAST"
} > "$OUT_DIR/$ID/post.txt"

echo "Draft post saved: $OUT_DIR/$ID/post.txt"
cat "$OUT_DIR/$ID/post.txt"