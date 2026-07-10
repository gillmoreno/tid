# Notes: CRDT, HTML Storage, and Sharing (from discussion 2026-07-09)

## 1. CRDT — What was the actual situation in Rooms?

Rooms used **Yjs** (a mature CRDT library) wrapped in `LocalFirstDoc`.

**What worked extremely well:**
- True offline editing on multiple devices.
- Conflict-free merging when people edited at the same time (two kids checking chores, two people adding expenses).
- The relay stayed 100% blind — it only ever saw encrypted Yjs update bytes or checkpoints.
- Client-driven compaction (send a clean full state → relay replaces its log).

**The "problems" (tradeoffs), especially for the new Synced Artifacts direction:**

- **Complexity for LLM generation**  
  Even with a nice wrapper, the model has to understand when and how to mutate the CRDT document. More surface area = more ways for the generated code to be subtly wrong.

- **Bundle & single-file weight**  
  yjs + y-indexeddb add kilobytes. For a beautiful self-contained HTML that someone might send over WhatsApp or open from a download, every byte matters.

- **History bloat**  
  Yjs keeps edit history. For forms with images or long-running rooms you need explicit compaction logic (Rooms already had this, but it's extra code the LLM would have to get right or we hide).

- **Overkill for many target cases**  
  Most of the early use cases ("are you coming?", sign-up lists, simple shared expenses) are mostly *append* or *add record*. A simple operation log + deterministic merge is dramatically easier to reason about and generates smaller, more debuggable code.

- **Magic factor**  
  CRDTs are powerful but feel magical when something merges in an unexpected way. For small apps we often prefer boring and predictable.

**Our decision:**
- Default path = **simple state + op log** (or even full-snapshot + last-writer-wins on keys for the simplest forms).
- The public API the LLM sees (`room.update(draft => {...})`) stays the same.
- We can later add an opt-in `strong: true` that switches the kit to use Yjs (or another CRDT) *under the hood* without the generated UI code changing.
- The relay doesn't care — it transports bytes either way.

Rooms proved the overall architecture works with CRDT. We're keeping the architecture and selectively using less powerful (but simpler and lighter) merging for the 80% case.

## 2. Single HTML / Static File — Storage & Persistence

Rooms relied heavily on **IndexedDB** (via y-indexeddb) + localStorage for the vault.

**When the artifact is hosted** (recommended):
- Stable origin → excellent IndexedDB behavior.
- Large quota.
- Survives browser restarts, works great in multiple tabs.
- Easy to turn into a PWA (Add to Home Screen) for an app-like icon and slightly better persistence guarantees.

**When someone downloads the .html and opens it directly (file:// or from a random server):**
- IndexedDB technically works in modern Chrome and Firefox, but the origin is "opaque".
- Behavior is less reliable across browsers and sessions.
- Data is tied to that specific browser profile + sometimes the file path.
- If the user clears "site data" or uses private/incognito, it disappears more easily.
- No good PWA install story.

**Practical recommendation we will bake into the harness and docs:**
1. Primary experience = hosted link (we can provide one-click hosting or the user hosts the generated bundle anywhere static).
2. The generated HTML should still be fully self-contained so it also works when downloaded.
3. Inside the kit we will do: IndexedDB first → graceful fallback to localStorage.
4. We will gently encourage "Add to Home Screen" when hosted.
5. For pure file sharing we accept slightly weaker local persistence (the important data is still safe in other people's browsers via the relay).

## 3. Shareability & Socialization — We Should Inherit Almost Everything from Rooms

Rooms did this part **very well** and it is one of the biggest things we want to keep:

- Secrets (room code, admin secret, passphrase hint) **only in the URL hash**.
  - Example: `https://.../my-form.html#c=amber-tiger-maple&admin=...`
  - The `#` part is never sent to any server. Perfect.

- Client-side QR codes that encode the full safe link.
- Simple "Copy invite link" that just reads the current URL (hash included).
- Member IDs generated locally + optional display names.
- Clear mental model: "Whoever has the link can participate."

**For the HTML artifact world we will do the same:**
- The kit will help build safe share URLs.
- Every generated app (via the harness) will include a standard "Share this" section with Copy + QR.
- The HTML itself can be generic. The specific room instance travels in the hash the creator shares on WhatsApp, Telegram, etc.
- If sharing the raw .html file, the receiver can still join by using the same room code (the creator just tells them or the link is documented inside the page).

We can improve on Rooms slightly:
- Make the QR + copy link components extremely easy to drop in (or auto-include in the harness template).
- Support a tiny "this link only works for responses, not editing the form" distinction if we add a light admin channel later.
- Optional passphrase flow (the `pp=1` hint Rooms used) so the joiner knows they need the extra secret.

The thing that was missing in Rooms was the freedom to make the actual page beautiful and custom. With this new approach we get the excellent sharing model + arbitrary beautiful frontends.

## Summary for Implementation

- CRDT: available as a power tool, not the default hammer.
- Storage: push hosted experience + robust fallbacks in the kit.
- Sharing: copy the Rooms discipline (hash secrets, client QR, easy links) and make it a first-class part of every generated artifact.
- The relay and encryption model stay exactly as powerful and ownership-focused as before.

This lets us say: "Rooms was excellent at the hard parts (sync, encryption, sharing, ownership). We are keeping all of that and removing the only thing that held it back — the restrictive UI layer."
