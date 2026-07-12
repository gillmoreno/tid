# Synced Artifacts — Implemented Client Kit and Bridge

**Status:** Stabilized fake-counter implementation
**Implementation baseline:** 2026-07-12

## Purpose and Scope

The current kit proves one local-first room type end to end. It does not yet
provide the previously proposed generic `initSyncedArtifact()` API, arbitrary
mutator functions, Yjs/CRDT compatibility, standalone operation, AI-generated
bundles, code publishing, or code migration.

The only generated bundle is a fake shared counter. The parent React/Vite shell
owns storage, cryptography, admission, mailbox synchronization, and WebRTC.
Generated code runs inside a sandboxed iframe and can interact only through the
counter bridge.

## Runtime Layout

```text
fake-counter iframe (opaque origin)
        ↕ postMessage: getState / increment / subscribe
React/Vite shell
        ├─ IndexedDB vault (canonical local room)
        ├─ AES-GCM crypto
        ├─ outbox, operation IDs, mailbox/signal cursors
        ├─ signaling API v2 and encrypted mailbox
        └─ WebRTC DataChannel when both peers are online
```

The application runs at `http://localhost:5200` with routes `/`,
`/rooms/:roomId`, and `/join/:inviteId`. Room opening does not create an
`about:blank`/`document.write` popup.

## Generated Bundle Contract

The encrypted bundle descriptor currently has one shape:

```ts
type CodeBundle = {
  version: 1
  kind: 'counter'
  title: string
}
```

After decrypting and validating this descriptor, the shell generates the
self-contained fake-counter HTML. A fresh nonce is created on each render and
used by both its CSP and bridge protocol.

The iframe is rendered with:

```html
<iframe sandbox="allow-scripts" srcdoc="...">
```

There is no `allow-same-origin`. The generated document has an opaque origin
and a CSP that includes:

```text
default-src 'none'
script-src 'nonce-…'
style-src 'nonce-…'
connect-src 'none'
img-src 'none'
font-src 'none'
object-src 'none'
base-uri 'none'
form-action 'none'
```

## Bridge Protocol

Requests are versioned messages on channel `meta-room`. Every request must:

- originate from the exact iframe `Window` (`event.source` validation);
- carry the current per-load nonce;
- use protocol version 1 and type `bridge.request`;
- have a bounded request ID;
- stay within the 8 KiB parent-side message limit;
- match the method-specific schema exactly.

Implemented methods:

```ts
type RoomBridge = {
  getState(): Promise<{ value: number }>
  update(operation: { type: 'counter.increment' }): Promise<{ value: number }>
  subscribe(listener: (state: { value: number }) => void): () => void
}
```

The iframe also checks `event.source === parent`, channel, version, and nonce
before accepting responses or state events. The parent disconnects old
subscriptions when the frame changes.

This is intentionally not a generic arbitrary-object mutation API. That API
remains future work.

## IndexedDB Vault

The local vault database is `meta-room-vault`. It stores:

- stable device ID, private device identity, and label;
- room ID, title, capacity, role, member ID, and device ID;
- member credential and, for owners, owner capability;
- owner device ID;
- the AES-GCM `roomDataKey` as a `CryptoKey`;
- encrypted bundle and encrypted room state;
- applied operation IDs;
- durable local outbox items;
- operation mailbox cursor and per-session signaling cursors;
- invitation locator/package data retained by the owner.

Room writes are serialized per room. `CryptoKey` storage relies on IndexedDB
structured cloning.

There is no localStorage fallback in the implemented demo. Local room bundle
and state are encrypted at rest in IndexedDB; credentials and metadata remain
available to the application origin.

## Persist-Before-Sync Update Path

For every counter increment:

1. Generate a unique operation ID.
2. Decrypt the current local counter state.
3. Ignore the operation if its ID was already applied.
4. Apply `delta: 1`.
5. AES-GCM encrypt the new state and operation payload.
6. In one serialized local flow, store the encrypted state, applied operation
   ID, and outbox entry in IndexedDB.
