#!/usr/bin/env python3
"""Write a dev-fallback analysis.json when grok is unavailable."""

import json
import sys


def main() -> None:
    out_path, hook = sys.argv[1], sys.argv[2]
    data = {
        "candidates": [
            {
                "start_time": "00:01:00",
                "end_time": "00:01:45",
                "hook": hook,
                "take": "Placeholder take — re-run with grok login for real analysis, or edit in the Post Factory UI.",
                "post_text": f"Speaker: {hook}\n\nPlaceholder take — edit before posting.\n\nSource: Podcast",
                "why_interesting": "Dev fallback candidate (grok unavailable)",
                "confidence": 0.3,
            }
        ]
    }
    with open(out_path, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2)


if __name__ == "__main__":
    main()