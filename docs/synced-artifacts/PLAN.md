# Synced Artifacts / Custom Rooms — Plan (Refined)

**STATUS: COMMITTED** (git 9075308 on 2026-07-10)

See also: 
- P2P-RELAY-IDEA.md (exploration)
- PRD-p2p-local-relay-monetization.md (detailed build spec with checklists)

**Locked decisions:**
- Mandatory signaling service for all initial connections (no free pure-P2P bypass).
- Local relay runs on user devices; actual data/code sync is P2P.
- Monetization primarily via the signaling service (tiers + self-host at higher levels).

This architecture direction is locked in per discussion:
- Code stored directly as compressed encrypted blob in room state
- Automatic UI updates to members
- Seamless transparent execution (iframe hidden from user)
- Meta-app centralizes everything; no per-room Git or subdomains

Ready for next (crazier) ideas.

**Core Vision (updated 2026-07-09):**  
A modern take on the original "Rooms" idea — **a meta-app that contains many custom apps ("rooms")**. 

- The meta-app (the "Rooms shell") provides the social layer, vault, contacts, room management, QR/links, and background sync.
- Each "room" inside it can be **arbitrary custom frontend code** (React, Svelte, vanilla, etc.) generated via an AI harness.
- No more schema constraints killing the design. Full visual and interaction freedom per room.
- Every custom room still gets E2EE sync + client-owned data via the same dumb blind relay model that made original Rooms powerful.
- Each custom room can potentially be installed as its own PWA.
- LLM harness is first-class, with a path for users to bring their own AI (MCP/tool calling) to avoid the builder bearing generation costs.

This keeps everything great about Rooms (relay blindness, ownership, easy sharing via safe links + QR, solo-first, small-group collab) while fixing the main shortcoming: lack of custom beautiful UIs.

**Core invariants (non-negotiable):**
- All data over the wire is encrypted. The relay is completely blind (same as original Rooms).
- Clients hold the canonical data.
- Relay is self-hostable (start with the proven Go implementation).
- LLM harness is first-class. Users should be able to bring their own AI (MCP/tool) so the builder doesn't pay for generation.
- Full custom code freedom per room (no schemas that dictate UI).
- Small-to-medium scope per room: forms to 3-5 "page" collab tools (expenses, lists, polls, light boards, etc.).
- Meta-app structure: One shell that hosts many custom-code rooms (inspired by original Rooms "app that contains apps").

## Key Decisions & Refined Direction (2026-07-09)

The pure "standalone single HTML" path is de-emphasized. The stronger model is a **meta-app** (Rooms-like shell) that hosts many custom-code rooms.

### Meta-App Structure ("An app that has custom apps")
- Central shell (React/Vite PWA, similar to current tid or original Rooms web).
- Users see "My Rooms", create new ones via AI harness, manage vault, contacts, etc.
- Each room loads/runs its own custom frontend code + the thin sync kit.
- The shell provides consistent social layer (safe links, QR, member profiles, invites) while the room itself is fully custom.
- Analogy to original Rooms: one meta experience, many independent "apps" (now with arbitrary code instead of templates/schemas).

### Custom Code Loading & Security
- Custom code must be loadable safely inside (or alongside) the meta-app.
- Options to explore: sandboxed iframes + postMessage bridge for the sync kit, Web Workers, or treating each room as an installable sub-PWA.
- The generated code must be "room-ready": it calls the standard sync kit API and expects a roomCode from the shell or URL hash.

### Creation Flow
- Inside the meta-app (or via external MCP/tool): user describes the desired room ("beautiful RSVP for cabin trip with live counts and dietary notes").
- Harness (AI) generates the custom code bundle.
- Code is "installed" into the user's vault as a new room (stored encrypted, associated with a relay room ID).
- User can immediately open it, share it, etc.

### PWA per Room
- Promising idea: each custom room can be exported/installed as its own standalone PWA.
- Benefits: full app icon, better storage guarantees, feels like a real dedicated app.
- The meta-app can help generate the manifest + standalone bundle (with the sync kit + room secrets injected safely via hash).

### Deployment, Code Storage, Sandboxing & "The JSON Equivalent" (Critical Section)

**The core worry**: Old Rooms was simple because each room was basically structured JSON data stored in the Yjs document (meta + template.* maps). No real "code deployment". Everything was data.

**The equivalent here**: Treat the **entire custom frontend** as just another piece of (encrypted) data inside the room.

#### Storage Model ("Code as Encrypted Data")
- In the room's state (the same encrypted Yjs doc / op log the relay sees):
  - `code`: { version: number, bundle: string (compressed), updatedAt, updatedBy }
    - `bundle` = the full self-contained frontend. Stored directly as a blob in the room state.
  - Plus the normal runtime data: `data.responses`, `data.expenses`, etc.
