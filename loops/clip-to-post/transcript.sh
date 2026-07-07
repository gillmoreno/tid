#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

URL=""
OUT_DIR="drafts"
DRAFT_ID=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --url) URL="$2"; shift 2 ;;
    --out) OUT_DIR="$2"; shift 2 ;;
    --id) DRAFT_ID="$2"; shift 2 ;;
    --slug) SLUG="$2"; shift 2 ;;
    *) echo "Unknown option $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$URL" ]]; then
  echo "Usage: $0 --url URL [--id ID | --slug slug] [--out DIR]" >&2
  exit 1
fi

ID="${DRAFT_ID:-$(default_draft_id "$URL" "${SLUG:-}")}"
mkdir -p "$OUT_DIR/$ID"

VIDEO_ID="$(echo "$URL" | sed -n 's/.*[?&]v=\([^&]*\).*/\1/p')"
[[ -n "$VIDEO_ID" ]] || VIDEO_ID="$(echo "$URL" | sed 's|.*youtu.be/||' | cut -c1-11)"

echo "Downloading auto-captions..."
yt-dlp --write-auto-sub --sub-lang en --skip-download \
  -o "$OUT_DIR/$ID/%(id)s" "$URL" 2>/dev/null || true

VTT="$OUT_DIR/$ID/${VIDEO_ID}.en.vtt"
if [[ ! -f "$VTT" ]]; then
  VTT="$(find "$OUT_DIR/$ID" -name '*.en.vtt' | head -1)"
fi

if [[ -n "$VTT" && -f "$VTT" ]]; then
  python3 - "$VTT" "$OUT_DIR/$ID/transcript.txt" <<'PY'
import re, sys
vtt_path, out_path = sys.argv[1], sys.argv[2]
ts_re = re.compile(r"^(\d{2}:\d{2}:\d{2})\.\d{3}\s+-->\s+(\d{2}:\d{2}:\d{2})\.\d{3}")
tag_re = re.compile(r"<[^>]+>")
lines = open(vtt_path, encoding="utf-8", errors="ignore").read().splitlines()
current_start = None
current_end = 0.0
text_lines = []
buckets = []

def flush():
    global text_lines, current_start, current_end
    if not text_lines or current_start is None:
        text_lines = []
        return
    text = tag_re.sub("", " ".join(text_lines))
    text = text.replace("&gt;", ">").replace("&lt;", "<").replace("&amp;", "&").strip()
    text_lines = []
    if len(text) < 3:
        return
    if buckets and current_start - buckets[-1][0] <= 2.5:
        if len(text) > len(buckets[-1][1]):
            buckets[-1] = (buckets[-1][0], text)
    else:
        buckets.append((current_start, text))

for line in lines:
    line = line.strip()
    if not line or line.startswith("WEBVTT") or line.startswith("Kind:") or line.startswith("Language:"):
        continue
    m = ts_re.match(line)
    if m:
        flush()
        h, mi, s = map(int, m.group(1).split(":"))
        current_start = h * 3600 + mi * 60 + s
        eh, emi, es = map(int, m.group(2).split(":"))
        current_end = eh * 3600 + emi * 60 + es
        continue
    if "-->" in line or line.startswith("align:") or line.startswith("position:"):
        continue
    if current_end - (current_start or 0) < 0.15:
        continue
    text_lines.append(line)
flush()

def fmt(sec):
    h = sec // 3600
    m = (sec % 3600) // 60
    s = sec % 60
    return f"{h:02d}:{m:02d}:{s:02d}"

out = [f"[{fmt(start)}] {text}" for start, text in buckets]
open(out_path, "w", encoding="utf-8").write("\n".join(out) + ("\n" if out else ""))
PY
  echo "Transcript saved: $OUT_DIR/$ID/transcript.txt"
else
  echo "No captions found. Add transcript manually or use whisper."
  touch "$OUT_DIR/$ID/transcript.txt"
fi

echo "$ID"