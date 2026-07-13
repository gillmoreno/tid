# Post Factory — Analysis Instruction

Find **3 to 7** interesting moments in this podcast transcript. Return as many strong candidates as exist — up to 7. Do not stop at 3 by default.

## Coverage (critical)

- Read the **entire** transcript — lines are timestamped `[HH:MM:SS]`
- Spread picks across the **full episode**, not just the opening
- For episodes over 20 minutes: at least **2 candidates from the first third**, **2 from the middle third**, and **2 from the final third**
- Never cluster more than 2 candidates in the first 5 minutes unless the episode is under 10 minutes total

## What to prioritize (Gil's interests)

Weight heavily toward moments about:
- **AI** — models, agents, labs, scaling, timelines, architecture bets
- **Compute** — chips, Moore's Law, infrastructure, who wins the stack
- **Money** — markets, venture, unit economics, valuations, who's making/losing money
- **Pragmatic builder insight** — mechanisms, tradeoffs, concrete numbers, falsifiable claims

Bonus points for clips that are **wild, bold, and attention-catching** — as long as they're grounded in what was actually said.

## For each moment

- Pick a clip between **30 seconds and 5 minutes**
- `start_time` and `end_time` must bracket the timestamped lines your post is based on
- Choose segments that stand alone without needing the full episode
- Pick **Format A** (build the case), **Format B** (quote excerpt), or **Format C** (Gil's commentary take) — whichever fits the clip
- Prefer **Format C** when the moment invites Gil's opinion — skeptical pushback, agreement with a twist, or "here's what this actually means"
- Write **post_text** ready to paste on X:
  - Tag people/companies mentioned using @ handles from the MENTIONS dictionary
  - Format A: staccato beats, reframe → proof → pattern → closer/question
  - Format B: topic header optional, then tight quote in speaker's voice
  - End with podcast @ handle only (e.g. @theallinpod) — **never** a YouTube URL
  - Natural human prose only: vary sentence length, avoid uniform rhythm, inflated diction, boilerplate, or AI tells. Sound like a careful person wrote it.
- Explain **why_interesting** in one sentence (for Gil's review UI)
- Set **confidence** 0.0–1.0 (how strong this clip is for posting)

Skip weak moments. Prefer bold, post-worthy clips over safe summaries.