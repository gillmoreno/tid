#!/usr/bin/env python3
"""Build the Post Factory system prompt from biases + instructions."""

import os
import sys


def main() -> None:
    biases = os.environ.get("FACTORY_BIASES", "")
    prompt = os.environ.get("FACTORY_PROMPT", "")
    if not biases or not prompt:
        print("FACTORY_BIASES and FACTORY_PROMPT required", file=sys.stderr)
        sys.exit(1)

    text = f"""You are the Post Factory analyzer for The Idea Guy.

BIASES (Gil's lens — curiosity mixed with skepticism):
{biases}

INSTRUCTIONS:
{prompt}

TASK:
Read the transcript and return ONLY valid JSON (no markdown fences) with this shape:
{{
  "candidates": [
    {{
      "start_time": "HH:MM:SS",
      "end_time": "HH:MM:SS",
      "hook": "one-line punchy claim",
      "take": "2-3 sentences of context in Gil's voice",
      "post_text": "full post ready for X: hook line, blank line, take, blank line, Source: Podcast",
      "why_interesting": "why this moment matters",
      "confidence": 0.0
    }}
  ]
}}

Rules:
- Return 2 to 5 candidates
- Each clip 30 seconds to 5 minutes
- Accurate timestamps when possible
- No emojis, no hashtags, no engagement bait
- Speaker name in hook when known"""
    print(text)


if __name__ == "__main__":
    main()