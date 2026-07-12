# Security and Sandboxing for Custom Code Rooms

**Implementation baseline:** 2026-07-12

## Current Security Boundary

The only generated bundle is the fake counter. It runs in an iframe with:

```html
sandbox="allow-scripts"
```

`allow-same-origin` is intentionally absent, so the frame has an opaque origin.
The frame cannot directly access the shell DOM, IndexedDB vault, local
credentials, `roomDataKey`, signaling client, mailbox client, or WebRTC
connection.

The generated document's CSP denies all resources by default, permits only
nonce-authorized inline script and style, and includes:

```text
connect-src 'none'
img-src 'none'
font-src 'none'
object-src 'none'
base-uri 'none'
form-action 'none'
```

The generated code therefore has no direct network channel in the implemented
demo.

## Bridge Validation

The parent accepts a bridge message only when all checks pass:

- `event.source` is the currently mounted iframe window.
- Channel, protocol version, and message type match.
- A fresh per-load nonce matches.
- Request ID is a bounded string.
- The serialized message is no larger than 8 KiB.
- The method is exactly `getState`, `update`, or `subscribe`.
- `update` contains exactly `{type: "counter.increment"}`; other methods carry
  no payload.

The iframe applies corresponding checks to parent messages, including
`event.source === parent` and the same nonce. A new frame receives a new nonce,
and old subscriptions are disconnected.

`postMessage(..., "*")` is used because a sandboxed frame without
`allow-same-origin` has an opaque origin. Security comes from source-window,
nonce, protocol, schema, and size validation rather than an origin string.

## Local Vault and Persistence

IndexedDB is the canonical local store. It contains:

- the room AES-GCM `CryptoKey`;
- encrypted bundle and state;
- applied operation IDs, outbox, and sync cursors;
- room/member/device identifiers and private device identity;
- member credential and owner capability where applicable;
- retained invitation data for the owner.

Room content is encrypted at rest in the vault. Credentials and metadata are
not separately encrypted from the application origin.

Local counter updates persist the new encrypted state, operation ID, and
encrypted outbox payload before synchronization starts. Remote operations are
schema-validated, deduplicated, and persisted before notification.

## End-to-End Encryption

AES-GCM is performed client-side for:

- room bundle and state;
- room-wide bootstrap checkpoints;
- operation payloads;
- offer, answer, and ICE signaling envelopes;
- wrapping `roomDataKey` for invitations.

Each use has random IV material and context-specific authenticated additional
data. `roomDataKey` never reaches the signaling or mailbox API.

The server stores opaque ciphertext but can observe metadata such as room and
device IDs, durable membership, timing, endpoints used, and message sizes.
E2EE protects content, not this metadata.

## Invitation and Capability Security

Sharing separates:

- a public locator URL, `/join/:inviteId`; and
- a separately shared private `roompkg1` package containing `inviteSecret`,
  room/owner-device identity, and the wrapped room data key.

The service receives only `inviteSecret` from the private package during
redemption; it does not receive the wrapped or raw room data key.

The client validates package size/schema and route identity, server room and
owner identity, checkpoint metadata, and decrypted bundle/state before writing
the room to IndexedDB. A wrong or mismatched invite is not persisted.

Server-side authorization is capability-based:

- owner capability controls invite issuance and revocation;
- member bearer credentials authorize room, device, signaling, and mailbox
  access;
- owner capability, invite secret, member credential, device identity, and
  idempotency material are stored only as hashes;
- invites are one-time, expiring, and revocable;
- redemption is transactional and idempotent.

There is no signup/login account model in the current demo.

## Signaling and Mailbox Security

Signaling is room-scoped, session-scoped, recipient-addressed, expiring, and
size/quota limited. Payload content is encrypted before upload. Device
discovery returns only admitted device IDs with owner/self flags.

The mailbox stores:

- one room-wide opaque checkpoint for offline first join; and
- recipient-addressed opaque operations for asynchronous convergence.

Mailbox ciphertext has TTL, payload-size, and record-count quotas. The server
is blind to content, but storing ciphertext means the product must not claim
that no room data bytes reside on the server.

Prototype `/admin`, `/dev/*`, `/register`, `/get`, and `/room/*` endpoints are
not registered.

## Capacity and Trust Limits

Capacity counts unique durable admitted members. Reconnects, signaling events,
ICE traffic, mailbox records, and P2P sessions do not consume seats.

Admission controls who receives capabilities and encrypted room material. It
cannot prevent an admitted peer from:

- reading or copying plaintext shown on that device;
- extracting data or keys from its own client environment;
- modifying the open client;
- sharing data out of band; or
- implementing another transport.

This is a collaboration boundary, not DRM.

## Current Limitations

- The demo targets two peers; there is no multi-peer mesh.
- There is no TURN fallback.
- Generated code is only the fake counter.
- There is no code signing/provenance, migration, or rollback.
- There is no member removal or room-key rotation.
- Mobile background behavior is not implemented.
- Production accounts, billing, abuse controls, and operational hardening are
  not implemented.
- Standalone PWA export is future work.

## Verification

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

As of 2026-07-12, Go tests/race, 16 frontend tests across seven files,
frontend lint/build, and the real-browser vertical slice pass. The browser test
requires Python Playwright and system Chrome, with signaling on 8081 and the
frontend on 5200.

`docs/synced-artifacts/spikes/meta-shell.html` is historical reference only.
Its looser prototype behavior is not the current security model.
