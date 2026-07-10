# PRD: Custom Rooms — Local P2P Relay with Mandatory Signaling + Monetization

**Status:** Draft for review  
**Date:** 2026-07-10  
**Related:** 
- PLAN.md (committed meta-app + code-as-encrypted-data model)
- P2P-RELAY-IDEA.md (exploration of local relay + P2P)

## 1. Vision & Problem

**Vision:**  
A meta-app ("Custom Rooms") where users create and share beautiful, fully custom small collaborative apps (forms, expense splitters, movie clubs, etc.). Each room runs arbitrary custom frontend code (generated via AI harness or hand-written) while keeping true data ownership: everything (code + data) is end-to-end encrypted on user devices. Sync happens peer-to-peer via local relays running on users' own machines. 

A tiny mandatory signaling service provides the only centralized "on-ramp" for initial connection (NAT traversal + handshake). This service is the primary monetization lever. Pure 100% peer-to-peer (no signaling) is not offered as an option.

**Problem Statement:**  
- AI makes it trivial to create beautiful UIs but not real multi-user synced apps.
- Existing solutions either centralize data (loss of ownership) or require complex self-hosting (poor UX).
- Original Rooms proved the power of a "blind relay + client-owned data" model but was limited by schema constraints.
- Full decentralization is hard to monetize and hard for normal users.
- We need a model that is extremely local/P2P for data + code, easy to use, and has a clear paid path.

**Key Constraints (non-negotiable):**
- No centralized storage of user data or custom code.
- The relay itself runs locally on user devices.
- Initial connection always goes through the controlled signaling service (minimal gating for monetization).
- Custom code runs seamlessly (user never feels the sandbox/iframe).
- Code updates are just more encrypted data in the room state.
- Self-hosting options exist at higher tiers, but the default path requires the service.

## 2. Goals & Success Metrics

**Primary Goals:**
- Deliver the "app of apps" experience with arbitrary beautiful custom frontends.
- Make P2P sync reliable enough for small groups (2–10 people) with intermittent connectivity.
- Create a monetizable product where the signaling service is the clear paid value.
- Preserve extreme data ownership (users own their code + data + run their own relay).

**Success Metrics (MVP):**
- Users can create a custom room (e.g. expenses or RSVP), invite 1–4 others via QR/link, and see live synced changes within seconds when both sides are online.
- Code blob (custom UI) loads and feels native (no visible iframe artifacts).
- 80%+ of connections succeed via the signaling service on typical home/mobile networks.
- Clear upgrade path: free tier with limits → paid tiers.

## 3. Personas & Use Cases

**Primary Persona:** Power user / small group organizer (e.g. co-parent, friend group leader, hobby club organizer). Wants beautiful custom tools without giving data to big tech.

**Key Use Cases:**
- Create expenses splitter with ex-partner.
- Create RSVP / interest form for event with 5–10 friends.
- Create shared movie/book list with family.
- Update the UI of an existing room (owner pushes new code blob; members get it automatically).
- Use on phone + laptop; sync opportunistically.

**Post-MVP / Nice to Have:**
- Export a room as standalone PWA (or let users host their own custom rooms on their websites). This is valuable later but out of MVP scope. The same underlying technology (code blob + local P2P relay + signaling) can power user-created PWAs and self-hosted sites in the future. For MVP everything lives inside the meta-app.

## 4. Architecture Overview

**High-Level Components:**
1. **Meta-App Shell** (single hosted PWA)
   - Room management, vault, contacts, invite generation, code updates.
   - Loads custom rooms.
   - Provides consistent social layer (safe hash links, QR, member attribution).

2. **Custom Room** (per-room state in encrypted storage)
   - `code.bundle` (compressed, versioned, encrypted) — the full custom frontend.
   - Runtime data (responses, expenses, etc.).
   - Stored and synced exactly like any other field.

