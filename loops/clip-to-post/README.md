# Clip → Post

Semi-automated pipeline for Podcast Alpha-style native X video posts from AI/tech YouTube clips.

**Philosophy:** composable pieces, no monolith. Change taste rules without touching prepare-post. Change prepare steps without touching copy generation.

## Architecture

| Piece | Role |
|-------|------|
| `meta.json` | Single source of truth per draft (text, clip path, speaker, source, status) |
| `taste.md` | What to select + how to write |
| `clip.sh` | YouTube URL + timestamps → `clip.mp4` |
| `transcript.sh` | YouTube URL → `transcript.txt` |
| `draft-copy.sh` | Transcript segment → `post.txt` (follows `taste.md`) |
| `draft.sh` | Orchestrator: runs clip + transcript + copy → draft folder |
| **`prepare-post.sh`** | **Dumb final step: clipboard + Chrome + Finder** (reads `meta.json` only) |
| `post_to_x.py` | Optional full browser automation (Playwright) — use only if you want hands-off post |

## Directory layout

```
loops/clip-to-post/
├── README.md              ← you are here (single source of truth)
├── taste.md               ← selection + voice rules
├── lib.sh                 ← shared helpers (sourced by scripts)
├── config.example.env     ← X API keys template (copy to .env, never commit)
├── clip.sh
├── transcript.sh
├── draft-copy.sh
├── draft.sh
├── prepare-post.sh        ← semi-automated posting prep
├── post_to_x.py           ← optional full automation
└── drafts/                ← gitignored per-draft output
    └── {id}/
        ├── meta.json
        ├── clip.mp4
        ├── transcript.txt
        └── post.txt
```

## Workflow (happy path)

### 1. Create a draft

```bash
cd loops/clip-to-post

./draft.sh \
  --url "https://www.youtube.com/watch?v=VIDEO_ID" \
  --start "00:54:05" \
  --end "00:54:45" \
  --speaker "Naval" \
  --podcast "Naval Podcast" \
  --slug "naval-taiwan-competition"
```

`--slug` is optional; default id is `{date}-{videoId}`.

Edit `drafts/{id}/post.txt` by hand if needed, then sync into `meta.json`:

```bash
# After editing post.txt, update meta.json post_text field manually or re-run draft-copy
```

Set `"status": "approved"` in `meta.json` when ready to post.

### 2. Prepare to post (semi-automated — preferred)

```bash
./prepare-post.sh 20260704-naval-taiwan-competition
```

This does exactly three things:

1. **Copies** `post_text` from `meta.json` to clipboard
2. **Opens** Chrome Default profile on https://x.com/compose/post
3. **Opens** Finder on the draft folder

You finish manually: **Cmd+V**, drag `clip.mp4`, click Post.

### 3. Optional: full browser automation

```bash
pip install playwright && playwright install chromium
python post_to_x.py --draft 20260704-naval-taiwan-competition
```

Uses logged-in Chrome Default profile via Playwright. Only when you want zero manual steps.

## meta.json schema

```json
{
  "id": "20260704-naval-taiwan-competition",
  "source_url": "https://www.youtube.com/watch?v=...",
  "speaker": "Naval",
  "podcast": "Naval Podcast",
  "start": "00:54:05",
  "end": "00:54:45",
  "status": "draft | approved | posted",
  "post_text": "Naval: ...\n\nContext...\n\nSource: Naval Podcast",
  "clip_path": "drafts/20260704-naval-taiwan-competition/clip.mp4",
  "created_at": "2026-07-04T13:45:00Z"
}
```

`prepare-post.sh` only reads `post_text` and the clip folder. Extra fields (`source_url`, etc.) are for your records and future automation.

## Prerequisites

```bash
brew install yt-dlp ffmpeg jq
```

## Git rules

- **Tracked:** scripts, `taste.md`, `README.md`, `config.example.env`
- **Never committed:** `drafts/`, `.env`, `*.mp4`, `config.local.env`

## Phase status

- [x] Phase 0: Manual pipeline proven (Naval Taiwan clip)
- [x] Phase 1: Core scripts (`clip`, `transcript`, `draft-copy`, `draft`)
- [x] Semi-automated prepare flow (`prepare-post.sh`)
- [ ] Phase 2: LLM-assisted copy via `taste.md` prompt in `draft-copy.sh`
- [ ] Phase 3: Go API draft queue in TID backend (optional)
- [ ] Phase 4: X API posting (`post.sh`, approve-gated)
- [ ] Phase 5: Scheduled / agent-driven loop

## For AI agents

**Default posting flow for Gil:** semi-automated `prepare-post.sh` — NOT full auto-post unless explicitly asked.

When a draft is approved, run:

```bash
./prepare-post.sh --draft <id>
```

Do not reinvent the workflow. This README is the single source of truth.