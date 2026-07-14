# Post Factory Loop — Agent Instructions

**You are in the right folder:** `tid/loops/clip-to-post/` inside The Idea Guy monorepo.

This loop feeds the **Post Factory** app: YouTube URL → SQLite → http://localhost:5180/factory

## Default workflow (when Gil sends a podcast URL)

```bash
# From repo root — ensure API is up first
just dev   # agent runs this if not already running

# One-shot: ingest + analyze → DB + UI
./loops/clip-to-post/run-factory.sh \
  --url "https://www.youtube.com/watch?v=VIDEO_ID" \
  --podcast "All-In Podcast"
```

Then tell Gil to review candidates at **http://localhost:5180/factory**.

## Pipeline (what run-factory triggers)

```
YouTube URL
    → transcript.sh (drafts/{source-id}/)
    → analyze.sh (biases + prompt from SQLite → analysis.json)
    → Go API inserts candidates into SQLite
    → React /factory UI lists them for edit / clip / schedule
```

**Biases** and **analysis prompt** live in SQLite (editable in UI). Defaults: `biases.default.md`, `prompt.default.md`.

## Gil's posting step (semi-automated)

When a candidate is scheduled and due:

```bash
just factory-tick
# or POST /api/factory/scheduler/tick
# or ./prepare-post.sh --draft {candidate-id}
```

`prepare-post.sh` = clipboard + Chrome Default + Finder. Gil pastes and drags clip. **Do not** full-auto-post unless asked.

## Composability rules

| Piece | Role |
|-------|------|
| `biases.default.md` | Seed for Gil's lens (curiosity + skepticism) |
| `prompt.default.md` | Seed for “find 2–5 moments” instruction |
| `humanize-instructions.md` | Distilled rules from the humanize-ai-text skill. Used by rewrite when "Humanize" is chosen in the UI or instruction contains humanize/de-ai. |
| `../llm.sh` | Shared OpenAI-first provider adapter; defaults to GPT-5.6 Sol with medium reasoning and falls back to Grok. |
| `transcript.sh` | YouTube → transcript |
| `analyze.sh` | Transcript + biases + prompt → JSON candidates |
| `clip.sh` | Timestamp range → `clip.mp4` |
| `prepare-post.sh` | Mechanical post prep only |
| `meta.json` | Per-candidate source of truth for posting |
| `data/factory/tid.db` | Sources, candidates, schedule |

**Humanizing post text (removing AI smell)**

The UI "Humanize" button (and any rewrite instruction containing "humanize", "de-ai", etc.) now loads `humanize-instructions.md` and applies a full anti-AI-tell pass:
- Strips boilerplate, inflated diction, machine rhythm, hedging, em-dashes, etc.
- Forces varied sentence length and natural texture.
- Never invents facts to fake humanity.

You can also paste any post_text here and run `$humanize-ai-text` (the full skill) for an even deeper edit, then paste the result back into the factory card.

To strengthen at generation time, the analyzer prompt now explicitly asks for "natural human prose".

**LLM configuration**

Put secrets in ignored repo-root `.env.local` or `config.local.env`:

```bash
OPENAI_API_KEY=...
FACTORY_LLM_PROVIDER=openai
FACTORY_LLM_FALLBACK_PROVIDER=grok
FACTORY_OPENAI_MODEL=gpt-5.6-sol
FACTORY_OPENAI_REASONING_EFFORT=medium
```

Set `FACTORY_LLM_FALLBACK_PROVIDER=none` to fail instead of falling back to Grok.

Never commit `drafts/`, `data/`, `.env`, or `*.mp4`.

## Manual / step-by-step (same DB as UI)

```bash
just factory ingest  --url URL --podcast "All-In Podcast"
just factory analyze --source SOURCE_ID
just factory clip    --candidate CANDIDATE_ID
just factory schedule --candidate ID --at 2026-07-04T16:00:00-07:00
```

## Legacy: single-draft clip → post

For one known timestamp (no factory analysis):

```bash
./draft.sh --url URL --start T --end T --speaker NAME --podcast NAME
# edit post.txt, set meta.json status approved
./prepare-post.sh --draft ID
```

## Reference

- App agents: `../../AGENTS.md`
- Human docs: `README.md`