7. Notify local subscribers.
8. Wake synchronization only after persistence completes.

Remote operations are decrypted, schema-validated, deduplicated by operation
ID, applied, and persisted before subscribers are notified.

The fake counter converges under concurrent increments because each valid,
unique operation is additive. This does not establish general conflict
resolution for arbitrary room state.

## Cryptography and Invitation Keys

- `roomDataKey` is a generated 256-bit AES-GCM `CryptoKey`.
- Bundle, state, operations, checkpoints, and signaling payloads use AES-GCM
  with random 96-bit IVs and context-specific authenticated additional data.
- The private invitation package uses prefix `roompkg1.`.
- Its `inviteSecret` derives the AES-GCM key that wraps `roomDataKey`.
- The raw room data key never reaches signaling or mailbox endpoints.

Sharing uses two separate values:

- Public locator: `/join/:inviteId`.
- Private `roompkg1` package: invite secret, room identity, owner device
  identity, and wrapped room key.

The joiner checks package size and schema, requires the package invite ID to
match the route, redeems the one-time invite, validates the server room and
owner, decrypts and validates the room-wide checkpoint, and only then writes
the room to IndexedDB. A wrong invite is not persisted.

## Checkpoint and Mailbox Sync

The client encrypts a room checkpoint containing:

- title and capacity;
- encrypted bundle;
- encrypted state;
- applied operation IDs;
- checkpoint creation time.

The signaling service stores this as one room-wide opaque envelope, allowing a
new admitted member to bootstrap while the owner and all existing members are
offline.

Each local outbox operation is posted as an opaque encrypted envelope to every
other admitted device's mailbox. Recipients poll with a durable operation
cursor, apply unseen operations, and update the room-wide checkpoint. The
service enforces TTL, payload, and record quotas.

## Live P2P

For the current two-peer demo:

- Room-scoped device discovery selects the other admitted device.
- A deterministic session ID is derived from room ID and the sorted device IDs.
- The lexically earlier device creates the ordered DataChannel.
- Offer, answer, and ICE payloads are AES-GCM encrypted client-side before
  being sent as addressed signaling envelopes.
- Once connected, the DataChannel carries encrypted operation payloads.
- Mailbox delivery remains the asynchronous fallback and does not consume room
  capacity.

There is no multi-peer mesh or TURN support yet.

## Signaling Capability Model

There are no signup/login accounts in this implementation.

- The owner capability administers invitations.
- Member bearer credentials authenticate room, device, signaling, and mailbox
  calls.
- The server stores only hashes of capabilities, invite secrets, device
  identities, credentials, and idempotency keys.
- Invite redemption is transactional and idempotent.
- Capacity counts unique durable admitted members only.

## Run and Verification

```sh
cd signaling
go run . -addr=:8081 -db=./signaling.db
```

```sh
cd meta-app
npm run dev
```

```sh
cd signaling
go test ./... -count=1
go test -race ./...

cd ../meta-app
npm test
npm run lint
npm run build
npm run test:e2e
```

Verified on 2026-07-12: Go tests and race detector pass; 16 frontend tests
across seven files pass; lint and build pass; and the real-browser vertical
slice passes offline first join, failed-invite non-persistence, idempotent
reconnect with two members, third-member rejection, same-host routing, mailbox
convergence, live P2P, and concurrent counter convergence.

The E2E command requires Python Playwright and system Chrome, with servers on
ports 8081 and 5200.

## Future Kit Work

- Generic generated-room API and real AI/MCP generation.
- Multi-peer mesh and TURN.
- Bundle signing/provenance, version negotiation, migration, and rollback.
- General CRDT/merge behavior.
- Member removal and room-key rotation.
- Mobile background synchronization.
- Production accounts, billing, and policy enforcement.
- Standalone PWA export.

`docs/synced-artifacts/spikes/meta-shell.html` is historical/reference material,
not the canonical implementation. `meta-app/src/rooms/` is the current client
implementation.
