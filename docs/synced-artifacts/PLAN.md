# Synced Artifacts / Custom Rooms — Plan (Refined)

**STATUS: STABILIZED LOCAL-FIRST ROOM DEMO IMPLEMENTED** (2026-07-12)

The current implementation is `meta-app/`, a React/Vite application served on
`http://localhost:5200`. It has same-host routes `/`, `/rooms/:roomId`, and
`/join/:inviteId`. Room navigation stays in the application; the old
`about:blank`/`document.write` popup flow is not used.

The only generated bundle today is a fake shared counter. It runs in a sandboxed
iframe through a narrow bridge with a per-load nonce, `event.source` checks,
strict schemas and payload limits. The iframe has no `allow-same-origin`, and
its CSP includes `connect-src 'none'`.

Implemented local-first and transport behavior:
- An IndexedDB vault stores the room `CryptoKey`, encrypted code bundle and
  state, applied operation IDs, outbox, sync cursors, and room/device
  credentials. Local updates commit to this vault before sync is attempted.
- AES-GCM encrypts checkpoints, operation payloads, and signaling envelopes in
  the browser. `roomDataKey` is never sent to the signaling or mailbox API.
- Sharing is split into a public locator URL and a separately shared private
  `roompkg1` package. The package contains the one-time `inviteSecret`, room and
  owner-device identity, and the wrapped room data key. Only `inviteSecret` is
  submitted when redeeming the invite.
- Signaling API v2 provides durable rooms and member seats, owner capability,
  one-time expiring/revocable invites, hashed credentials, idempotent
  redemption, room-scoped device discovery, deterministic session-scoped
  addressed signals, and opaque encrypted mailboxes. Prototype/admin/dev
  routes are disabled.
- Capacity counts unique admitted members. Reconnects, ICE candidates, offers,
  answers, and mailbox traffic do not consume seats.
- A room-wide encrypted checkpoint lets a new member join while the owner is
  offline. Per-device encrypted operation mailboxes provide asynchronous
  delivery. A WebRTC DataChannel provides live P2P delivery when both demo
  peers are online.

The honest product model is **client-canonical, E2EE, and P2P-first, with blind
encrypted store-and-forward fallback**. It is not accurate to claim that no
room data bytes are ever stored on the server: the mailbox retains ciphertext
under TTL and quota limits. The server can also observe room/device identifiers,
membership, timing, and message sizes.

`docs/synced-artifacts/spikes/meta-shell.html` is a historical bridge and UX
spike only. It is not the canonical implementation and its signup/login,
bootstrap-counter, and popup behavior must not be used as evidence for the
current capability-based demo.

Current demo scope is two peers. Multi-peer mesh, TURN, code signing plus
version migration/rollback, real AI/MCP generation, mobile background
operation, production billing/accounts, member removal/key rotation, and
standalone PWA export remain future work.

See also: 
- P2P-RELAY-IDEA.md (exploration)
- PRD-p2p-local-relay-monetization.md (detailed build spec with checklists)

**Locked decisions:**
- Service-mediated admission and rendezvous remain the product gate.
- Clients remain canonical and use live P2P when available, with a blind
  encrypted mailbox for offline/reliability fallback.
- Monetization can cover admission/invite policy, signaling/rendezvous,
  TURN/reliability, and encrypted mailbox quotas. It cannot stop an admitted
  peer from copying room data or implementing another transport.

This long-term architecture direction remains locked:
- Treat code as encrypted room data; compression and version handling still
  require implementation.
- Propagate validated UI updates automatically once code update safeguards
  exist.
- Seamless transparent execution (iframe hidden from user)
- Meta-app centralizes everything; no per-room Git or subdomains

**Core Vision (updated 2026-07-09):**  
A modern take on the original "Rooms" idea — **a meta-app that contains many custom apps ("rooms")**. 

- The meta-app (the "Rooms shell") provides the social layer, vault, contacts, room management, QR/links, and background sync.
- Each "room" inside it can be **arbitrary custom frontend code** (React, Svelte, vanilla, etc.) generated via an AI harness.
- No more schema constraints killing the design. Full visual and interaction freedom per room.
- Every custom room still gets E2EE sync + client-owned data through live P2P
  and a blind encrypted store-and-forward fallback.
- Standalone installation/export for an individual room remains a future path.
- LLM harness is first-class, with a path for users to bring their own AI (MCP/tool calling) to avoid the builder bearing generation costs.

This keeps everything great about Rooms (relay blindness, ownership, easy sharing via safe links + QR, solo-first, small-group collab) while fixing the main shortcoming: lack of custom beautiful UIs.