- When the AI harness "creates" a room, it generates the initial `code.bundle` and writes it into the state via the normal `update()` mechanism.
- Updating the UI later = owner does another `update` that replaces `code.bundle` with a new version.
- **Compression**: Bundle will be compressed (e.g. gzip) before storage/encryption. Real-world perf (load time, snappiness, storage size) needs validation with actual generated bundles.
- **Encryption**: The code blob is encrypted with the exact same room key as the runtime data before it ever leaves the client or hits the relay. The relay sees another opaque encrypted frame. No difference.
- **Centralization without sprawl**:
  - No Git repo per room.
  - No subdomain per room.
  - Everything for a room (its code + its data) lives in one relay room ID.
  - The meta-app is the single hosted PWA that knows how to load any room.

**User decision (2026-07-10)**: Store the code blob directly in the room state (compressed). Performance, loading feel, and serving snappiness to be measured in practice.

#### Loading a Custom Room in the Meta-App
1. User opens a room from the meta-app shell (using the normal roomCode + sync kit).
2. Shell receives the full state (including the current `code.bundle`).
3. Shell creates a **sandboxed iframe** (for isolation):
   - `sandbox="allow-scripts allow-forms allow-modals"`
   - Uses `srcdoc` or blob URL from the (decrypted) bundle.
4. Establishes a **postMessage bridge** (strictly validated):
   - The iframe receives state, member info, status.
   - The iframe sends only approved `update()` mutations.
5. The custom code inside the iframe uses a tiny bridge library that feels identical to the normal kit:
   ```js
   const room = initRoomBridge(); // postMessage under the hood
   room.update(d => { d.responses.push(...) });
   room.subscribe(s => render(s));
   ```
6. The shell owns the real sync kit, persistence, and relay. The bridge is invisible to the end user.

**User decisions (2026-07-10)**:
- Code updates are pushed automatically to members (no "accept" step).
- The iframe + bridge must be completely transparent — the user should not perceive any difference from a native app.
- Sandboxing will be used for safety; the exact strictness (pros/cons of tighter vs looser policies) needs further discussion. The goal is strong isolation while keeping the experience seamless.

#### Versioning & "Deployment"
- Versioning is just state updates. Every time the owner changes the UI, they publish a new `code.version`.
- Because it's inside the CRDT/op-log, other devices get the new code automatically when they sync.
- History of code versions lives in the room (you can roll back if wanted).
- No separate deploy step. "Deploying" a new UI = updating the code field in the room state.
- **User decision**: Members automatically receive the new UI. No "accept version" prompt.

#### Sandboxing Details (iframe + postMessage)
- We use a sandboxed iframe + postMessage bridge for isolation.
- **User note**: The exact level of strictness needs more discussion (pros/cons of tighter policies). The goal is strong security while making the experience feel completely native.
- The bridge must feel invisible:
  - Custom code inside the iframe should not perceive any difference from running natively.
  - Strict message validation on the parent side.
  - No direct DOM or relay access leaked to the iframe.
- Since code originates from the room owner (via AI), risk is contained.

#### Standalone / Exported PWA Path (Post-MVP)
- Export to standalone PWA or user self-hosted site is explicitly out of MVP scope.
- Nice-to-have for later: users can create their own PWAs or host rooms on their own websites using the same code-blob + local-relay technology.
- For MVP: everything lives inside the meta-app. No export flow required.

#### Why This Avoids the Nightmares
- Endless GitHub repo? Code lives encrypted inside the room's relay data. No source control sprawl unless the user explicitly exports.
- Subdomain per app? One meta-app domain. All rooms are just different relay room IDs + different code blobs inside their state.
- Deployment complexity? Updating a room's UI is the same operation as adding a response — a normal encrypted state update.

This keeps the "app of apps" spirit of original Rooms while allowing real custom code. The "JSON" has been upgraded to "encrypted code bundle + data".

**User decisions locked (2026-07-10)**:
- Store code blob directly in room state (with compression; real perf to be validated).
- Automatic propagation of code updates to all members.
- The bridge experience must be seamless — no visible iframe artifacts for end users.

### LLM Harness & Cost Model
- First-class.
- Strong preference to avoid the operator bearing AI generation costs.
- Primary path: expose the harness as an **MCP server / tool** so users connect their own Claude, Grok, local models, etc.
- Secondary: optional hosted generation inside the meta-app (with clear costs or limits).
- The harness must output code that works inside the meta-app (embedded via the bridge). Support for exported standalone PWAs is post-MVP.