3. **Local Relay** (small Go binary/service running on user devices)
   - Manages per-room, per-peer outbound queues.
   - Handles P2P connections (WebRTC DataChannels preferred).
   - Performs handshake using existing crypto derivation (room secrets + passphrase).
   - Applies logical clocks / timestamps for ordering.
   - Stores local persistence for queues + cursors.

4. **Signaling Service** (tiny centralized service — the monetization gate)
   - Mandatory for initial connection setup (STUN + WebRTC signaling / candidate exchange).
   - Auth via tokens tied to user account / room / subscription.
   - Does **not** see or store encrypted blobs.
   - Rate-limited per tier.

5. **Sync Layer / Kit**
   - Thin API exposed to custom code (`update()`, `getState()`, `subscribe()`, member info).
   - Abstracts transport (local relay + signaling).
   - Reuses existing rooms framing (MSG_UPDATE, MSG_CHECKPOINT), crypto, and merge logic.
   - Supports code blob as first-class data.

**Data Flow (after bootstrap):**
- Change → local apply + queue per peer → direct P2P encrypted frame → peer receives + merges.
- Code updates are just another update to the `code` field.

**Connectivity Model:**
- Phones: opportunistic (connect when active) or background if feasible.
- Desktop: better background support.
- Offline: queues hold updates with timestamps.
- LAN: nice-to-have mDNS.

## 5. Core Features & Requirements

### 5.1 Custom Rooms & Code as Data
- [ ] Room state contains `code` (version + compressed bundle) + runtime data.
- [ ] Owner can update code blob; members receive automatically.
- [ ] Code bundle can be a self-contained HTML/JS (or structured) that the loader can execute.
- [ ] Bundle is compressed before encryption.
- [ ] Performance target (to validate): reasonable load time for typical small-app bundles (e.g. < 500KB compressed).

### 5.2 Seamless Execution
- [ ] Custom code runs in sandboxed iframe (or equivalent isolation).
- [ ] PostMessage bridge is completely invisible to the end user.
- [ ] Bridge exposes narrow safe API matching the sync kit.
- [ ] Theming, full-screen feel, and interactions feel native.
- [ ] No visible borders, loading spinners from the bridge, or "powered by iframe" artifacts.

### 5.3 Local Relay (P2P)
- [ ] Runs as small service/binary on desktop and (ideally) mobile.
- [ ] Supports multiple independent rooms simultaneously.
- [ ] Per-room, per-peer message queues with persistence.
- [ ] Uses existing rooms relay framing and crypto for payloads.
- [ ] Logical clocks (Lamport or hybrid) + device IDs for ordering and conflict resolution.
- [ ] Automatic reconnection and queue drain on peer availability.
- [ ] Support for opportunistic mobile connections ("connect when they can").

### 5.4 Mandatory Signaling Service
- [ ] Required for all initial connection setup / NAT traversal.
- [ ] Provides STUN + signaling for WebRTC candidate exchange (or equivalent).
- [ ] Token-based auth tied to user / room / subscription.
- [ ] Does not store or forward encrypted data/code.
- [ ] Cheap to operate; rate-limited per tier.
- [ ] Fallback behavior when direct P2P fails (documented in tiers).
- [ ] No free bypass path — pure direct/manual P2P is not exposed.

### 5.5 Handshake & Invites
- [ ] Invites (QR / hash link) always include signaling token or require account.
- [ ] Cryptographic handshake using existing key derivation (roomCode + secrets + optional passphrase).
- [ ] Mutual authentication before exchanging data.
- [ ] LAN discovery (mDNS) as optional enhancement.

### 5.6 Monetization & Tiers
- [ ] Free tier: limited rooms/peers, basic signaling reliability.
- [ ] Paid tiers: higher limits, better reliability, fallback relays, priority.
- [ ] Higher tier: self-hosted signaling configuration + support.
- [ ] Account system to gate signaling tokens.
- [ ] Clear value messaging: "Your data stays P2P on your devices. We make it easy to initially connect."
- [ ] Self-host option for the signaling binary at top tier (with documentation).

