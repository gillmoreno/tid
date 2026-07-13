# Roomworks Docker Public Pilot

- **Date:** 2026-07-12
- **Hosts:** `rooms.the-idea-guy.com`, `rooms-api.the-idea-guy.com`
- **Runtime:** Docker Compose on Gil's Mac (nginx 1.30.3, Go 1.25.12) + one named Cloudflare Tunnel
- **Status:** Live since 2026-07-13; signed creator permits deployed; edge rate-limit rules and room-creating production tests remain pending

The operator selected the same infrastructure shape as Roger: a unified Docker image containing
the Vite frontend, Go service, nginx, and supervisor. This supersedes the earlier Cloudflare Pages
plan. The existing TID container, apex deployment, and Post Factory remain separate and unchanged.

This remains a pilot, not an always-on or TURN-backed deployment. The Mac and Docker Desktop must
stay powered on, awake, online, and running. Cloudflare Tunnel carries HTTP(S) only. Browsers still
attempt direct STUN-assisted WebRTC and use the encrypted mailbox when direct P2P fails.

## Topology

```text
rooms.the-idea-guy.com -----\
                              named tunnel -> 127.0.0.1:4510 -> roomworks container nginx
rooms-api.the-idea-guy.com -/                                  |-- static Vite app + /api/*
                                                               `-- Go API 127.0.0.1:8081
                                                                        |
                                                                        `-- /data/roomworks-signaling.db
```

Docker publishes only `127.0.0.1:4510`; it does not bind a LAN address and needs no router port
forward. The Go process binds only to the container's loopback because nginx and Go share the same
container. The data directory is a host bind mount outside the repository.

## Security Changes

- `POST /v2/rooms` requires a signed `X-Room-Creator-Permit`. Each Ed25519 permit is bound to an
  exact unique-member capacity, expires after at most seven days, and is consumed once in the same
  SQLite transaction that creates the room. Only a SHA-256 hash of its public token ID is stored.
- The private Ed25519 seed exists only on the Mac. The web-exposed Go process receives the public
  verification key, so it cannot mint permits. There is no minting HTTP endpoint and the retired
  deployment-wide pilot secret is no longer present in Docker.
- The browser stores the pasted permit only in the current tab's `sessionStorage`, sends it only
  during room creation, and clears it after success or a permit rejection. It decodes capacity only
  as an untrusted UI hint; Go verifies the signature and exact capacity. The permit never enters a
  Vite variable, room record, invitation package, URL, or checked-in file.
- API CORS permits only `https://rooms.the-idea-guy.com` and explicit local development. CORS is
  not used as authorization.
- Cloudflare zone HTTPS settings were unavailable through the current API credentials, so nginx
  enforces the policy at the localhost-only tunnel ingress. It trusts the tunnel's
  `X-Forwarded-Proto`, redirects requests forwarded as HTTP to the same HTTPS host and URI with
  `308`, and sends HSTS only for HTTPS requests in the two configured host blocks.
- Go's normal request logs contain request ID, method, route template, status, and duration, not
  resource IDs. nginx access logs are disabled, but its warn-level error logs can include the full
  request URI on upstream failures and therefore may contain room, invite, or session IDs; protect
  them like pilot data. Neither service logs creator permits, credentials, request headers, or
  request bodies.
- nginx sets CSP, framing, MIME, referrer, and permissions policies on the parent app. The
  self-contained sandbox frame at `/room-frame.html?v=3` uses CSP-hashed inline JavaScript and CSS,
  retains `connect-src 'none'`, and has no external asset dependency.
- The production browser uses same-origin `/api`, matching Roger's unified-container pattern. nginx
  strips that prefix before proxying to Go. The service worker explicitly bypasses `/api/*`, so
  authenticated API responses never enter the static cache. `rooms-api.*` remains a direct API
  alias for operator checks and future separation, not a browser availability dependency.
- The active worker is registered at the versioned `/sw-v2.js` path, uses the
  `roomworks-shell-v5` cache, and is served with `no-store`. Cloudflare reports that path as a
  cache bypass. The versioned URL also avoids an older edge-cached `/sw.js`; the current API token
  returned `401` for an explicit cache purge.

## Docker Files

| File | Purpose |
|---|---|
| `Dockerfile.roomworks` | Node build, Go build, nginx/supervisor runtime |
| `docker-compose.roomworks.yml` | Localhost publication, external data mount, health, restart, log rotation |
| `deploy/roomworks/nginx.conf` | Host routing, SPA fallback, static headers, API proxy |
| `deploy/roomworks/supervisord.conf` | Restarts nginx and Go inside the container |
| `deploy/roomworks/signaling.env.example` | Placeholder-only Compose environment |
| `deploy/roomworks/cloudflared.yml.example` | Both hostnames routed to `127.0.0.1:4510` |

