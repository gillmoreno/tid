#!/bin/bash
set -euo pipefail

# analyze_moment.sh — spot moment → 3 post takes (short / medium / long)
# Usage: ./analyze_moment.sh --input drafts/{source}/moment-input.json --out drafts/{source}/moment-analysis.json

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
TRANSCRIPT="$(jq -r '.transcript' "$INPUT")"
FOCUS_NOTE="$(jq -r '.focus_note // ""' "$INPUT")"
WINDOW_START="$(jq -r '.window_start' "$INPUT")"
WINDOW_END="$(jq -r '.window_end' "$INPUT")"

OUT_DIR="$(dirname "$OUT")"
mkdir -p "$OUT_DIR"

export FACTORY_BIASES="$BIASES"
export FACTORY_MENTIONS="$MENTIONS"
export FACTORY_FOCUS_NOTE="$FOCUS_NOTE"
export FACTORY_WINDOW_START="$WINDOW_START"
export FACTORY_WINDOW_END="$WINDOW_END"
SYSTEM_PROMPT="$(python3 "$SCRIPT_DIR/build_moment_prompt.py")"

if command -v grok >/dev/null 2>&1 && [[ -f "$HOME/.grok/auth.json" || -n "${XAI_API_KEY:-}" ]]; then
  PROMPT_TEXT="${SYSTEM_PROMPT}

TRANSCRIPT (timestamped, Gil's selected range ${WINDOW_START} → ${WINDOW_END}):
${TRANSCRIPT}"

  grok --no-auto-update -p "$PROMPT_TEXT" --output-format plain > "$OUT_DIR/moment-analyze-raw.txt" 2>/dev/null || true
  if [[ -s "$OUT_DIR/moment-analyze-raw.txt" ]]; then
    python3 - "$OUT_DIR/moment-analyze-raw.txt" "$OUT" <<'PY'
import json, re, sys
raw = open(sys.argv[1]).read()
m = re.search(r'\{.*\}', raw, re.S)
if m:
    data = json.loads(m.group(0))
    json.dump(data, open(sys.argv[2], "w"), indent=2)
    sys.exit(0)
sys.exit(1)
PY
    if [[ -f "$OUT" ]]; then
      echo "Moment analysis via grok → $OUT"
      cat "$OUT"
      exit 0
    fi
  fi
fi

echo "ERROR: grok unavailable or moment analysis failed" >&2
exit 1