### 5.7 Meta-App Shell Features
- [ ] Create room (via AI harness or manual).
- [ ] Manage rooms, members, invites.
- [ ] Update room code (owner flow).
- [ ] (Post-MVP) Export room as standalone PWA or support user-hosted versions. MVP scope: no export. All rooms inside the meta-app.
- [ ] Background / opportunistic sync indicators.
- [ ] Vault + local storage management (reuse patterns from original Rooms).

### 5.8 AI Harness Integration
- [ ] Harness generates code that works with the bridge + local relay.
- [ ] Generated code includes proper use of sync API.
- [ ] Harness can be called via MCP/tool (bring-your-own AI) or optionally hosted.
- [ ] Output is "room-ready" (works embedded in meta-app and as exported PWA).

### 5.9 Conflict Resolution & Timestamps
- [ ] Every update carries logical timestamp + origin.
- [ ] Merge logic handles concurrent edits (op log or CRDT underneath).
- [ ] Audit history preserved where useful (e.g. expense logs).

### 5.10 Export & Portability (Post-MVP)
- Export to standalone PWA or self-hosted site is explicitly out of MVP scope.
- This is a high-value nice-to-have for later phases. The architecture (code as encrypted blob + local relay + mandatory signaling) is designed to support it eventually, so users can create their own PWAs or host rooms on their own domains.
- For MVP: all rooms live inside the meta-app. No export flow.

## 6. Non-Functional Requirements

- **Security**
  - All payloads encrypted end-to-end (reuse existing AES-GCM + Argon2 derivation).
  - Signaling service never sees plaintext.
  - Sandboxed execution with strict postMessage validation.
  - Invite secrets only in URL hash.

- **Performance & Scale**
  - Small groups (target 2–10 peers per room).
  - Code bundle load time acceptable on mobile (validate with real bundles).
  - Queue pruning after successful sync.
  - Background impact on phone battery minimal (or opportunistic).

- **Reliability**
  - Queues survive app restarts.
  - Graceful handling of offline peers.
  - Automatic retry with backoff.

- **Privacy**
  - Signaling service is minimal and auditable.
  - No central storage of room content.
  - Users can self-host signaling at higher tier.

## 7. Technical Foundations (Reuse & Evolution)

- Reuse heavily from `projects/rooms`:
  - Relay framing and protocol.
  - Crypto (deriveKey, deriveRoom, encrypt/decrypt, admin/public material).
  - Invite / link / QR patterns (evolve to include signaling tokens).
  - Local persistence patterns.
- Transport: WebRTC DataChannels (primary) + WebSocket fallback.
- Local relay: refactor existing Go relay into dual-mode (local peer + optional classic).
- Sync kit: evolve LocalFirstDoc or create simpler version that supports local relay transport.
- Meta-app: new or evolved from existing Rooms web (Next.js or Vite/React).

## 8. Implementation Phases & Checklists

### Phase 0: Foundations & Decisions (this PRD)
- [x] Architecture locked (meta-app + code blob + mandatory signaling).
- [ ] Finalize token model for signaling auth.
- [ ] Choose primary transport (WebRTC vs. raw WS with hole-punching).
- [ ] Decide on logical clock library/implementation.

### Phase 1: Core P2P Local Relay MVP
- [ ] Port/refactor Go relay to run locally and speak to other local instances.
- [ ] Implement per-peer queues + cursors.
- [ ] Add basic direct connection (manual IP:port first for testing).
- [ ] Add logical timestamps.
- [ ] Basic offline queue drain on reconnect.
- [ ] Integrate with existing crypto and framing.
- [ ] Simple test: two desktops exchanging updates for a room.

