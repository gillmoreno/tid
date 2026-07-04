# Hermes Agent Mission: Loops with Taste — Clip → Post

**Copy everything below the line into Hermes as your task prompt.**

---

## MISSION

Build an end-to-end **Clip → Post** pipeline for **The Idea Guy (TID)** that replicates the Podcast Alpha / X clip-post format:

1. Take a **YouTube podcast URL** Gil finds interesting
2. **Extract a short clip** (30–90 seconds) with the best moment
3. **Write post copy** in Gil's voice (hook + context + attribution)
4. **Prepare a native X video post** (mp4 upload — NOT a YouTube link card)
5. Stop at **draft/approve** before publishing (Phase 1). Automate posting only after Gil approves the first manual run.

**Repo:** `/Users/gilbertomorenocruz/Desktop/Projects/tid`  
**GitHub:** https://github.com/gillmoreno/tid  
**Brand:** The Idea Guy — signal vs noise, one builder, real loops, no hype  
**Section:** Loops with Taste → `clip-to-post`

---

## CONTEXT (read first)

### What Podcast Alpha-style posts actually are

- **Text:** 1–3 sentences. Bold hook naming speaker + claim. Short context. Optional "Show more" length is fine.
- **Video:** A **native X-uploaded mp4** — NOT an embedded YouTube player. They download/stream-cut from YouTube, optionally add captions/date stamp, upload to X.
- **Source:** Attribution in text (speaker, podcast, link in reply or text).

### What TID already has

- Go + React monorepo (Roger pattern)
- `go-backend/internal/api/content.go` — seed content for sections/items
- `frontend/` — brand site at the-idea-guy.com
- Item slug: `clip-to-post` under `loops-with-taste`

### Gil's taste filter (encode in prompts)

- AI / tech / builder podcasts (All-In, Dwarkesh, Lex, etc.)
- Moments that are **contrarian, specific, or actionable** — not generic hype
- Copy tone: direct, sharp, no emojis, no "🧵", no LinkedIn slop
- Prefer **one clear claim** per clip
- Do NOT expose the full automation recipe publicly — the hook is the output

---

## PHASE 0 — Prove the manual pipeline (do this FIRST)

Before any automation, run one clip end-to-end by hand and document exact commands that worked.

### 0.1 Prerequisites (install if missing)

```bash
brew install yt-dlp ffmpeg jq
# Optional: whisper for transcript if YT captions unavailable
```

### 0.2 Manual clip extraction

```bash
URL="https://www.youtube.com/watch?v=VIDEO_ID"
START="00:04:01"   # from transcript review
END="00:04:50"

# Get direct stream URL (no full download)
STREAM=$(yt-dlp -g -f "bv*+ba/b" "$URL" | head -1)

# Cut clip
mkdir -p data/clips
ffmpeg -ss "$START" -to "$END" -i "$STREAM" \
  -c:v libx264 -c:a aac -movflags +faststart \
  "data/clips/$(date +%Y%m%d)-clip.mp4"
```

### 0.3 Transcript for moment-finding

```bash
# Prefer YouTube auto-captions
yt-dlp --write-auto-sub --sub-lang en --skip-download -o "data/clips/%(id)s" "$URL"
# Or: whisper data/clips/clip.mp4 (if installed)
```

### 0.4 Verify output

- Clip plays in QuickTime/VLC
- Duration 30–90s
- Audio clear
- File size < 512MB (X limit)

**Deliverable:** `data/clips/README.md` with the exact commands that worked on macOS for one real video.

---

## PHASE 1 — Build `loops/clip-to-post/` in TID repo

Create this structure:

```
tid/
├── loops/
│   └── clip-to-post/
│       ├── README.md
│       ├── taste.md              # Gil's selection + voice rules
│       ├── config.example.env    # X API keys, paths (no secrets committed)
│       ├── clip.sh               # URL + start + end → mp4
│       ├── transcript.sh         # URL → transcript text/json
│       ├── draft-copy.sh         # transcript excerpt → post text (LLM or template)
│       ├── draft.sh              # orchestrator: url → clip + copy + metadata json
│       └── drafts/               # gitignored output per run
│           └── {id}/
│               ├── meta.json
│               ├── clip.mp4
│               ├── transcript.txt
│               └── post.txt
```

### `meta.json` schema

```json
{
  "id": "20260704-friedberg-anthropic",
  "source_url": "https://youtube.com/watch?v=...",
  "speaker": "David Friedberg",
  "podcast": "All-In Podcast",
  "start": "00:04:01",
  "end": "00:04:50",
  "status": "draft",
  "post_text": "...",
  "clip_path": "drafts/.../clip.mp4",
  "created_at": "ISO8601"
}
```

### `draft.sh` usage (target)

```bash
./loops/clip-to-post/draft.sh \
  --url "YOUTUBE_URL" \
  --start "00:04:01" \
  --end "00:04:50" \
  --speaker "David Friedberg" \
  --podcast "All-In Podcast"
```

