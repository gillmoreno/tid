#!/usr/bin/env python3
"""Build the Post Factory system prompt from biases + instructions + mentions."""

import os
import sys


def main() -> None:
    biases = os.environ.get("FACTORY_BIASES", "")
    prompt = os.environ.get("FACTORY_PROMPT", "")
    mentions = os.environ.get("FACTORY_MENTIONS", "")
    if not biases or not prompt:
        print("FACTORY_BIASES and FACTORY_PROMPT required", file=sys.stderr)
        sys.exit(1)

    text = f"""You are the Post Factory analyzer for The Idea Guy.

BIASES (Gil's lens — curiosity mixed with skepticism):
{biases}

INSTRUCTIONS:
{prompt}

MENTIONS DICTIONARY (tag these @ handles when names appear — podcast selection maps to attribution):
{mentions}

TASK:
Read the transcript and return ONLY valid JSON (no markdown fences) with this shape:
{{
  "candidates": [
    {{
      "start_time": "HH:MM:SS",
      "end_time": "HH:MM:SS",
      "post_text": "full post ready for X with @ tags; end with podcast @ only (never YouTube URL)",
      "why_interesting": "one sentence for Gil's review — format choice, angle, why this clip",
      "confidence": 0.0
    }}
  ]
}}

Rules:
- Return 2 to 5 candidates
- Each clip 30 seconds to 5 minutes
- Transcript lines are prefixed with [HH:MM:SS] — start_time and end_time MUST bracket the lines your post_text is based on
- post_text must only describe what is actually said between start_time and end_time (no mixing topics from other timestamps)
- Choose Format A (essay) or Format B (quote) per clip
- Tag people/companies from MENTIONS dictionary when referenced
- End post_text with podcast @ handle — not a YouTube link
- No emojis, no hashtags"""
    print(text)


if __name__ == "__main__":
    main()