## Fixed Host Paths

| Purpose | Path | Mode |
|---|---|---|
| Compose environment | `/Users/ai-gil/.config/roomworks/roomworks.env` | `0600` |
| Creator signing seed | `/Users/ai-gil/.config/roomworks/creator-signing.seed` | `0600` |
| SQLite database | `/Users/ai-gil/Library/Application Support/Roomworks/data/roomworks-signaling.db` | `0600` |
| Backups | `/Users/ai-gil/Library/Application Support/Roomworks/backups/` | `0700` directory, `0600` files |
| Tunnel config | `/Users/ai-gil/.cloudflared/roomworks-signaling.yml` | `0600` |
| Tunnel credential | `/Users/ai-gil/.cloudflared/<tunnel-uuid>.json` | `0600` |
| Tunnel logs | `/Users/ai-gil/Library/Logs/Roomworks/` | `0700` directory |

No database, backup, private signing seed, permit, tunnel credential, or runtime log belongs in
this repository.

## Docker Deployment

Create the directories and initialize the external signing seed locally:

```sh
install -d -m 700 \
  "$HOME/.config/roomworks" \
  "$HOME/Library/Application Support/Roomworks/data" \
  "$HOME/Library/Application Support/Roomworks/backups" \
  "$HOME/Library/Logs/Roomworks"
cd signaling
go run ./cmd/roomworks-token init \
  --seed-file "$HOME/.config/roomworks/creator-signing.seed"
```

The `init` command refuses to replace an existing seed and prints only the public verification
key. Keep the private seed at `0600`; do not put it in Docker, chat, shell arguments, or Git.

`~/.config/roomworks/roomworks.env` must contain:

```sh
ROOMWORKS_DATA_DIR=/Users/ai-gil/Library/Application Support/Roomworks/data
SIGNALING_ALLOWED_ORIGINS=https://rooms.the-idea-guy.com,http://localhost:5200
ROOMWORKS_CREATOR_VERIFY_KEY=<base64url public key printed by init>
```

Set the file to `0600`, then deploy:

```sh
just roomworks-deploy "$HOME/.config/roomworks/roomworks.env"
just roomworks-status "$HOME/.config/roomworks/roomworks.env"
curl -fsS -H 'Host: rooms-api.the-idea-guy.com' http://127.0.0.1:4510/healthz
```

The Compose service uses `restart: unless-stopped`, and supervisor separately restarts nginx or Go
after an internal process failure. Docker Desktop must be configured to start at login for recovery
after a Mac restart. Inspect and operate it with:

```sh
just roomworks-logs "$HOME/.config/roomworks/roomworks.env"
docker inspect --format '{{json .State.Health}}' roomworks
docker compose --env-file "$HOME/.config/roomworks/roomworks.env" \
  -f docker-compose.roomworks.yml restart roomworks
just roomworks-down "$HOME/.config/roomworks/roomworks.env"
```

`down` removes the container/network only. It does not remove the external SQLite directory.

## Mint Room Creator Permits

The project-local skill is `.agents/skills/roomworks-mint-creator-token`. Natural requests such as
"create me a Roomworks token for 11 people" run its offline script, which copies a one-use permit
to the macOS clipboard without calling the API or printing the token. Capacity includes the owner.
The default lifetime is 24 hours; accepted capacity is 2 through 50.

Direct operator equivalent:

```sh
.agents/skills/roomworks-mint-creator-token/scripts/mint-token.sh \
  --capacity 11 --ttl 24h
```

The permit is at-most-once. Once `POST /v2/rooms` commits, the permit is consumed even if its `201`
response is lost or later client setup fails before owner credentials are durably stored. That
failure window includes WebCrypto setup, initial invitation or checkpoint creation, and IndexedDB
persistence. The permit then reports used and may leave an empty or orphaned room; mint another
permit. Retry-safe recovery would require a separate client-generated credential/idempotency
design.

## Named Tunnel

A dedicated `roomworks` tunnel now serves both public hostnames. To recreate its DNS routes:

```sh
cloudflared tunnel login
cloudflared tunnel create roomworks
cloudflared tunnel route dns roomworks rooms.the-idea-guy.com
cloudflared tunnel route dns roomworks rooms-api.the-idea-guy.com
```

Copy `deploy/roomworks/cloudflared.yml.example` to
`~/.cloudflared/roomworks-signaling.yml`, replace the UUID and credential path locally, set it to
`0600`, and validate both ingress rules:

```sh
cloudflared tunnel --config "$HOME/.cloudflared/roomworks-signaling.yml" ingress validate
cloudflared tunnel --config "$HOME/.cloudflared/roomworks-signaling.yml" ingress rule \
  https://rooms.the-idea-guy.com/
cloudflared tunnel --config "$HOME/.cloudflared/roomworks-signaling.yml" ingress rule \
  https://rooms-api.the-idea-guy.com/healthz
```

