#!/bin/bash
set -euo pipefail

# analyze.sh — biases + prompt + transcript → analysis.json
# Uses grok headless when available; otherwise heuristic fallback for dev.
#
# Usage: ./analyze.sh --input drafts/{source}/analyze-input.json --out drafts/{source}/analysis.json

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
PROMPT="$(jq -r '.prompt' "$INPUT")"
MENTIONS="$(jq -r '.mentions // ""' "$INPUT")"
TRANSCRIPT="$(jq -r '.transcript' "$INPUT")"

OUT_DIR="$(dirname "$OUT")"
mkdir -p "$OUT_DIR"

export FACTORY_BIASES="$BIASES"
export FACTORY_PROMPT="$PROMPT"
export FACTORY_MENTIONS="$MENTIONS"
SYSTEM_PROMPT="$(python3 "$SCRIPT_DIR/build_system_prompt.py")"

if command -v grok >/dev/null 2>&1 && [[ -f "$HOME/.grok/auth.json" || -n "${XAI_API_KEY:-}" ]]; then
  PROMPT_TEXT="${SYSTEM_PROMPT}

TRANSCRIPT (truncated to 12000 chars):
${TRANSCRIPT:0:12000}"

  grok --no-auto-update -p "$PROMPT_TEXT" --output-format plain > "$OUT_DIR/analyze-raw.txt" 2>/dev/null || true
  if [[ -s "$OUT_DIR/analyze-raw.txt" ]]; then
    python3 - "$OUT_DIR/analyze-raw.txt" "$OUT" <<'PY'
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
      echo "Analysis via grok → $OUT"
      cat "$OUT"
      exit 0
    fi
  fi
fi

# Dev fallback: single candidate from first substantive chunk
HOOK="$(echo "$TRANSCRIPT" | tr '\n' ' ' | sed 's/  */ /g' | cut -c1-100)"
python3 "$SCRIPT_DIR/analyze_fallback.py" "$OUT" "$HOOK"

echo "Analysis fallback → $OUT"
cat "$OUT"