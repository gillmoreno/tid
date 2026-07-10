/**
 * Synced Artifacts — Minimal Kit Stub (self-contained friendly)
 *
 * This is an early reference implementation of the public API described in KIT-SPEC.md.
 * Goal: small enough to inline in a single HTML.
 *
 * In a real version we will expand this with:
 * - Proper Web Crypto encryption (AES-GCM)
 * - WebSocket + framing (MSG_UPDATE / MSG_CHECKPOINT)
 * - IndexedDB persistence
 * - Reconnect logic
 * - Optional integration with rooms-style deriveRoom / deriveChannelKey
 *
 * For now: works with localStorage + cross-tab events so you can test the
 * *exact* usage pattern an LLM-generated app will have.
 */

export function initSyncedArtifact(opts = {}) {
  const storageKey = 'sa:' + (opts.appId || 'default') + ':' + getRoomCodeFromLocation();

  let state = load();
  const listeners = new Set();
  let memberId = localStorage.getItem('sa:member') || ('m_' + Math.random().toString(36).slice(2, 10));
  localStorage.setItem('sa:member', memberId);

  function getRoomCodeFromLocation() {
    const hash = location.hash.startsWith('#') ? location.hash.slice(1) : '';
    const params = new URLSearchParams(hash);
    return params.get('c') || params.get('room') || 'demo-room';
  }

  function load() {
    try {
      const raw = localStorage.getItem(storageKey);
      return raw ? JSON.parse(raw) : { _meta: { created: Date.now() } };
    } catch {
      return { _meta: { created: Date.now() } };
    }
  }

  function save() {
    localStorage.setItem(storageKey, JSON.stringify(state));
    // Cross-tab "sync" simulation
    window.dispatchEvent(new CustomEvent('sa:demo-sync', { detail: storageKey }));
  }

  function notify() {
    listeners.forEach(fn => {
      try { fn(getState(), getStatus()); } catch (e) {}
    });
  }

  function getState() {
    return structuredClone ? structuredClone(state) : JSON.parse(JSON.stringify(state));
  }

  function getStatus() {
    return { loaded: true, connected: true }; // real version will report WS status
  }

  function update(mutator) {
    // Simple draft style (the LLM will love this)
    const draft = getState();
    mutator(draft);
    state = draft;
    save();
    notify();
  }

  function subscribe(listener) {
    listeners.add(listener);
    // Immediately call with current
    try { listener(getState(), getStatus()); } catch {}
    return () => listeners.delete(listener);
  }

  // Cross tab (demo only)
  window.addEventListener('storage', (e) => {
    if (e.key === storageKey) {
      state = load();
      notify();
    }
  });
  window.addEventListener('sa:demo-sync', (e) => {
    if (e.detail === storageKey) {
      state = load();
      notify();
    }
  });

  return {
    getState,
    update,
    subscribe,
    getMemberId: () => memberId,
    getStatus,
    setDisplayName(name) {
      localStorage.setItem('sa:name:' + memberId, name);
    },
    getDisplayName(id = memberId) {
      return localStorage.getItem('sa:name:' + id) || undefined;
    }
  };
}

// Optional: a tiny helper the harness can tell the LLM to include
export function buildSafeShareUrl(roomCode) {
  const base = location.origin + location.pathname;
  return `${base}#c=${encodeURIComponent(roomCode)}`;
}
