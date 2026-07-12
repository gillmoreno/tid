# PRD: Custom Rooms — Local-First, P2P-First, Encrypted Fallback

**Status:** Stabilized two-peer demo implemented
**Implementation baseline:** 2026-07-12

## 1. Product Definition

Custom Rooms is a React/Vite meta-app that hosts small custom applications while
keeping room content client-canonical and end-to-end encrypted. The current
demo runs at `http://localhost:5200` and provides:

- `/` for local rooms and room creation.
- `/rooms/:roomId` for the room experience.
- `/join/:inviteId` for invitation redemption.

Navigation is same-host React routing. There is no `about:blank` or
`document.write` popup flow.

The implemented generated application is deliberately narrow: a fake shared
counter. It proves local persistence, capability admission, encrypted offline
bootstrap, asynchronous convergence, and live WebRTC delivery. It does not
claim arbitrary AI-generated applications yet.

## 2. Product and Privacy Model

The accurate model is **client-canonical, E2EE, and P2P-first, with blind
encrypted store-and-forward fallback**.

- Room bundle and state are canonical in each admitted client's IndexedDB
  vault.
- Local updates persist before network synchronization is attempted.
- When both demo peers are online, a WebRTC DataChannel carries encrypted
  operations directly.
- A room-wide opaque checkpoint supports first join while the owner is offline.
- Recipient-addressed opaque operation mailboxes provide asynchronous fallback.
- The service stores ciphertext under TTL and quota limits. Therefore, “no data
  bytes on the server” is false even though the service cannot read content.
- The server can observe metadata including room and device identifiers,
  durable membership, request timing, message sizes, and mailbox activity.

## 3. Implemented Architecture

### React/Vite meta-app

- Runs on port 5200.
- Uses an IndexedDB vault for `CryptoKey`, encrypted bundle/state, applied
  operation IDs, outbox, mailbox and signaling cursors, device identity, member
  credential, and owner capability where applicable.
- Generates only the fake-counter bundle.
- Renders that bundle in an opaque-origin sandboxed iframe.

### Sandboxed room bridge

- Uses `sandbox="allow-scripts"` without `allow-same-origin`.
- Creates a new nonce for every bundle load.
- Validates `event.source`, protocol channel/version, nonce, request ID,
  method-specific payload shape, and total message size.
- Exposes only `getState`, `update({type: "counter.increment"})`, and
  `subscribe`.
- Applies a generated-document CSP with `connect-src 'none'` and no undeclared
  external resources.

### Local cryptography

- AES-GCM encrypts the room bundle, room state, room-wide checkpoints,
  operation payloads, and WebRTC signaling envelopes in the browser.
- Context strings are authenticated as AES-GCM additional data.
- `roomDataKey` is stored locally as a `CryptoKey`; it never reaches the
  signaling or mailbox API.

### Signaling API v2

The SQLite-backed service provides:

- Durable rooms with a separate owner capability.
- Durable member seats and room-scoped device discovery.
- One-time, expiring, revocable invitations.
- Hashed owner capability, invite secret, device identity, member credential,
  and idempotency material.
- Transactional, idempotent invite redemption.
- Deterministic, session-scoped, recipient-addressed signaling.
- A room-wide opaque encrypted checkpoint and per-device opaque operation
  mailboxes.
- Disabled prototype/admin/dev routes.

The service uses capabilities, not signup/login accounts. A reconnect from an
already admitted device is not a new member.

### Capacity

Capacity means the number of unique durable admitted members.

- The owner consumes one seat.
- A successful one-time invitation can create one additional durable member.
- Idempotent reconnects do not consume another seat.
- Offers, answers, ICE candidates, mailbox records, retries, and P2P
  connections do not consume seats.
- The current product demo targets exactly two peers, although the API accepts
  broader room capacity values.

## 4. Invitation Model

Sharing deliberately separates public location from private authority:

1. Public locator URL: `/join/:inviteId`.
2. Separately shared private `roompkg1` package containing:
   - version and `inviteId`;
   - one-time `inviteSecret`;
   - room identity;
   - owner device identity;
   - `roomDataKey` wrapped with an invite-secret-derived AES-GCM key.

The joiner submits only `inviteSecret` from this private package to redeem the
server invite. The raw room data key is never submitted.

Before persisting a joined room, the client validates the package against the
URL, checks server room and owner identity, unwraps the key, decrypts the
checkpoint, and validates bundle/state metadata. Failed or mismatched
invitations are not written to the local vault.

## 5. Sync Semantics

The counter uses uniquely identified `counter.increment` operations.

- An increment is applied to encrypted local state and appended to the durable
  local outbox before sync.
- Applied operation IDs make retry and duplicate delivery idempotent.
- Deterministic addition makes concurrent increments converge in this demo.
- Outbox operations are sent to recipient mailboxes and, when available, the
  live DataChannel.
- Mailbox cursors and deterministic signaling-session cursors persist locally.
- An updated opaque room checkpoint is uploaded after synchronization.

This is not a general CRDT, multi-peer replication protocol, or durable code
versioning system.

## 6. Monetization Boundary

The service can sell capabilities it actually controls:

- Admission and invite policy.
- Signaling and rendezvous limits.
- TURN and higher connection reliability.
- Encrypted checkpoint/mailbox retention, size, and operation quotas.
- Operational support and managed/self-hosted service options.

It cannot prevent an admitted peer from copying plaintext available on that
peer, exporting room data, modifying a client, or building another transport.
Monetization must not be described as controlling content after admission.

Production pricing, billing, accounts, entitlement enforcement, and TURN are
not implemented.

## 7. Demo Acceptance Evidence

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

Run backend verification:

```sh
cd signaling
go test ./... -count=1
go test -race ./...
```

Run frontend verification:

```sh
cd meta-app
npm test
npm run lint
npm run build
npm run test:e2e
```

Verified on 2026-07-12:

- Go tests and race detector pass.
- Frontend has 16 passing tests across seven files.
- Frontend lint and production build pass.
- The real-browser vertical slice passes offline first join, rejection of a
  wrong invite without local persistence, idempotent reconnect with
  `memberCount=2`, rejection of a third unique member, same-host routing,
  mailbox convergence, live P2P connection, and concurrent counter
  convergence.

`npm run test:e2e` requires Python Playwright and system Chrome, with signaling
on port 8081 and the frontend on port 5200.

## 8. Explicitly Future Work

- Multi-peer mesh and general peer selection.
- TURN and production connection reliability.
- Code signing/provenance, version migration, compatibility, and rollback.
- Real AI/MCP generation and arbitrary room bundles.
- Mobile background operation.
- Production billing, accounts, and entitlements.
- Member removal and room-key rotation.
- Standalone PWA export or user-hosted room packaging.
- General CRDT/merge semantics beyond the fake counter.

The historical `docs/synced-artifacts/spikes/meta-shell.html` is reference-only.
Its signup/login, popup, and old bootstrap-counter behavior are not the current
implementation.
