import { randomToken } from './crypto'
import type { CounterState } from './types'

const CHANNEL = 'meta-room'
const MAX_MESSAGE_SIZE = 8_192
const subscriptions = new WeakMap<object, () => void>()

function postToFrame(source: MessageEventSource, message: unknown): void {
  (source as Window).postMessage(message, '*')
}

export type RoomBridge = {
  getState: () => Promise<CounterState>
  update: (operation: { type: 'counter.increment' }) => Promise<CounterState>
  subscribe: (listener: (state: CounterState) => void) => () => void
}

type BridgeRequest = {
  channel: typeof CHANNEL
  version: 1
  nonce: string
  type: 'bridge.request'
  requestId: string
  method: 'getState' | 'update' | 'subscribe'
  payload?: { type: 'counter.increment' }
}

function isBridgeRequest(value: unknown, nonce: string): value is BridgeRequest {
  if (!value || typeof value !== 'object') return false
  const message = value as Record<string, unknown>
  if (
    message.channel !== CHANNEL
    || message.version !== 1
    || message.nonce !== nonce
    || message.type !== 'bridge.request'
    || typeof message.requestId !== 'string'
    || message.requestId.length > 128
  ) return false
  if (!['getState', 'update', 'subscribe'].includes(String(message.method))) return false
  if (message.method === 'update') {
    if (!message.payload || typeof message.payload !== 'object') return false
    const payload = message.payload as Record<string, unknown>
    if (payload.type !== 'counter.increment' || Object.keys(payload).length !== 1) return false
  } else if (message.payload !== undefined) {
    return false
  }
  return true
}

export async function handleBridgeMessage(
  event: MessageEvent<unknown>,
  expectedSource: MessageEventSource,
  nonce: string,
  bridge: RoomBridge,
): Promise<boolean> {
  if (event.source !== expectedSource) return false
  try {
    if (JSON.stringify(event.data).length > MAX_MESSAGE_SIZE || !isBridgeRequest(event.data, nonce)) return false
  } catch {
    return false
  }

  const request = event.data
  const respond = (ok: boolean, state?: CounterState, error?: string) => {
    postToFrame(expectedSource, {
      channel: CHANNEL,
      version: 1,
      nonce,
      type: 'bridge.response',
      requestId: request.requestId,
      ok,
      state,
      error,
    })
  }

  try {
    if (request.method === 'getState') respond(true, await bridge.getState())
    if (request.method === 'update') respond(true, await bridge.update(request.payload!))
    if (request.method === 'subscribe') {
      disconnectBridgeSource(expectedSource)
      const unsubscribe = bridge.subscribe((state) => {
        postToFrame(expectedSource, {
          channel: CHANNEL,
          version: 1,
          nonce,
          type: 'bridge.state',
          state,
        })
      })
      subscriptions.set(expectedSource as object, unsubscribe)
      respond(true, await bridge.getState())
    }
  } catch {
    respond(false, undefined, 'The room could not apply this update.')
  }
  return true
}

export function disconnectBridgeSource(source: MessageEventSource): void {
  const unsubscribe = subscriptions.get(source as object)
  unsubscribe?.()
  subscriptions.delete(source as object)
}

function escapeHtml(value: string): string {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;')
}

export function createCounterBundle(title: string, nonce = randomToken(18)): { html: string; nonce: string } {
  const safeTitle = escapeHtml(title)
  const csp = [
    "default-src 'none'",
    `script-src 'nonce-${nonce}'`,
    `style-src 'nonce-${nonce}'`,
    "connect-src 'none'",
    "img-src 'none'",
    "font-src 'none'",
    "object-src 'none'",
    "base-uri 'none'",
    "form-action 'none'",
  ].join('; ')
  return {
    nonce,
    html: `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta http-equiv="Content-Security-Policy" content="${csp}">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>${safeTitle}</title>
  <style nonce="${nonce}">
    :root { color-scheme: dark; font-family: Georgia, serif; background: #141713; color: #f0ecdf; }
    * { box-sizing: border-box; }
    body { margin: 0; min-height: 100vh; display: grid; place-items: center; padding: 24px; }
    main { width: min(420px, 100%); padding: 42px; border: 1px solid #4c5248; background: #1b1f1a; box-shadow: 12px 12px 0 #090b09; }
    .eyebrow { color: #a4c48c; font: 700 11px/1 monospace; letter-spacing: .18em; text-transform: uppercase; }
    h1 { margin: 12px 0 30px; font-size: clamp(24px, 8vw, 42px); font-weight: 400; line-height: 1; }
    output { display: block; color: #f2b65d; font: 700 86px/1 monospace; letter-spacing: -.09em; margin-bottom: 24px; }
    button { width: 100%; border: 0; padding: 15px 20px; background: #e7e1d2; color: #161915; font: 700 14px/1 monospace; cursor: pointer; }
    button:active { transform: translate(2px, 2px); }
    .status { min-height: 18px; margin-top: 14px; color: #858c80; font: 11px/1.4 monospace; }
  </style>
</head>
<body>
  <main>
    <div class="eyebrow">Shared instrument</div>
    <h1>${safeTitle}</h1>
    <output id="count">—</output>
    <button id="increment" type="button">Add one</button>
    <div id="status" class="status">Connecting to the room bridge…</div>
  </main>
  <script nonce="${nonce}">
    (() => {
      const NONCE = ${JSON.stringify(nonce)};
      const pending = new Map();
      let sequence = 0;
      const count = document.getElementById('count');
      const status = document.getElementById('status');
      function call(method, payload) {
        const requestId = 'request-' + (++sequence);
        parent.postMessage({ channel: 'meta-room', version: 1, nonce: NONCE, type: 'bridge.request', requestId, method, payload }, '*');
        return new Promise((resolve, reject) => {
          pending.set(requestId, { resolve, reject });
          setTimeout(() => { if (pending.delete(requestId)) reject(new Error('Bridge timeout')); }, 5000);
        });
      }
      function render(state) {
        if (!state || !Number.isSafeInteger(state.value) || state.value < 0) return;
        count.textContent = String(state.value);
        status.textContent = 'Saved locally · ready to sync';
      }
      window.addEventListener('message', (event) => {
        if (event.source !== parent) return;
        const message = event.data;
        if (!message || message.channel !== 'meta-room' || message.version !== 1 || message.nonce !== NONCE) return;
        if (message.type === 'bridge.state') render(message.state);
        if (message.type === 'bridge.response' && pending.has(message.requestId)) {
          const request = pending.get(message.requestId);
          pending.delete(message.requestId);
          if (message.ok) { render(message.state); request.resolve(message.state); }
          else request.reject(new Error(message.error || 'Bridge update failed'));
        }
      });
      document.getElementById('increment').addEventListener('click', async () => {
        status.textContent = 'Saving…';
        try { await call('update', { type: 'counter.increment' }); }
        catch { status.textContent = 'Update failed · try again'; }
      });
      call('subscribe').catch(() => { status.textContent = 'Room bridge unavailable'; });
    })();
  </script>
</body>
</html>`,
  }
}
