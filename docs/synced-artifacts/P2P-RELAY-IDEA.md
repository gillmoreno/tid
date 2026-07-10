# P2P Local Relay Idea (Crazy Direction)

**Date:** 2026-07-10  
**Status:** Initial exploration after committing the meta-app + code-as-data model.

## The Vision (in user's words)

- Run the relay **on the user's own device** (phone or computer).
- No centralized relay server at all.
- Peer-to-peer connections between the specific people who share a room.
- One local relay can handle multiple independent P2P "channels":
  - Expenses room with ex-wife
  - Movie club room with 4 friends
- Updates are queued locally per peer.
- When peers come online, they handshake (using shared crypto material) and exchange queued messages.
- Timestamps / logical clocks for ordering and conflict resolution.
- Each person owns their copy of the data + their copy of the relay.

Goal: Take the "blind relay" concept to its logical extreme — the relay itself becomes personal infrastructure.

## High-Level Architecture

```
Device A (you)                          Device B (ex-wife)
┌─────────────────────┐                ┌─────────────────────┐
│  Meta-App Shell     │                │  Meta-App Shell     │
│  - Custom rooms     │                │  - Custom rooms     │
│  - Custom code blob │                │  - Custom code blob │
│                     │                │                     │
│  Local Relay        │◄──────────────►│  Local Relay        │
│  (Go binary)        │   P2P channel  │  (Go binary)        │
│  - Per-room queues  │   (WS/WebRTC)  │  - Per-room queues  │
│  - Handshake logic  │                │  - Handshake logic  │
│  - Logical clocks   │                │  - Logical clocks   │
└─────────────────────┘                └─────────────────────┘

Each device also has its own local persistence (IDB / files)
for its copy of the encrypted state + queues.
```

The "relay" is no longer a separate service you point at — it is a small always-available component on your machine that knows how to talk directly to the specific peers for each room.

## Key Technical Pieces

### 1. Transport (the hard part)
Current relay uses plain WebSocket (centralized listener).

For P2P:
- **WebRTC Data Channels** — best for real-world NAT traversal (uses STUN for free, optional self-hosted TURN).
- Or raw WebSocket/TCP with manual hole punching / UPnP / port forwarding.
- Hybrid: try direct P2P first, fall back to a user-run "personal relay" on a home computer if phones are too constrained.

### 2. Discovery & Handshake + Minimal Signaling Service
No global directory for data.

**User decision (2026-07-10)**: There must be a monetizable path. Pure 100% peer-to-peer (no signaling service at all, even for the first handshake) is **not offered**. Every initial connection must go through the controlled signaling service as a minimal gate. This creates "minimal gating" for monetization.

- The signaling service is mandatory for bootstrap / NAT traversal / candidate exchange.
- Once the direct P2P connection is established, the service is no longer involved.
- Higher tiers can offer premium reliability, more concurrent connections, self-hosted signaling support, or hosted meta-app + signaling bundles.
- Self-hosting the signaling is possible at higher tiers, but the default path always requires going through the service (or a paid self-host integration).
- Goal: lots of users + revenue without ever hosting or seeing user data/code.

Sharing flow (evolution of current Rooms invites):
- When creating a room, generate a P2P invite containing:
  - Room identifier (derived like today)
  - Cryptographic material (room secret / public key material)
  - Optional reachability hints or a token for the signaling service
- Share via QR / hash link / out-of-band.
- Recipient imports → local relays use the signaling service (if authorized) to exchange connection candidates (WebRTC ICE or equivalent).
- Once candidates exchanged, direct P2P connection is attempted (WebRTC DataChannel preferred for NAT punching).
- Mutual authentication using the shared secrets (same derivation as current Rooms crypto).
- After first successful direct connection, peers can remember addresses for future direct attempts.

The signaling service is **only for bootstrap** — it never sees encrypted data blobs, never stores room content, and can be rate-limited per paid account.

Handshake needs:
- Exchange of addresses / connection candidates via signaling.
- Key confirmation.
- Once established, the connection is treated like a "peer" in the old relay hub. Subsequent messages go direct P2P.

**Monetization angle** (to be explored next):
- Free tier: limited concurrent connections or slow/unreliable signaling (encourages self-host or manual).
- Paid: reliable always-on signaling, better STUN/TURN fallback, higher limits, perhaps hosted meta-app + signaling bundle.
- Self-host option: users run their own tiny signaling instance (open source the signaling server code).
- This creates a "tiny bit that depends on me" without compromising the core data ownership (data and code blobs never touch the signaling server).

