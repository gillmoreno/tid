#!/bin/bash
set -euo pipefail

# analyze.sh — biases + prompt + article text → analysis.json (post candidates)
# Uses grok headless when available; otherwise a heuristic fallback for dev.
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
PUBLICATION="$(jq -r '.publication // ""' "$INPUT")"
ARTICLE="$(jq -r '.article' "$INPUT")"

OUT_DIR="$(dirname "$OUT")"
mkdir -p "$OUT_DIR"

export FACTORY_BIASES="$BIASES"
export FACTORY_PROMPT="$PROMPT"
export FACTORY_MENTIONS="$MENTIONS"
export FACTORY_PUBLICATION="$PUBLICATION"
SYSTEM_PROMPT="$(python3 "$SCRIPT_DIR/build_system_prompt.py")"

if command -v grok >/dev/null 2>&1 && [[ -f "$HOME/.grok/auth.json" || -n "${XAI_API_KEY:-}" ]]; then
  # Cap the article length so we do not blow the prompt budget.
  TRIMMED_ARTICLE="$(printf '%s' "$ARTICLE" | head -c 24000)"
  PROMPT_TEXT="${SYSTEM_PROMPT}

ARTICLE:
${TRIMMED_ARTICLE}"

  grok --no-auto-update -p "$PROMPT_TEXT" --output-format plain > "$OUT_DIR/analyze-raw.txt" 2>/dev/null || true
  if [[ -s "$OUT_DIR/analyze-raw.txt" ]]; then
    python3 - "$OUT_DIR/analyze-raw.txt" "$OUT" <<'PY'
import json, re, sys
raw = open(sys.argv[1]).read()
m = re.search(r'\{.*\}', raw, re.S)
if m:
    data = json.loads(m.group(0))
    if data.get("candidates"):
        json.dump(data, open(sys.argv[2], "w"), indent=2)
        sys.exit(0)
sys.exit(1)
PY
    if [[ -f "$OUT" ]]; then
      echo "Article analysis via grok → $OUT"
      cat "$OUT"
      exit 0
    fi
  fi
fi

# Dev fallback: one candidate from the article's opening so the loop still works.
python3 - "$OUT" "$PUBLICATION" <<'PY'
import json, sys
out, publication = sys.argv[1], (sys.argv[2] or "").strip()
try:
    article = json.load(open(out.rsplit("/", 1)[0] + "/article.json"))
    text = article.get("text", "")
    title = article.get("title", "")
except Exception:
    text, title = "", ""
snippet = " ".join(text.split())[:220]
tag = ""  # fallback keeps attribution to the Go layer
post = snippet if snippet else (title or "New article ingested.")
data = {
    "title": title,
    "candidates": [
        {
            "post_text": post,
            "why_interesting": "Dev fallback candidate (grok unavailable) — edit or re-analyze with grok configured.",
            "confidence": 0.2,
        }
    ],
}
json.dump(data, open(out, "w"), indent=2)
print("Article analysis fallback →", out)
PY

cat "$OUT"
