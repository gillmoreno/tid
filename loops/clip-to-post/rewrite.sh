#!/bin/bash
set -euo pipefail

# rewrite.sh — apply Gil's lens + instruction to hook/take/post_text
#
# Usage: ./rewrite.sh --input rewrite-input.json --out rewrite-output.json

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INPUT=""
OUT=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --input) INPUT="$2"; shift 2 ;;
    --out) OUT="$2"; shift 2 ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

[[ -n "$INPUT" && -f "$INPUT" ]] || { echo "Missing --input file" >&2; exit 1; }
[[ -n "$OUT" ]] || { echo "Missing --out path" >&2; exit 1; }

BIASES="$(jq -r '.biases' "$INPUT")"
MENTIONS="$(jq -r '.mentions // ""' "$INPUT")"
INSTRUCTION="$(jq -r '.instruction' "$INPUT")"
HOOK="$(jq -r '.hook' "$INPUT")"
TAKE="$(jq -r '.take' "$INPUT")"
POST_TEXT="$(jq -r '.post_text' "$INPUT")"
PODCAST="$(jq -r '.podcast' "$INPUT")"
PODCAST_HANDLE="$(jq -r '.podcast_handle // ""' "$INPUT")"

OUT_DIR="$(dirname "$OUT")"
mkdir -p "$OUT_DIR"

cat >"$OUT_DIR/prompt.txt" <<EOF
You are the Post Factory copy editor for The Idea Guy (Gil).

BIASES (Gil's lens — curiosity mixed with skepticism):
${BIASES}

MENTIONS DICTIONARY (tag @ handles when names appear):
${MENTIONS}

INSTRUCTION:
${INSTRUCTION}

CURRENT COPY:
hook: ${HOOK}
take: ${TAKE}
post_text:
${POST_TEXT}

TASK:
Rewrite hook, take, and post_text to follow the instruction while keeping Gil's voice.
- Stay faithful to the underlying claim — do not invent facts
- Direct tone. No emojis. No hashtags. No engagement bait
- Tag people/companies from MENTIONS dictionary with @ handles when referenced
- post_text ends with podcast @ only: @${PODCAST_HANDLE:-theallinpod} — never a YouTube URL
- Format A (essay beats) or Format B (tight quote) per instruction
- Elicit curiosity; skeptical-curious, not cynical

Return ONLY valid JSON (no markdown fences):
{"hook": "...", "take": "...", "post_text": "..."}
EOF

PROMPT_TEXT="$(cat "$OUT_DIR/prompt.txt")"
rm -f "$OUT_DIR/prompt.txt" 2>/dev/null || true

if command -v grok >/dev/null 2>&1 && [[ -f "$HOME/.grok/auth.json" || -n "${XAI_API_KEY:-}" ]]; then
  grok --no-auto-update -p "$PROMPT_TEXT" --output-format plain > "$OUT_DIR/rewrite-raw.txt" 2>/dev/null || true
  if [[ -s "$OUT_DIR/rewrite-raw.txt" ]]; then
    python3 - "$OUT_DIR/rewrite-raw.txt" "$OUT" <<'PY'
import json, re, sys
raw = open(sys.argv[1]).read()
m = re.search(r'\{.*\}', raw, re.S)
if m:
    data = json.loads(m.group(0))
    for key in ("hook", "take", "post_text"):
        if key not in data or not str(data[key]).strip():
            sys.exit(1)
    json.dump(data, open(sys.argv[2], "w"), indent=2)
    sys.exit(0)
sys.exit(1)
PY
    if [[ -f "$OUT" ]]; then
      echo "Rewrite via grok → $OUT"
      cat "$OUT"
      exit 0
    fi
  fi
fi

echo "ERROR: grok unavailable or rewrite failed" >&2
exit 1