**Core invariants (non-negotiable):**
- All content over the wire is encrypted. The service is content-blind but can
  observe transport and membership metadata.
- Clients hold the canonical data.
- The service is self-hostable; production packaging and support are future work.
- LLM harness is first-class. Users should be able to bring their own AI (MCP/tool) so the builder doesn't pay for generation.
- Full custom code freedom per room (no schemas that dictate UI).
- Small-to-medium scope per room: forms to 3-5 "page" collab tools (expenses, lists, polls, light boards, etc.).
- Meta-app structure: One shell that hosts many custom-code rooms (inspired by original Rooms "app that contains apps").

## Key Decisions & Refined Direction (2026-07-09)

The pure "standalone single HTML" path is de-emphasized. The stronger model is a **meta-app** (Rooms-like shell) that hosts many custom-code rooms.

### Meta-App Structure ("An app that has custom apps")
- Central React/Vite shell (implemented in `meta-app/`; standalone PWA export is
  not part of this demo).
- The intended product lets users see "My Rooms", create via an AI harness,
  and manage a vault and social layer. The current demo implements local room
  listing/creation without AI generation or contacts.
- The current room loads the fake-counter bundle; arbitrary custom frontends
  remain the target.
- The shell provides consistent social layer (safe links, QR, member profiles, invites) while the room itself is fully custom.
- Analogy to original Rooms: one meta experience, many independent "apps" (now with arbitrary code instead of templates/schemas).

### Custom Code Loading & Security
- The fake counter bundle is loaded into a sandboxed iframe with scripts only;
  it does not receive same-origin authority.
- The parent bridge validates the iframe window via `event.source`, requires a
  fresh per-load nonce, enforces a versioned message schema and payload limits,
  and only accepts `getState`, `update`, and `subscribe`.
- The generated document's CSP denies network connections and all undeclared
  resources. It communicates only through the parent bridge.

### Creation Flow
- In the current demo, the user names a room and chooses its unique-member
  capacity. The client creates the durable server room and capability, generates
  the fake-counter bundle and room data key, encrypts the initial bundle/state,
  stores them locally, creates one invite, and uploads an encrypted room
  checkpoint.
- Real AI/MCP generation is intentionally not implemented yet.

### PWA per Room
- Exporting/installing a room as its own standalone PWA is a post-demo feature.
- Future export must preserve capability admission, key separation, sandboxing,
  and local-first storage; it must not inject durable secrets into public URLs.

### Deployment, Code Storage, Sandboxing & "The JSON Equivalent" (Critical Section)

**The core worry**: Old Rooms was simple because each room was basically structured JSON data stored in the Yjs document (meta + template.* maps). No real "code deployment". Everything was data.

**The equivalent here**: Treat the **entire custom frontend** as just another piece of (encrypted) data inside the room.

**User direction (2026-07-11):** The meta-app should eventually be installable
and offline-capable, with standalone room export as a later option. The current
demo keeps data entry and canonical state local but still uses the centralized
service for admission, rendezvous, encrypted offline bootstrap, and encrypted
mailbox delivery.

#### Storage Model ("Code as Encrypted Data")
- IndexedDB is the local canonical vault. It stores the `CryptoKey`, encrypted
  fake-counter bundle and state, operation IDs, outbox, mailbox/signal cursors,
  member credential, and owner capability where applicable.
- The bundle and state are encrypted separately with AES-GCM. A room-wide,
  AES-GCM-encrypted checkpoint contains those encrypted values plus bootstrap
  metadata, allowing first join while existing clients are offline.
- Counter increments become uniquely identified encrypted operations. The
  local encrypted state, applied-ID set, and outbox entry are persisted before
  synchronization starts.
- There is no Yjs/CRDT document, compression pipeline, code publishing/version
  history, migration, or rollback in the current demo.
- The long-term decision remains to treat generated code as encrypted room
  data, without a Git repository or subdomain per room. Compression and
  versioning must be validated when real generated bundles are introduced.

#### Loading a Custom Room in the Meta-App
1. User opens `/rooms/:roomId`; the shell loads the encrypted room record from
   its IndexedDB vault.
2. The shell decrypts and validates the fake-counter bundle descriptor and
   current counter state.
3. Shell creates a **sandboxed iframe** with `sandbox="allow-scripts"` and
   `srcDoc`; `allow-same-origin` is deliberately absent.