### Sharing & Social
- Inherit the best of original Rooms: URL hash for secrets, client-side QR, WhatsApp-friendly links, member attribution.
- Two main sharing modes:
  1. Via the meta-app (deep link like `/room/abc123` — recipient must have the meta-app or it prompts install).
  2. As a standalone PWA/exported bundle (recipient opens the link or installs directly).
- Discovery inside the meta-app (contacts, recent rooms, etc.) is a nice-to-have.

### 1. CRDT
(kept from previous; see full details in NOTES file) Default to simple op-log for most generated rooms. Optional stronger merging hidden in the kit. Relay remains byte-agnostic.

### 2. Deployment & Hosting
- Meta-app itself is hosted (one PWA).
- Individual rooms: either run inside the meta-app shell or exported as their own static + PWA bundles.
- Pure raw single-file HTML is still possible for maximum portability but is no longer the primary target.
- Storage: Full IndexedDB power when running as hosted PWA (meta or per-room).

### 3. State / Sync Model
- Keep relay exactly as-is (Go, framed encrypted blobs, checkpoints, limits).
- Client kit exposes a tiny, stable, LLM-friendly API (e.g. `update(mutator)`, `getState()`, `subscribe`, `getMemberId()`).
- All business logic, UI, derived views stay 100% in the generated frontend code.
- Encryption + framing can reuse (or closely mirror) the rooms room-kit crypto + relayProtocol.

## Tasks (A, B, C, D) — Updated for Meta-App Direction

**A. LLM Harness (first-class, MCP-friendly)**
- System prompt + few-shot that produces "room code" suitable for both:
  - Loading inside a meta-app shell.
  - Export as standalone PWA.
- Must include: proper use of the sync kit, safe hash-based sharing, member attribution, PWA manifest hints, graceful loading states.
- Support for being called via MCP/tool (structured output that can be consumed by Claude/Cursor/etc.).

**B. Minimal Client Sync Kit + Shell Bridge**
- Tiny API for custom room code (`update`, `getState`, `subscribe`, etc.).
- Works when the room runs inside the meta-shell (via postMessage or direct injection) **and** standalone.
- Local persistence strategy that works for both modes.
- Sharing helpers (build safe link with hash, QR generation).

**C. Meta-App Shell Spike + Room Integration**
- Basic meta-app shell (list of rooms, create via harness stub, open a room).
- Demonstrate loading a custom code "room" (start with the existing RSVP spike adapted).
- Per-room PWA export flow (generate manifest + standalone bundle).
- Vault + room management using the same patterns as original Rooms.

**D. Supporting Work**
- Security model for loading custom code (iframes? sandboxing?).
- MCP server/tool definition (how users connect their own AI).
- Cost model notes (user pays for their AI calls).
- Social layer inheritance (deep links inside meta-app vs exported PWAs, contacts/personas).
- Relay: still the Go one; multiple custom rooms just use different derived IDs.
- PWA export mechanics and storage guarantees.
- Deployment: where the meta-app lives (inside tid? separate?).
- Sharing UX exploration (meta-app links vs standalone installs).

## Success Criteria (MVP)
- A meta-app shell exists where a user can "create room with AI" (via harness stub or MCP simulation) and get a working custom-code room.
- The custom room has full UI freedom, uses the sync kit, and syncs via the blind Go relay.
- Sharing works (safe hash links + QR) both inside the meta-app and for exported standalone versions.
- At least one room can be exported/installed as its own PWA.
- Data ownership invariants preserved (E2EE, client storage, self-hostable relay).

## Open Questions & Later Work
- Full hosting flow + cost model.
- Real MCP server implementation for the harness.
- Sandboxing story for custom code in the meta-app (iframe + postMessage details, threat model).
- Iterative refinement loop inside the shell ("this looks good, make the colors warmer and add a total at top").
- When / how to support exporting a room as a fully independent PWA vs always requiring the meta-app.
- Cross-room features (contacts, personas) from original Rooms.
- Where this lives in the tid / the-idea-guy ecosystem.
- Performance: code bundle size limits, compression, lazy loading of code.
- Migration / evolution of a room's code over long periods.

## Files to Create / Modify (initial)
- `docs/synced-artifacts/PLAN.md` (this, refined for meta-app model)
- `docs/synced-artifacts/LLM-HARNESS-PROMPT.md` (update for meta-shell + PWA export + MCP)
- `docs/synced-artifacts/KIT-SPEC.md` + shell bridge
- `docs/synced-artifacts/spikes/` (adapt RSVP as loadable room + new meta-shell spike)
- Security & loading model notes
- MCP harness spec
- PWA export mechanics notes
- Possibly integrate relay or reference the Go code from the-idea-guy/projects/rooms/relay/

Start executing A/B/C/D autonomously after this plan is reviewed in conversation.
