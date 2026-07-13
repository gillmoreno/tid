# Article Factory + Sources dictionary page

_Date: 2026-07-08_

Two additions to the admin factory suite:

1. **Article Factory** — a new factory that turns a written-article URL into multiple standalone X post candidates (text only, no video).
2. **Sources page** — a dedicated, table-based UI for managing the `@` mentions/sources dictionary across four categories: Podcasts, News feeds, Companies, People. Replaces raw-JSON editing (raw JSON kept as an advanced fallback).

Both live under `/factory` and are gated by `VITE_FACTORY_ENABLED` (local-only, never exposed in production).

---

## 1. Article Factory

Route: **`/factory/articles`**

### Flow

```
Article URL + publication
  → extract clean text (trafilatura)
  → AI analysis (shared biases + article prompt + dictionary)
  → 3–7 standalone X post candidates
  → edit / AI-refine per candidate
  → schedule or post now (clipboard + Chrome compose)
```

Mirrors the podcast clip factory but has **no** transcript, clipping, or trimming — each candidate is one self-contained post. One article produces multiple distinct post candidates.

### What's editable / shared

- **Biases** and the **`@` dictionary** are shared with the clip factory (edit in Sources).
- The **article analysis prompt** is separate and editable in the Article Factory settings panel. Stored in `prompt_templates` under the name `article-analysis` (kept `is_active = 0` so it never collides with the clip factory's single active prompt).

### Extraction dependency

Article text extraction uses **trafilatura**:

```bash
pip3 install -r loops/article-to-post/requirements.txt
```

Falls back to a basic urllib + HTML-strip extractor when trafilatura is unavailable.

### Backend

| Piece | Path |
|-------|------|
| Tables | `article_sources`, `article_candidates` in `go-backend/internal/db/schema.go` |
| Store | `go-backend/internal/factory/articles.go` |
| Runner | `go-backend/internal/factory/article_runner.go` (+ `ArticleLoopsDir` in `runner.go`) |
| Handlers | `go-backend/internal/api/handlers_articles.go` |
| Loop scripts | `loops/article-to-post/` |

New tables are additive (`CREATE TABLE IF NOT EXISTS`), so existing DBs migrate cleanly on boot. Scheduling lives directly on `article_candidates` (`scheduled_at`/`prepared_at`) rather than the clip factory's `scheduled_posts` table, to keep the video path untouched.

### HTTP API (`/api/factory/articles/…`)

| Method | Route | Purpose |
|--------|-------|---------|
| GET/PUT | `/articles/prompt` | Editable article analysis prompt |
| GET/POST | `/articles/sources` | List / ingest article sources |
| GET/DELETE | `/articles/sources/{id}` | Get / delete a source |
| POST | `/articles/sources/{id}/analyze` | Extract + analyze → candidates |
| GET | `/articles/candidates?source_id=` | List candidates |
| GET/PATCH/DELETE | `/articles/candidates/{id}` | Candidate CRUD |
| POST | `/articles/candidates/{id}/rewrite` | AI refine |
| POST | `/articles/candidates/{id}/schedule` | Schedule (RFC3339) |
| POST | `/articles/candidates/{id}/post-now` | Prepare post now |
| GET | `/articles/scheduled` | Scheduled queue |
| POST | `/articles/scheduler/tick` | Run due posts |

### Frontend

- Page: `frontend/pages/ArticleFactoryPage.tsx`
- Components: `frontend/components/factory/article/*`
- API: `frontend/api/articles.ts`
- Types: `frontend/types/factory.ts` (`ArticleSource`, `ArticleCandidate`, …)

---

## 2. Sources page

Route: **`/factory/sources`**

A prettier, organized replacement for editing the mentions dictionary as raw JSON. Renders four editable tables:

| Category | Used for |
|----------|----------|
| **Podcasts** | Clip factory ingest dropdown + `@` attribution |
| **News feeds** | Article factory ingest dropdown + `@` attribution _(new category)_ |
| **Companies** | `@` tagging in posts |
| **People** | `@` tagging in posts |

Each row has Name, `@` Handle, Aliases (comma-separated), and (for News feeds / Companies) an optional URL. A sticky Save bar writes the whole dictionary; a collapsible **Raw JSON** panel remains as an advanced fallback.

### Dictionary changes

- `MentionDictionary` gained a `news_feeds` category and each `MentionEntry` gained an optional `url` (`go-backend/internal/factory/mentions.go`). Storage is unchanged (single JSON blob in `mention_dictionaries`), so no schema migration is required.
- Existing dictionaries without `news_feeds` are backfilled with defaults on boot (`Store.EnsureDictionaryNewsFeeds`).
- New endpoint `GET /api/factory/publications` returns the news-feed options for the article ingest dropdown (parallel to `/factory/podcasts`).
- Seed defaults in `loops/clip-to-post/mentions.default.json` now include a `news_feeds` array.

### Frontend

- Page: `frontend/pages/SourcesPage.tsx`
- API helpers: `fetchDictionary` / `saveDictionary` / `parseDictionary` in `frontend/api/articles.ts`
- Nav: `frontend/components/factory/FactoryNav.tsx` gained **Articles** and **Sources** tabs.

---

## Verification

- `go build ./...` + `go vet ./...` clean.
- `npm run build` (tsc + vite) clean.
- Live smoke test: seeded publications + article prompt; ingested a real article; grok analysis produced 6 distinct post candidates; source marked `analyzed`; invalid publication rejected.

## Notes / follow-ups

- The `analyze`/`rewrite` scripts invoke the `grok` CLI (same pattern as the clip factory) with a dev fallback when grok is unavailable.
- No CLI (`cmd/factory`) subcommands were added for articles yet — the API + UI cover the flow.
