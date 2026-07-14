#!/bin/bash
set -euo pipefail

# rewrite.sh — apply Gil's lens + instruction to an article post_text
#
# Usage: ./rewrite.sh --input rewrite-input.json --out rewrite-output.json

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../llm.sh"
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

OUT_DIR="$(dirname "$OUT")"
mkdir -p "$OUT_DIR"

PROMPT_FILE="$OUT_DIR/rewrite-prompt.txt"
python3 - "$INPUT" "$PROMPT_FILE" <<'PY'
import json, sys

inp = json.load(open(sys.argv[1]))
handle = (inp.get("publication_handle") or "").strip()
attribution = f"@{handle}" if handle else (inp.get("publication") or "the source publication")
prompt = f"""You are the Article Factory copy editor for The Idea Guy (Gil).

BIASES (Gil's lens — curiosity mixed with skepticism):
{inp.get('biases', '')}

MENTIONS DICTIONARY (tag @ handles when names appear):
{inp.get('mentions', '')}

INSTRUCTION:
{inp.get('instruction', '')}

CURRENT POST TEXT:
{inp.get('post_text', '')}

TASK:
Rewrite post_text to follow the instruction while keeping Gil's voice.
- Stay faithful to the underlying claim — do not invent facts
- Direct tone. No emojis. No hashtags. No engagement bait
- Tag people/companies from MENTIONS dictionary with @ handles when referenced
- End post_text with the publication attribution: {attribution}
- Elicit curiosity; skeptical-curious, not cynical

Return ONLY valid JSON (no markdown fences):
{{"post_text": "..."}}"""
open(sys.argv[2], "w", encoding="utf-8").write(prompt)
PY

if factory_generate "$PROMPT_FILE" "$OUT_DIR/rewrite-raw.txt"; then
  if [[ -s "$OUT_DIR/rewrite-raw.txt" ]]; then
    : > "$OUT"
    python3 - "$OUT_DIR/rewrite-raw.txt" "$OUT" <<'PY'
import json, re, sys
raw = open(sys.argv[1]).read()
m = re.search(r'\{.*\}', raw, re.S)
if m:
    data = json.loads(m.group(0))
    if "post_text" not in data or not str(data["post_text"]).strip():
        sys.exit(1)
    json.dump(data, open(sys.argv[2], "w"), indent=2)
    sys.exit(0)
sys.exit(1)
PY
    if [[ -s "$OUT" ]]; then
      echo "Rewrite via $FACTORY_GENERATION_PROVIDER → $OUT"
      cat "$OUT"
      exit 0
    fi
  fi
fi

echo "ERROR: rewrite failed" >&2
exit 1