4. Establishes a **postMessage bridge** (strictly validated):
   - Both sides validate the source window and per-load nonce.
   - The iframe receives counter state and can request only `getState`,
     `update({type: "counter.increment"})`, or `subscribe`.
5. The fake-counter bundle contains the small bridge client and updates through
   `postMessage`; an arbitrary mutator API is not implemented.
6. The shell owns persistence, cryptography, signaling/mailbox access, and
   WebRTC. The bridge is invisible to the end user.

The long-term UX decision remains that validated code updates should propagate
without an accept step and the bridge should feel native. Code update
propagation is not implemented by this fake-counter demo.

#### Versioning & "Deployment" (Future)
- Future generated bundles need signed/provenanced versions, schema migration,
  compatibility checks, and rollback. None is implemented in the demo.
- The intended product behavior remains automatic propagation after those
  safeguards exist; there is no current owner publish flow.

#### Sandboxing Details (iframe + postMessage)
- Implemented restrictions are documented in `SECURITY-AND-SANDBOX.md`.
- The opaque-origin iframe cannot access the shell DOM, vault, credentials,
  room key, signaling client, or network. The parent owns persistence, crypto,
  admission, mailbox, and P2P transport.
- Owner provenance alone is not a security boundary. Future arbitrary
  generated code still needs signing/provenance and migration policy.

#### Standalone / Exported PWA Path (Future)
- Export to a standalone PWA or user-hosted site is not in the implemented
  demo. All rooms currently run inside the meta-app.
- This is one future portability path, not an MVP success criterion.

### Sharing and Admission
- The share locator is public:
  `/join/:inviteId`.
- The private material is a separately shared, bounded `roompkg1` package with
  the one-time invite secret, room ID, owner device ID, and room key wrapped by
  an invite-derived AES-GCM key.
- The service receives the invite secret for redemption but never receives the
  room data key. The client validates package/route identity, server room and
  owner identity, checkpoint metadata, and decrypted bundle/state before
  writing the joined room to IndexedDB.
- There are no signup/login accounts in this demo. Room membership uses
  possession-based owner and member capabilities stored as hashes server-side.

### Current State and Sync Model
- The fake counter uses idempotent operation IDs and deterministic application,
  which is enough for concurrent increments in the two-peer demo. It is not a
  general CRDT.
- Local state is canonical. Each update is persisted to IndexedDB before the
  synchronizer sends it.
- If both peers are online, encrypted operations can flow over an ordered
  WebRTC DataChannel. The same encrypted operations are also placed in
  recipient-addressed mailboxes for asynchronous convergence.
- The signaling service stores encrypted signaling envelopes and mailbox
  ciphertext durably with TTL and quota. It cannot decrypt them.

## Stabilized Demo Success Criteria
- React/Vite shell on `:5200` supports `/`, `/rooms/:roomId`, and
  `/join/:inviteId` without popup routing.
- A capacity-two room can be created, privately invited, joined from an
  isolated browser vault while the owner is offline, and reopened later.
- A wrong invitation package is rejected without persisting a room.
- Repeating redemption from the same admitted device is idempotent and keeps
  `memberCount=2`; a third unique member is rejected.
- Both browser vaults converge through mailbox fallback and establish a live
  P2P DataChannel when simultaneously online.
- Concurrent counter increments converge.

## Verified Evidence (2026-07-12)
Run the signaling service:

```sh
cd signaling
go run . -addr=:8081 -db=./signaling.db
```

Run the frontend:

```sh
cd meta-app
npm run dev
```

Verification commands:

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

The Go suite and race detector pass. The frontend has 16 passing tests across
seven files; lint and production build pass. The real-browser vertical slice
passes offline first join, wrong invite not persisted, idempotent reconnect
with `memberCount=2`, third unique member rejection, same-host routing,
mailbox convergence, live P2P connection, and concurrent counter convergence.
`npm run test:e2e` requires Python Playwright and system Chrome, with servers
running on ports 8081 and 5200.

## Later Work
- Multi-peer mesh beyond the current two-peer selection.
- TURN and production-grade connectivity/reliability.
- Signed bundle provenance, versioning, state migration, compatibility, and
  rollback.
- Real AI/MCP generation and iterative room refinement.
- Mobile background operation.
- Production accounts, billing, quotas, and operational controls.
- Member removal plus room-key rotation.
- Standalone PWA export and self-hosted room packaging.
- Broader generated-app merge semantics beyond the counter operation model.

The historical `spikes/meta-shell.html` remains useful as a design reference,
but `meta-app/` and `signaling/` are the implementation evidence as of
2026-07-12.
