# LLM Harness Prompt — Synced Artifacts (Meta-App + Custom Rooms)

This prompt turns a natural language request into a **beautiful, fully custom room** that runs inside the Rooms meta-app shell (or can be exported standalone later).

The meta-app is "an app that contains many custom apps". Each room is arbitrary custom frontend code stored as an encrypted blob inside the room state. Code updates are just another state change — members receive the new UI automatically with zero deploy steps.

## Primary target (MVP)
- The generated code will be loaded by the meta shell into a **sandboxed iframe**.
- It must call the bridge API: `const room = initRoomBridge();`
- The experience must be completely seamless (no visible iframe borders or "app inside app" feel).
- The shell provides consistent social features (share link, QR, member attribution).
- The custom code owns 100% of the UI, layout, logic, and derived views.

## Also support (post-MVP / export)
- Same code should be able to run standalone as its own PWA when given a real `initSyncedArtifact` kit instead of the bridge.

The generated output is typically the **inner content** (markup + script) that the shell wraps. It can also be a full self-contained HTML for direct use or export.

## Non-Negotiable Rules (the model must obey these)

1. **UI Freedom is absolute.** Build whatever beautiful interface fits the request. No schema restrictions. Make it feel like a dedicated premium app.

2. **All logic in the frontend.** Calculations, validation, derived views, animations — all in the generated JS. The relay and shell never execute business logic.

3. **Use the bridge / kit API only.** In the meta shell you MUST use:
   ```js
   const room = initRoomBridge();
   room.update(draft => { ... });
   room.subscribe(s => render(s));
   room.getMemberId();
   ```
   Never reach for window.parent, postMessage, WebSocket, or crypto yourself.

4. **State is a plain JS object.** You design the exact shape. Keep it JSON-serializable.

5. **Every persistent change goes through `room.update(mutator)`.** The mutator receives a draft you can mutate directly (Immer-style). This is how code + data sync.

6. **Graceful loading & connection states.** Use `room.getStatus()` and the subscribe callback.

7. **Member attribution + display names.** Use `getMemberId()` and `setDisplayName(name)`.

8. **Sharing is provided by the shell.** You may still include nice "Share" affordances. The shell owns the canonical safe link + QR. For standalone/export builds you must include the copy link + QR yourself.

9. **Offline-first.** Always render from local state immediately.

10. **Code updates are automatic.** When the room owner publishes a new `code.bundle`, members receive it on next sync / reload of the room. Do not add "accept new UI" flows.

11. **Tailwind via CDN or clean inline styles is encouraged** for beautiful low-dependency rooms.

12. **No external services.** Only the sync bridge/kit talks to the outside world.

## The Standardized Sync API (bridge when inside meta shell)

When loaded inside the meta-app shell, `initRoomBridge()` is provided for you (injected by the shell).

```js
const room = initRoomBridge();   // always available inside the shell iframe

room.getState()                  // your app data object
room.update(draft => {           // mutate the draft — this is the primary mutation API
  draft.responses = draft.responses || [];
  draft.responses.push({...});
});
room.subscribe((state, status) => { /* re-render your UI */ });

room.getMemberId();
room.getStatus();                // { loaded, connected }
room.setDisplayName(name);
room.getDisplayName(id);
```

The same conceptual API will be exposed by the real client kit (`initSyncedArtifact`) for standalone / exported PWAs. Use the mutator + subscribe style everywhere.

When running standalone the kit will handle real encryption + relay/P2P connection. Inside the shell the bridge forwards everything to the shell's sync kit. The custom code never knows the difference.

Example usage inside your beautiful form:

```js
// When someone submits the interest form
function submitResponse(responseData) {
  room.update(s => {
    s.responses = s.responses || [];
    s.responses.push({
      id: crypto.randomUUID(),
      ...responseData,
      by: memberId,
      at: Date.now()
    });
  });
}

// In your render / effect
room.subscribe(s => {
  renderResponsesList(s.responses || []);
  renderSummary(computeWhoIsGoing(s.responses)); // your pure function
});
```

**Important for the model:**
- Treat `room.update(...)` like a setState that also syncs.
- You can read `room.getState()` at any time.
- Put derived/computed values in your render functions — do **not** store them in state unless necessary.
- For lists that multiple people append to, just push. The kit + simple merge rules will handle most cases.

## Few-Shot / Characteristics (meta-shell edition)

The generated code is the inner experience. The shell wraps it and provides the bridge.

Example request: "A beautiful RSVP form for our weekend trip to the cabin. People say if they're coming, how many, dietary notes, and a fun comment."

Excellent output:
- Premium hero with the event details (use data from state so owner can later mutate event info too if desired).
- Form that on submit calls `room.update(d => { d.responses.push(...) })`.
- Live list + computed totals (pure functions in your render).
- Subtle connection / member count.
- No direct DOM-to-relay anything.
- Uses Tailwind CDN when helpful for polish.
- Small delightful details (empty states, optimistic UI, attribution using memberId + display names).

For expenses: implement balances and "who owes what" completely in JS from the list of expenses in state.

The shell (or exported kit) will supply sharing, QR, vault, and cross-device sync. Your job is the custom UI and domain logic.

## Creation & Update Flow (meta-app)

1. User (inside meta shell or via MCP tool) describes the desired room.
2. Harness returns the custom code (inner markup + script using `initRoomBridge`).
3. Shell writes it as `code.bundle` into the room's encrypted state (same as any data update).
4. The room immediately becomes available to the creator and (after sync) to members.
5. Later the owner can say "make the colors warmer and add a running total" — the harness produces a new bundle. Shell publishes it as a normal state update. Everyone receives the new UI automatically.

No Git, no subdomains, no separate deploy. Code is just another encrypted field in the room.

## Extra Guidance for Quality

- Make the UI feel premium and specific to the request (good emojis, clear labels, nice empty states, animations if they add delight without complexity).
- Mobile-first responsive.
- Clear "This data lives only on your devices and is end-to-end encrypted" trust note (small, tasteful).
- When the app first loads with no room, auto-generate a nice room code and put it in the hash (so refreshing keeps the instance).
- Support a simple local "preview mode" if the relay is not reachable (still useful).
- For owner actions (e.g. close the form, reset), you can store an `ownerId` or use a simple admin secret the kit supports.

## Output Shape Guidance

- Preferred for meta-shell: return the **inner document fragment** (top-level divs + styles + one script block). The shell will wrap it + inject the bridge + Tailwind when needed.
- For standalone/export: you may emit a complete `<!doctype html>` ... `</html>`. It should still call the same API (the standalone kit will polyfill `initRoomBridge` as `initSyncedArtifact` or provide a compat shim).
- Always include a visible trust note somewhere tasteful: "End-to-end encrypted. Data lives only on participants' devices."

## What NOT to Do

- Do not output a dead form that only does `console.log` or `alert`.
- Do not hardcode sample data that can't be changed.
- Do not talk to any server except through the bridge/kit.
- Do not use heavy frameworks (keep small and self-contained).
- Do not put room secrets in the visible URL path or in the emitted HTML source.

---

**Implementation note:**
The harness is first-class and should be callable via MCP so the user can bring their own model (Claude, Grok, local, etc.). The builder never pays for generation in the primary path.

Store this prompt + a couple of full generated examples as few-shot. Support iterative refinement inside the shell ("this looks good, make the header use a warmer palette and add a total at the top").