Install `deploy/roomworks/com.the-idea-guy.roomworks-tunnel.plist.example` in
`~/Library/LaunchAgents` after replacing its absolute paths. This is a dedicated launch agent so it
cannot collide with other `cloudflared` services:

```sh
plutil -lint "$HOME/Library/LaunchAgents/com.the-idea-guy.roomworks-tunnel.plist"
launchctl bootstrap "gui/$(id -u)" \
  "$HOME/Library/LaunchAgents/com.the-idea-guy.roomworks-tunnel.plist"
launchctl kickstart -k "gui/$(id -u)/com.the-idea-guy.roomworks-tunnel"
cloudflared tunnel info roomworks
```

Cloudflare's current documentation covers [locally managed named tunnels](https://developers.cloudflare.com/tunnel/advanced/local-management/create-local-tunnel/)
and [macOS service supervision](https://developers.cloudflare.com/tunnel/advanced/local-management/as-a-service/macos/).

## Legacy Deployment Retirement

The prior Rooms Docker container in the adjacent `the-idea-guy` repository was removed on
2026-07-13. Its external data volume was retained rather than deleted. The old `rooms.*` and
`relay.*` ingress entries were removed from the shared `choreboard-relay` tunnel configuration,
and the obsolete `relay.the-idea-guy.com` DNS record was deleted.

The shared tunnel itself was not deleted because it still serves `inkanto.the-idea-guy.com`.
Its launch agent was restarted after the scoped configuration change, and Inkanto was verified
publicly. The TID, apex, and Roger containers were not changed.

## DNS, TLS, And CORS

Both public DNS records must target the named tunnel. There is no Pages project for Roomworks.
Cloudflare terminates public TLS; nginx applies the HTTP redirect and HSTS policy from the trusted
tunnel `X-Forwarded-Proto`. Verify:

```sh
curl -fsSI https://rooms.the-idea-guy.com/
curl -fsS https://rooms-api.the-idea-guy.com/healthz
curl -fsSI http://rooms.the-idea-guy.com/
curl -fsSI http://rooms-api.the-idea-guy.com/healthz
curl -i -X OPTIONS https://rooms-api.the-idea-guy.com/v2/rooms \
  -H 'Origin: https://rooms.the-idea-guy.com' \
  -H 'Access-Control-Request-Method: POST' \
  -H 'Access-Control-Request-Headers: content-type,x-room-creator-permit'
curl -i https://rooms-api.the-idea-guy.com/healthz \
  -H 'Origin: https://unrelated.example'
```

Both HTTP requests must return `308` to the same HTTPS host and URI. HTTPS responses must carry
`Strict-Transport-Security`. Production preflight must return `204` and echo only the Roomworks
origin. The unrelated origin must return `403`. Direct requests to `/join/:inviteId` and
`/rooms/:roomId` must return the SPA.

## Cloudflare Rate Limits

**Not active as of 2026-07-13.** The available Cloudflare API token can manage DNS but returned
`403 Authentication error` when reading the zone's `http_ratelimit` entrypoint ruleset. Complete
this section with a token that has zone read plus the applicable zone WAF/rulesets edit permission
before widening the pilot. The application-level signed creator-permit gate is active independently.

Create zone-level rate limiting rules scoped to `rooms-api.the-idea-guy.com`, characterized by
`ip.src` and `cf.colo.id`. Use Block with HTTP `429`; do not use request-body fields or enable
payload logging.

Both host/path shapes must be covered: browsers use `rooms.the-idea-guy.com/api/v2/*`, while the
operator alias uses `rooms-api.the-idea-guy.com/v2/*`. Scoping only the alias leaves the primary
same-origin API unprotected.

| Rule | Expression | Pilot threshold | Mitigation |
|---|---|---:|---:|
| Room creation | `http.request.method eq "POST" and ((http.host eq "rooms.the-idea-guy.com" and http.request.uri.path eq "/api/v2/rooms") or (http.host eq "rooms-api.the-idea-guy.com" and http.request.uri.path eq "/v2/rooms"))` | 5 / 10 min | 1 hour |
| Invite redemption | `http.request.method eq "POST" and ((http.host eq "rooms.the-idea-guy.com" and starts_with(http.request.uri.path, "/api/v2/invites/") and ends_with(http.request.uri.path, "/redeem")) or (http.host eq "rooms-api.the-idea-guy.com" and starts_with(http.request.uri.path, "/v2/invites/") and ends_with(http.request.uri.path, "/redeem")))` | 20 / min | 10 min |
| Signaling/mailbox | `(http.host eq "rooms.the-idea-guy.com" and starts_with(http.request.uri.path, "/api/v2/rooms/")) or (http.host eq "rooms-api.the-idea-guy.com" and starts_with(http.request.uri.path, "/v2/rooms/"))` | 600 / min | 5 min |

Cloudflare deploys these in the `http_ratelimit` phase; see its [Rulesets API guide](https://developers.cloudflare.com/waf/rate-limiting-rules/create-api/)
and [parameter reference](https://developers.cloudflare.com/waf/rate-limiting-rules/parameters/).
Confirm plan-specific periods in the dashboard. Start briefly in Log mode if available, then enable
Block before sharing the pilot.

## Backup And Restore

Run a daily online backup from the bind-mounted host database with 14-day retention:

```sh
just backup-roomworks \
  "/Users/ai-gil/Library/Application Support/Roomworks/data/roomworks-signaling.db" \
  "/Users/ai-gil/Library/Application Support/Roomworks/backups" 14
```

The script uses SQLite's online `.backup`, verifies `PRAGMA quick_check`, and prints only the backup
path. A production online backup and an isolated restore/`PRAGMA quick_check` exercise passed on
2026-07-13. The live production database was not replaced during that exercise.

Restore procedure:

1. Run `just roomworks-down "$HOME/.config/roomworks/roomworks.env"`.
2. Copy the current database to a dated quarantine path and keep it at `0600`.
3. Run `sqlite3 "$DB_PATH" ".restore '$BACKUP_PATH'"`.
4. Require `sqlite3 "$DB_PATH" 'PRAGMA quick_check;'` to return `ok` and set the DB to `0600`.
5. Redeploy the Compose service and verify local and public health plus existing member counts.
6. Preserve the quarantine database until the restored pilot is verified.

Restart tests must prove that container replacement preserves rooms and admitted members. Tunnel
restart must restore both stable public URLs.

## Production Test Record

Do not create production rooms until explicitly approved.

| Check | Result |
|---|---|
| Container healthy; host port bound only to `127.0.0.1:4510` | Pass, 2026-07-13 |
| Runtime image versions | nginx 1.30.3 and Go 1.25.12, pass, 2026-07-13 |
| Public frontend/API health and production CORS | Pass, 2026-07-13 |
| Tunnel-aware HTTP redirect and host-scoped HSTS on both hosts | `308` preserves HTTPS host/URI; HSTS present on HTTPS, pass, 2026-07-13 |
| Desktop/mobile browser render, console, and horizontal overflow | Pass, 2026-07-13 |
| Direct `/join/:inviteId` and `/rooms/:roomId` SPA navigation | Pass, 2026-07-13 |
| Self-contained `/room-frame.html?v=3`: hashed inline JS/CSS, compatible framing CSP, and offline opaque/storage-blocked execution | Pass, 2026-07-13 |
| Missing/invalid creator permit rejected without creating a room | Pass, 2026-07-13 |
| Offline signer/public verifier match; 11-person permit prefills read-only capacity | Pass without creating a room, 2026-07-13 |
| Isolated Docker signed-permit flow: offline join, idempotent reconnect, durable capacity, same-origin routing, and convergent counter sync | Pass, 2026-07-13 |
| Offline-owner admission/checkpoint bootstrap | Not run; creates production records |
| Cross-network encrypted mailbox convergence | Not run; creates production records |
| Direct WebRTC: home Wi-Fi to cellular | Not run; creates production records |
| Direct WebRTC: second home/office network | Not run; creates production records |
| P2P failure with mailbox fallback, if observed | Not run; creates production records |
| Wi-Fi/cellular switch recovery | Not run; creates production records |
| Idempotent reopen and two-member capacity | Not run; creates production records |
| Container/tunnel restart recovery | Pass for health and public routing, 2026-07-13; record persistence not exercised |
| Retired `/admin`, `/dev/*`, `/register`, `/get`, `/room/*` routes return `404` on both public hosts | Pass, 2026-07-13 |
| Production online backup and isolated restore validation | Pass, 2026-07-13 |

Classify admission/mailbox and direct P2P separately. A STUN-only P2P failure is not a tunnel
failure. If direct P2P commonly fails, authenticated TURN is the next reliability milestone.

## Rollback

1. Boot out `com.the-idea-guy.roomworks-tunnel` to remove public reachability immediately.
2. Run Compose `down` for `docker-compose.roomworks.yml`; do not use `-v` and do not delete data.
3. Remove only the two Roomworks tunnel DNS routes. Do not touch TID/apex or another tunnel.
4. Preserve the SQLite database, last known-good backup, external env, credential, and logs.
5. Redeploy the last known-good Roomworks image/config if rollback is due to an application change.
6. Verify the existing TID and Roger containers were not changed.

The first rollback priority is pilot data and credential protection, not uptime.