### 3. Queues & Offline Sync
- Each local relay maintains a **per-peer, per-room outbound queue**.
- When you make a change (add expense, update code blob, etc.):
  - It is applied locally.
  - The encrypted frame is appended to the queue for every peer in that room.
- When a peer connects (or reconnects):
  - Exchange "what have you seen?" cursors (similar to current sync cursors or Yjs state vectors).
  - Send queued frames.
  - Receive their queued frames.
  - Acknowledge receipt so queues can be pruned.

This is basically a personal store-and-forward relay.

### 4. Conflict Resolution
- Every message carries a timestamp.
- Use **hybrid logical clocks** (or simple Lamport timestamps + device ID) so we can order events even without wall-clock sync.
- Combine with the existing merge strategy (op log or CRDT underneath).
- For simple cases: last-writer-wins per item + full history log for audit.
- Owner of a room can still have special authority if desired.

### 5. Multiple Rooms on One Local Relay
The local relay becomes a small "hub" on your machine:
- It listens for inbound P2P connections.
- It initiates outbound connections for rooms you care about.
- Rooms are isolated by their derived room ID + keys (same as today).
- One process can manage dozens of independent P2P channels.

## Integration with the Committed Meta-App Model

This is an evolution of the **transport layer**, not a full rewrite.

- The sync kit (the thing custom code talks to) stays mostly the same API.
- Underneath, instead of "connect to wss://central-relay", it connects to "local relay process" which then does P2P.
- Code blobs, data, everything is still stored the same way (encrypted in the room state).
- The meta-app shell still provides the UI for managing rooms, generating invites, showing QR, etc.
- When you export a room as standalone PWA, it can bundle a small local-relay component or assume the user has the relay running.

The "blind" property is preserved (or actually improved) because there is no third-party relay at all.

## Feasibility Assessment

**Possible?** Yes.

**Easy?** No — this is significantly harder than a centralized (or self-hosted) dumb relay.

Biggest challenges:
1. **Connectivity on real networks** (especially mobile).
   - Phones go to sleep, change networks, are behind carrier NAT.
   - Solution space: background service + push for signaling, always-on home hub as "anchor", WebRTC + TURN (self-hosted or paid), or accept that sync only happens when both devices are reasonably online.
2. **Signaling / initial contact** without any central service.
   - Purely manual (share current IP + port + key material) works for tech users.
   - For better UX you probably still want a tiny, optional public signaling service just for connection setup (STUN is free and public; a small signaling server can be run by you or offered as a free bootstrap).
3. **Running a server on phones**.
   - Desktop: trivial (just run the Go binary).
   - Phone: needs special packaging (Termux, custom app, or the local relay lives only on a home computer and phones act as clients that wake it up).
4. **Multiple devices per person**.
   - If you have phone + laptop, do they both run independent relays? How do they coordinate so the ex-wife doesn't get duplicates?
   - This might require a small "device mesh" per person.

Many successful local-first apps have gone down similar roads (e.g., some use libp2p, some use Syncthing-style protocols, some use Matrix as a P2P-friendly transport, some just accept a personal always-on server).

## Relation to Current Rooms Code

The existing `relay/main.go` + framing + crypto is an excellent starting point.

We can refactor it into something that can run in two modes:
- Classic "dumb central" mode (for easy onboarding).
- "Local P2P peer" mode (the crazy idea).

The framing (MSG_UPDATE / MSG_CHECKPOINT), encryption, and deriveRoom logic can stay almost identical. Only the connection management changes from "accept many WS clients for one room" to "maintain direct P2P links per known peer for that room".

The LocalFirstDoc / sync logic in the kit can be extended to support a local transport.

## Next Steps if We Pursue This

- Prototype a local relay that can speak the current framing (manual direct connect first, then add signaling).
- Add per-peer queues + cursors.
- Add logical clocks / timestamps.
- Integrate minimal signaling server (cheap to run, rate-limited by account).
- Design paid signaling auth (invite tokens or account-linked).
- Evolve the invite format to carry signaling tokens.
- Explore monetization packaging (signaling tiers, self-host instructions, bundled with meta-app hosting).
- LAN discovery as nice-to-have.

This direction turns the "tiny public signaling service" into the monetization surface while keeping data and code fully P2P and user-owned.

---

This is a very pure realization of the original local-first + ownership philosophy. It trades simplicity and reliability for maximum decentralization.

The current committed model (meta-app + blind relay + code-as-data) is still extremely valuable as the "easy path". This P2P version can be an advanced / power-user / self-hosted mode.

What do you want to explore next on this idea? Or jump straight to one of the other crazier ideas?