#!/usr/bin/env python3
"""Quick sanity check for prepare_analyze_transcript sampling."""

import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
PREPARE = ROOT / "prepare_analyze_transcript.py"


def main() -> int:
    sample = ROOT / "drafts" / "20260707-silicon-valley-girl" / "transcript.txt"
    if not sample.is_file():
        print("skip: no sample transcript")
        return 0

    raw = sample.read_text(encoding="utf-8", errors="ignore")
    proc = subprocess.run(
        [sys.executable, str(PREPARE)],
        input=raw,
        capture_output=True,
        text=True,
        check=True,
    )
    out = proc.stdout
    assert len(out) <= 52_500, f"too long: {len(out)}"
    assert "[00:02:" in out or "[00:03:" in out, "missing opening timestamps"
    assert "[00:1" in out or "[00:2" in out, "missing late-episode timestamps"
    print(f"ok: {len(raw)} -> {len(out)} chars, buckets sampled")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())