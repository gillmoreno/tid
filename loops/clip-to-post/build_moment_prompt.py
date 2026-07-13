#!/usr/bin/env python3
"""Build system prompt for spot-moment analysis (3 takes at different lengths)."""

import os
import sys


def main() -> None:
    biases = os.environ.get("FACTORY_BIASES", "")
    mentions = os.environ.get("FACTORY_MENTIONS", "")
    focus_note = os.environ.get("FACTORY_FOCUS_NOTE", "")
    window_start = os.environ.get("FACTORY_WINDOW_START", "")
    window_end = os.environ.get("FACTORY_WINDOW_END", "")

    if not biases or not window_start or not window_end:
        print("FACTORY_BIASES, FACTORY_WINDOW_START, and FACTORY_WINDOW_END required", file=sys.stderr)
        sys.exit(1)

    has_take = bool(focus_note.strip())
    note_block = focus_note.strip() or "(Gil did not add a take — infer a skeptical-curious angle from the transcript.)"

    commentary_rule = (
        "- Candidate 2 (medium) must use Format C (Gil's commentary): weave Gil's take into the post — "
        "his opinion first, then anchor in what the speaker said. Do not just quote the segment.\n"
        if has_take
        else "- Candidate 2 (medium) must use Format C (Gil's commentary) — infer Gil's skeptical-curious angle from the transcript.\n"
    )

    text = f"""You are the Post Factory spot-moment analyzer for The Idea Guy.

Gil selected an exact transcript range while listening to a podcast.

BIASES (Gil's lens):
{biases}

MENTIONS DICTIONARY:
{mentions}

SPOT MOMENT:
- Transcript range: {window_start} → {window_end} (read the full segment below)
- Gil's take (optional): {note_block}

TASK:
Return ONLY valid JSON (no markdown fences):
{{
  "candidates": [
    {{
      "start_time": "HH:MM:SS",
      "end_time": "HH:MM:SS",
      "post_text": "full post ready for X with @ tags; end with podcast @ only",
      "why_interesting": "length tier + angle (e.g. short take — skeptical reframe)",
      "confidence": 0.0
    }}
  ]
}}

Rules:
- Return exactly 3 candidates — same video clip ({window_start} → {window_end}) for ALL three; only post_text LENGTH and angle differ:
  1. SHORT post — punchy hook, 2–4 lines max, one sharp claim or reframe
  2. MEDIUM post — fuller take, ~6–10 lines, mechanism + implication
  3. LONG post — full essay beats (Format A or C), ~10–15 lines, evidence ladder or deep commentary
- Set start_time to {window_start} and end_time to {window_end} on every candidate (identical clip)
- All posts must draw from this transcript range only — do not invent facts
{commentary_rule}- Other candidates may use Format A (essay) or Format B (quote) as appropriate
- Formats A/B: stay faithful to what is said. Format C: Gil's opinion is allowed; facts must come from the clip
- Prioritize AI, compute, money, pragmatic insight when relevant
- Tag people/companies from MENTIONS; end with podcast @ only — never YouTube URL
- No emojis, no hashtags"""
    print(text)


if __name__ == "__main__":
    main()