### Phase 2: Mandatory Signaling Service
- [ ] Build minimal signaling server (STUN + candidate exchange).
- [ ] Token validation + rate limiting.
- [ ] Integrate signaling into invite generation.
- [ ] Local relay code always uses signaling for bootstrap (no bypass).
- [ ] Test NAT traversal on typical home + mobile networks.
- [ ] Account/tier integration stub for token issuance.

### Phase 3: Meta-App Shell + Seamless Rooms
- [ ] Basic meta-app with room list, create, invite.
- [ ] Load custom code blob into sandboxed iframe with invisible bridge.
- [ ] Bridge API matching sync kit expectations.
- [ ] Owner can update code blob and push to members.
- [ ] Automatic code propagation.
- [ ] Reuse vault / local storage patterns.

### Phase 4: Mobile & Opportunistic
- [ ] Local relay packaging for mobile (or opportunistic mode).
- [ ] Background / wake-up strategy (or document "connect when active").
- [ ] Test phone ↔ desktop and phone ↔ phone.

### Phase 5: Monetization & Tiers
- [ ] Account system + subscription management.
- [ ] Token issuance gated by tier.
- [ ] Free tier limits (rooms, peers, reliability).
- [ ] Paid tier features (higher limits, better fallbacks).
- [ ] Self-host signaling option (top tier) + documentation.
- [ ] Pricing page / messaging.

### Phase 6: Polish & AI Harness
- [ ] Update LLM harness prompt to produce bridge-compatible code.
- [ ] MCP/tool interface for harness (bring-your-own AI).
- [ ] LAN discovery (nice-to-have).
- [ ] Conflict resolution testing with clocks.
- [ ] Performance validation (bundle size, load time, queue behavior).
- [ ] Security review (sandbox, token handling, crypto reuse).

**Export / Standalone PWA (Post-MVP, after MVP)**
- Export flow to generate standalone bundles or support user-hosted rooms.
- Ensure the signaling requirement is preserved in exported versions.
- This can reuse the exact same code blob + relay + signaling tech.

### Phase 7: Self-Host & Advanced
- [ ] Full self-host instructions for signaling + meta-app.
- [ ] Advanced tier features (multi-device coordination, TURN fallback, etc.).
- [ ] Migration / compatibility with existing Rooms data if relevant.

## 9. Risks & Mitigations

- **Mobile background limitations** — Mitigate with opportunistic mode + clear UX ("syncs when app is open").
- **Signaling reliability becomes single point of frustration** — Invest in monitoring, fallbacks, and clear "why you need the paid tier" messaging.
- **NAT traversal not 100%** — Document fallbacks; higher tiers include TURN.
- **Users want full pure P2P** — Offer self-host signaling at premium; do not expose free bypass.
- **Bundle size / perf** — Validate early with real AI-generated bundles; add compression + lazy loading.
- **Monetization friction** — Keep the gate minimal and the value obvious (reliable connections for your rooms).

## 10. Open Questions

- Exact pricing and free-tier limits?
- Preferred transport stack (WebRTC DataChannel priority)?
- How aggressive to be with phone background execution attempts?
- Should exported standalone PWAs have a time-limited or usage-limited signaling token?
- Multi-device sync coordination details (beyond basic queues)?
- Integration depth with existing tid project (is this a new sub-project or inside current factory/meta structure)?

## 11. Out of Scope (MVP)

- Full enterprise features.
- Large-scale rooms (hundreds of members).
- Built-in video/voice (focus on data sync).
- Public discovery / marketplace of rooms.
- Advanced CRDT features unless needed for conflict cases.
- Export to standalone PWA or user self-hosted rooms (nice-to-have for later; the architecture supports it but is explicitly not in MVP scope. All rooms live inside the meta-app for MVP).

---

**Next Actions After PRD Approval:**
1. Lock this document.
2. Break into engineering tasks / tickets.
3. Prototype Phase 1 (local relay direct connect).
4. Define signaling server MVP spec.
5. Begin AI harness updates for bridge compatibility.

This PRD is intended as the single source of truth for scope while we build. Update it as decisions are made.