Or with AI-assisted moment pick (Phase 2):

```bash
./loops/clip-to-post/draft.sh --url "YOUTUBE_URL" --auto-select
```

---

## PHASE 2 — Copy generation (taste)

### Post text template

```
{Speaker} @{handle if known}: {bold one-line hook}

{2-3 sentences of context — what was said, why it matters, no fluff}

Source: {Podcast name} · {optional YT link for replies, not main post}
```

### Example (Podcast Alpha style)

```
David Friedberg @friedberg: Anthropic is trying to commoditize everyone's business.

Friedberg, a life-sciences CEO and All-In host, on Anthropic's enterprise pitch to biotech: share your proprietary data, get model access. Nearly every company he spoke with reached the same conclusion.

Source: All-In Podcast
```

### Rules for the copy-writer (LLM prompt in `taste.md`)

- Max 280 chars for hook line if possible; full post can be longer (X premium)
- No hashtags unless Gil adds them manually
- No "What do you think?" engagement bait
- Name the speaker in line 1
- Quote or paraphrase accurately from transcript segment

---

## PHASE 3 — Draft queue API (optional, in Go backend)

Add to `go-backend/internal/api/`:

| Endpoint | Purpose |
|----------|---------|
| `GET /api/loops/clip-to-post/drafts` | List drafts |
| `GET /api/loops/clip-to-post/drafts/{id}` | One draft + metadata |
| `POST /api/loops/clip-to-post/drafts` | Register new draft folder |
| `PATCH /api/loops/clip-to-post/drafts/{id}` | Update status: draft → approved → posted |

Store draft metadata in SQLite or JSON files under `data/loops/clip-to-post/`.

Frontend (later): simple admin page under `/loops/clip-to-post` to preview clip + text and click Approve.

---

## PHASE 4 — X posting (ONLY after Gil approves first draft)

### Requirements

- X API v2 access with **tweet write** + **media upload** scopes
- Env vars in `loops/clip-to-post/.env` (gitignored):
  - `X_API_KEY`
  - `X_API_SECRET`
  - `X_ACCESS_TOKEN`
  - `X_ACCESS_TOKEN_SECRET`

### Media upload flow

1. `POST /2/media/upload` (chunked if > 5MB)
2. `POST /2/tweets` with `{ "text": "...", "media": { "media_ids": ["..."] } }`

### `post.sh` (approve-gated)

```bash
./loops/clip-to-post/post.sh --draft-id 20260704-friedberg-anthropic
# Only works if meta.json status == "approved"
```

**Do NOT auto-post in Phase 1.** Always require `status: approved`.

---

## PHASE 5 — The loop (later)

Once Phases 0–4 work:

1. Gil drops a YouTube URL (Telegram/CLI/email to Hermes)
2. Hermes runs `draft.sh --auto-select` or with timestamps
3. Hermes notifies Gil with preview (clip path + post text)
4. Gil replies "approve" → Hermes runs `post.sh`
5. Optional: cron / Hermes scheduled task for watchlist channels

---

## HERMES EXECUTION RULES

1. **Work in** `/Users/gilbertomorenocruz/Desktop/Projects/tid`
2. **Commit incrementally** — one phase per commit, Gil's format:
   ```
   [loops]
   + <what changed>
   ```
3. **Never commit** `.env`, `data/clips/*`, `drafts/*` with video files
4. **Test each script** on a real YouTube URL before moving on
5. **Report after each phase:** what worked, what failed, exact commands
6. **Ask Gil** for:
   - First YouTube URL to use as test
   - X API credentials when reaching Phase 4
   - Approve before first real post
7. Use **headless Grok** for code generation if needed:
   ```bash
   grok --no-auto-update --always-approve -p "..." --cwd /Users/gilbertomorenocruz/Desktop/Projects/tid
   ```

---

## SUCCESS CRITERIA

- [ ] Phase 0: One real clip extracted to `data/clips/`
- [ ] Phase 1: `draft.sh` produces `meta.json` + `clip.mp4` + `post.txt`
- [ ] Phase 2: Post copy matches Podcast Alpha format and Gil's taste
- [ ] Phase 3: Draft listable via API (optional)
- [ ] Phase 4: Approved draft posts to X with inline video
- [ ] Phase 5: Hermes can run draft → notify → approve → post loop

---

## START COMMAND FOR HERMES

```
Read /Users/gilbertomorenocruz/Desktop/Projects/tid/docs/HERMES-CLIP-TO-POST.md and docs/CONCEPT.md.

Execute Phase 0 first: pick a test YouTube URL (ask me if none provided), extract one 30-90s clip with yt-dlp + ffmpeg, pull transcript, write draft post copy in my taste, save everything under loops/clip-to-post/drafts/.

Do NOT post to X yet. Show me the draft post text and clip path for approval before Phase 4.
```