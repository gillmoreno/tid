# Article Factory — Analysis Instruction

Read this written article and extract **3 to 7** standalone X posts. Return as many strong candidates as the article supports — up to 7. Do not stop at 3 by default.

## What each candidate is

- A single, self-contained X post — no threads, no timestamps
- Each post draws on a **distinct** idea, claim, statistic, or tension from the article
- Do not repeat the same angle across candidates

## What to prioritize (Gil's interests)

Weight heavily toward angles about:
- **AI** — models, agents, labs, scaling, timelines, architecture bets
- **Compute** — chips, infrastructure, who wins the stack
- **Money** — markets, venture, unit economics, valuations, who's making/losing money
- **Pragmatic builder insight** — mechanisms, tradeoffs, concrete numbers, falsifiable claims

Bonus points for posts that are **bold and attention-catching** — as long as they're grounded in what the article actually says.

## For each post

- Write **post_text** ready to paste on X:
  - Lead with the sharpest hook — the reframe, the number, the tension
  - Tag people/companies mentioned using @ handles from the MENTIONS dictionary
  - Keep Gil's skeptical-curious voice: what would have to be true for this to matter?
  - End with the publication @ handle (from news_feeds) — never a raw URL
- Explain **why_interesting** in one sentence (for Gil's review UI)
- Set **confidence** 0.0–1.0 (how strong this post is)

Stay faithful to the article. Do not invent facts or numbers. No emojis, no hashtags.
