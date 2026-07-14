#!/usr/bin/env python3
"""Write a dev-fallback analysis.json when no LLM provider is available."""

import json
import sys


def main() -> None:
    out_path, snippet = sys.argv[1], sys.argv[2]
    data = {
        "candidates": [
            {
                "start_time": "00:01:00",
                "end_time": "00:01:45",
                "post_text": f"{snippet}\n\nPlaceholder — re-run with an LLM provider configured, or edit in the Post Factory UI.\n\n@theallinpod",
                "why_interesting": "Dev fallback candidate (LLM unavailable)",
                "confidence": 0.3,
            }
        ]
    }
    with open(out_path, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2)


if __name__ == "__main__":
    main()
