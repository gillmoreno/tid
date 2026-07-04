# The Idea Guy (TID) — Concept

**Domain:** theideaguy.com  
**Brand:** The Idea Guy — a personal lab for testing AI ideas in public.

## What this is

Everyone talks about AI loops, agents, and "10x" workflows. Most of it is hype. TID is where Gil actually runs the experiments — builds the loops, ships the software, and shares what's real vs. what's noise.

The question behind every project: **what can one person actually build with today's AI?**

## Three sections

### 1. AI Guild of Dev (`aigil.dev`)

The developer community lane. A guild for people building with AI — what's worth learning, what's worth shipping, and what's just Twitter noise.

- **Live project:** [AIGil](https://aigil.dev/)
- **On this site:** links out to aigil.dev; section page describes the guild's role in the lab.

### 2. Loops with Taste

Content automation loops — but human-filtered. The hook is the *output* (posts, shorts), not the pipeline recipe. Gil doesn't publish the full loop mechanics for video content; that stays as the content moat.

| Loop | What it does | Status |
|------|-------------|--------|
| **Clip → Post** | Takes a slice of an AI YouTube clip, extracts the insight, rewrites as tweets/X posts — filtered through personal taste | Testing |
| **Grok Imagine Shorts** | Creates short-form video using Grok Imagine — fast visuals + tight edits for reels/shorts | Testing |

### 3. Software Ideas

Software experiments and products — small bets on what AI makes possible for a solo builder.

| Project | What it is | Status |
|---------|-----------|--------|
| **Roger** | Multi-restaurant reservation platform (Go + React) | Live |
| **TID** | This site — meta-project documenting the lab | Building |

## Tech stack (Roger pattern)

```
tid/
├── go-backend/     Go 1.23 + chi — content API
├── frontend/       React 19 + Vite 6 + Tailwind — SPA
├── docker-compose  Nginx + Go in one image for production
└── justfile        `just dev` / `just deploy`
```

**Production:** single Docker image — Nginx serves the Vite build, proxies `/api` to Go on `:8080`.

**Development:** Go API in Docker (`:8016`), Vite dev server on `:5180` with API proxy.

## API (seed content, v0)

| Endpoint | Returns |
|----------|---------|
| `GET /health` | Health check |
| `GET /api/site` | Site metadata |
| `GET /api/sections` | All three sections |
| `GET /api/sections/{slug}` | Section + its items |
| `GET /api/items/{slug}` | Single idea/loop/project |

Content is seeded in Go for now. SQLite + admin CRUD can come later when you want to edit ideas without redeploying.

## What's next (iterate together)

- [ ] Logo and final brand identity
- [ ] Design polish (colors, typography, motion)
- [ ] Real content for each loop (screenshots, example posts)
- [ ] SQLite-backed content store + simple admin
- [ ] Deploy to theideaguy.com (Docker on VPS or Cloudflare)
- [ ] Analytics (privacy-friendly)

## Dev commands

```bash
# Full dev (Docker Go + Vite hot reload)
just dev

# Production build
just deploy
# → http://localhost:8000

# Or run Go locally without Docker
cd go-backend && go run ./cmd/server
cd frontend && VITE_API_URL=/api VITE_API_PROXY_TARGET=http://localhost:8080 npm run dev
```