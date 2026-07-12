# Local-First Room Demo Stabilization

**Date:** 2026-07-12
**Status:** Implemented and verified

## User-Visible Changes

- The room demo is now a React/Vite application on
  `http://localhost:5200`.
- The home, room, and invitation flows use same-host routes: `/`,
  `/rooms/:roomId`, and `/join/:inviteId`.
- Rooms no longer open through an `about:blank`/`document.write` popup.
- A user can create the current fake-counter room, share a public invitation
  locator plus a separate private package, and join from another isolated
  browser vault.
- A first-time join works while the owner is offline by loading an encrypted
  room-wide checkpoint.
- Edits apply locally first, then converge through the encrypted mailbox; when
  both peers are online, the UI reports a live P2P DataChannel.
- Reconnecting an already admitted device is idempotent. A capacity-two room
  remains at two members, and a third unique member is rejected.

## Implemented Architecture

The browser is canonical:

- IndexedDB stores the room `CryptoKey`, encrypted fake-counter bundle and
  state, operation IDs, durable outbox, signaling/mailbox cursors, device
  identity, and room credentials.
- Counter updates persist locally before synchronization starts.
- AES-GCM encrypts room checkpoints, operation payloads, and signaling
  envelopes client-side.
- The fake-counter bundle runs inside a sandboxed, opaque-origin iframe behind
  a narrow `postMessage` bridge.
- The bridge uses a fresh nonce per load, validates `event.source`, enforces
  versioned schemas and payload limits, and exposes only state read,
  counter-increment, and subscription operations.
- The generated document's CSP has `connect-src 'none'`, and the iframe does
  not receive `allow-same-origin`.

The signaling v2 service provides durable rooms, owner capabilities, hashed
credentials, one-time expiring/revocable invites, idempotent redemption,
durable member seats, room-scoped device discovery, and deterministic
session-scoped addressed signaling. Prototype/admin/dev routes are disabled.

For offline reliability, the service stores one room-wide opaque encrypted
checkpoint and per-device opaque encrypted operation mailboxes. WebRTC
DataChannel delivery is the live P2P path when both demo peers are online.

## Invitation and Key Handling

Invitation sharing is intentionally split:

- Public locator URL: `/join/:inviteId`.
- Private `roompkg1` package: `inviteSecret`, room identity, owner device
  identity, and a wrapped `roomDataKey`.

Only the invite secret is sent to the service during redemption.
`roomDataKey` never reaches signaling or mailbox endpoints.

Before persisting a join, the client verifies package/route identity, server
room and owner identity, checkpoint metadata, and decrypted bundle/state. A
wrong invite is rejected without creating a local room.

## Security and Privacy Semantics

The precise product claim is:

> Client-canonical, end-to-end encrypted, and P2P-first, with blind encrypted
> store-and-forward fallback.

It is not accurate to say that no room data bytes are stored server-side. The
mailbox stores ciphertext under TTL, payload-size, and record-count quotas. The
service cannot decrypt content, but it can observe metadata including room and
device identifiers, durable membership, timing, and message sizes.

Capacity is admission policy: it counts unique durable members. Reconnects,
offers, answers, ICE candidates, mailbox records, and P2P sessions do not
consume seats.

The monetizable boundary is admission/invite policy, signaling/rendezvous,
TURN/reliability, and encrypted mailbox quotas. Once admitted, a peer can copy
plaintext visible on its own device or build another transport; this system is
not DRM.

## Run Locally

Start signaling:

```sh
cd signaling
go run . -addr=:8081 -db=./signaling.db
```

Start the frontend:

```sh
cd meta-app
npm run dev
```

The frontend is available at `http://localhost:5200`.

## Verification Evidence

Backend:

```sh
cd signaling
go test ./... -count=1
go test -race ./...
```

Frontend:

```sh
cd meta-app
npm test
npm run lint
npm run build
```

Real-browser vertical slice:

```sh
cd meta-app
npm run test:e2e
```

The E2E test requires Python Playwright and system Chrome, with signaling on
8081 and the frontend on 5200.

Verified results:

- Go tests pass.
- Go race detector passes.
- Frontend: 16 tests across seven files pass.
- Frontend lint and production build pass.
- Browser E2E passes offline first join, wrong invite not persisted,
  idempotent reconnect with `memberCount=2`, third unique member rejection,
  same-host routing, mailbox convergence, live P2P connection, and concurrent
  counter convergence.

## Remaining Limitations

- The demo targets two peers; multi-peer mesh is not implemented.
- TURN is not implemented.
- The fake counter is the only generated bundle.
- Code signing/provenance, version migration, compatibility, and rollback are
  not implemented.
- Real AI/MCP generation is not implemented.
- Mobile background operation is not implemented.
- Production accounts, billing, entitlement enforcement, and operational
  controls are not implemented.
- Member removal and room-key rotation are not implemented.
- Standalone PWA export is future work.
- Insecure prototype rooms from the old `meta-rooms-v2` localStorage format are
  not migrated into the new authenticated IndexedDB vault.

`docs/synced-artifacts/spikes/meta-shell.html` remains historical/reference
material only. It is not the canonical current implementation.
