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

export function createCounterBundle(title: string, nonce = randomToken(18)): { src: string; nonce: string } {
  const fragment = new URLSearchParams({ nonce, title }).toString()
  return {
    nonce,
    src: `/room-frame.html?v=3#${fragment}`,
  }
}
