#!/bin/bash
set -euo pipefail

# rewrite.sh — apply Gil's lens + instruction to post_text
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

BIASES="$(jq -r '.biases' "$INPUT")"
MENTIONS="$(jq -r '.mentions // ""' "$INPUT")"
INSTRUCTION="$(jq -r '.instruction' "$INPUT")"
POST_TEXT="$(jq -r '.post_text' "$INPUT")"
PODCAST="$(jq -r '.podcast' "$INPUT")"
PODCAST_HANDLE="$(jq -r '.podcast_handle // ""' "$INPUT")"

OUT_DIR="$(dirname "$OUT")"
mkdir -p "$OUT_DIR"

# Resolve humanize instructions relative to this script for reliable inclusion
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HUMANIZE_FILE="$SCRIPT_DIR/humanize-instructions.md"
HUMANIZE_CONTENT=""
if [[ -f "$HUMANIZE_FILE" ]]; then
  HUMANIZE_CONTENT="$(<"$HUMANIZE_FILE")"
fi


PROMPT_FILE="$OUT_DIR/rewrite-prompt.txt"

# Pass humanize content as a 3rd argument to the python builder (avoids __file__ issues)
python3 - "$INPUT" "$PROMPT_FILE" "$HUMANIZE_CONTENT" <<'PY'
import json, sys

inp = json.load(open(sys.argv[1]))
handle = inp.get("podcast_handle") or "theallinpod"
instruction = inp.get('instruction', '') or ''
post_text = inp.get('post_text', '') or ''
humanize_raw = sys.argv[3] if len(sys.argv) > 3 else ""

humanize_section = ""
lower_inst = (instruction or "").lower()
needs_humanize = any(k in lower_inst for k in ["humanize", "de-ai", "remove ai", "ai smell", "ai fingerprint"])

if needs_humanize and humanize_raw.strip():
    humanize_section = "\n\nHUMANIZE RULES (apply these as the final pass — highest priority for this edit):\n" + humanize_raw

prompt = f"""You are the Post Factory copy editor for The Idea Guy (Gil).

BIASES (Gil's lens — curiosity mixed with skepticism):
{inp.get('biases', '')}

MENTIONS DICTIONARY (tag @ handles when names appear):
{inp.get('mentions', '')}

INSTRUCTION:
{instruction}
{humanize_section}

CURRENT POST TEXT:
{post_text}

TASK:
Rewrite post_text to follow the instruction while keeping Gil's voice.
- Stay faithful to the underlying claim — do not invent facts not in the clip
- When the instruction asks for Gil's take, commentary, or opinion: use Format C — Gil's skeptical-curious opinion first, anchored in what the speaker said (not a full quote dump)
- Direct tone. No emojis. No hashtags. No engagement bait
- Tag people/companies from MENTIONS dictionary with @ handles when referenced
- post_text ends with podcast @ only: @{handle} — never a YouTube URL
- Format A (essay beats) or Format B (tight quote) per instruction
- Elicit curiosity; skeptical-curious, not cynical

When humanizing is requested (via instruction or HUMANIZE RULES present):
- Follow the HUMANIZE RULES above rigorously as the final editing pass.
- Remove AI tells: boilerplate openers/scaffolding, inflated diction, mechanical parallel structures and tricolons at machine frequency, uniform sentence rhythm, em-dash overload, stacked hedging, and conclusions that just restate.
- Vary sentence length aggressively. Follow long clauses with short ones or fragments. Allow the prose to have texture and a slightly uneven point of view.
- Keep every real claim, concrete number, name, and specific detail exactly as-is. Do not fabricate facts, anecdotes, or "more human" color.
- Result must read like the same content written by a careful, opinionated person — not a language model performing confidence.

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
