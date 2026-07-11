# Security & Sandboxing for Custom Code Rooms

## Model
- Custom frontend code runs in a **sandboxed iframe**.
- `sandbox="allow-scripts allow-forms allow-modals"`
- No `allow-same-origin` (stronger isolation).
- Communication only via a tightly controlled `postMessage` bridge.
- The bridge only accepts specific message types from the iframe origin (for the spike we use '*" + validation on shape; real version will use more origin + nonce checks).

## What the iframe CANNOT do
- Access the parent DOM
- Access the real relay WebSocket / WebRTC / local relay process
- Access other rooms' state
- Reach `localStorage` / `IndexedDB` of the meta shell origin (it gets its own opaque origin when sandboxed)
- Execute arbitrary navigation or top-level redirects that escape the shell (modals are allowed for UX)

## What the iframe CAN do (via bridge)
- Read current app state
- Propose mutations via `update(draft => ...)`
- Receive state updates pushed by the shell
- Use normal web APIs inside its own document (canvas, storage inside iframe, etc.)

## Code provenance
- Code blobs originate from:
  - The room owner (via AI harness they called with their own tokens, or hand-written)
  - Subsequent owner updates
- Risk is contained to the people who were explicitly invited to the room (they already have the room secret).
- The shell can still offer "review diff" or "pin version" as a future power-user control, but per locked decisions **code updates propagate automatically** with no accept step.

## Future hardening options (not MVP)
- Content-Security-Policy inside the generated srcdoc (restrict connect-src, script-src, etc.).
- Subresource integrity if we ever serve kits externally.
- Capability flags in the bridge (read-only rooms, owner-only mutations).
- Static analysis or runtime guards on the generated bundle before first load.

## Data model security
- Everything that leaves the client (code or data) is encrypted with the room key before the relay or peers ever see it.
- The relay (central or local P2P) is completely blind.
- Losing all copies of the room secret + local data = permanent loss (by design).

See PLAN.md for the locked user decisions around automatic propagation and seamless experience.
