# TID — Grok Build / Agent Instructions

**Repo:** The Idea Guy (`tid`) — Go API + React + SQLite + composable loops.

**Local app:** `just dev` → http://localhost:5180
**Factory UIs (local only, `VITE_FACTORY_ENABLED=true`):**
- Podcast clips: http://localhost:5180/factory
- Articles (article URL → X posts): http://localhost:5180/factory/articles
- Ideas backlog: http://localhost:5180/factory/ideas
- Sources dictionary (podcasts / news feeds / companies / people): http://localhost:5180/factory/sources

Article Factory loop: `loops/article-to-post/` (needs `pip3 install -r loops/article-to-post/requirements.txt` for trafilatura).

## Primary loop: Post Factory

When Gil gives you a **YouTube podcast URL** to turn into clip candidates + posts:

1. **Ensure dev is running** (you run this — do not tell Gil to run it):
   ```bash
   just dev
   ```
   Or check: `curl -sf http://localhost:8080/api/factory/biases`

2. **Run the factory loop** from repo root:
   ```bash
   ./loops/clip-to-post/run-factory.sh --url "YOUTUBE_URL" --podcast "All-In Podcast"
   ```

3. **Tell Gil** to open http://localhost:5180/factory — candidates are in SQLite and the UI. He edits, clips, schedules there.

4. **At post time:** semi-automated `prepare-post` (clipboard + Chrome + Finder). Do not full-auto-post unless Gil asks.

### What gets written where

| Layer | Path |
|-------|------|
| SQLite | `data/factory/tid.db` |
| Biases + prompt | DB (seeded from `loops/clip-to-post/biases.default.md`, `prompt.default.md`) |
| Transcript / analysis artifacts | `loops/clip-to-post/drafts/{source-id}/` (gitignored) |
| Clips + meta for posting | `loops/clip-to-post/drafts/{candidate-id}/` |

### Factory CLI (same DB as UI)

```bash
just factory ingest  --url URL --podcast "All-In Podcast"
just factory analyze --source SOURCE_ID
just factory clip    --candidate CANDIDATE_ID
just factory schedule --candidate ID --at 2026-07-04T16:00:00-07:00
just factory-tick
```

### HTTP API (alternative)

Base: `http://localhost:8080/api/factory/`

- `POST /sources` → ingest
- `POST /sources/{id}/analyze` → transcript + analyze → candidates in DB
- `GET /candidates?source_id=` → list
- `POST /candidates/{id}/clip` → clip video + meta.json
- `POST /candidates/{id}/schedule` → `{"scheduled_at":"RFC3339"}`
- `POST /scheduler/tick` → due posts → `prepare-post.sh`

## Loop detail

Full script + composability rules: `loops/clip-to-post/AGENTS.md` and `loops/clip-to-post/README.md`.

## Do not

- Commit `data/`, `drafts/`, `.env`, or video files
- Expose Post Factory in production (no `VITE_FACTORY_ENABLED` on Cloudflare)
- Skip `prepare-post` semi-automation for posting unless Gil explicitly wants full auto
