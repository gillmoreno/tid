import { describe, expect, it, vi } from 'vitest'
import { handleBridgeMessage, type RoomBridge } from './bridge'

function eventWithSource(data: unknown, source: MessageEventSource): MessageEvent<unknown> {
  const event = new MessageEvent('message', { data })
  Object.defineProperty(event, 'source', { value: source })
  return event
}

describe('sandbox room bridge', () => {
  it('accepts only the exact iframe source and per-load nonce', async () => {
    const expectedSource = { postMessage: vi.fn() } as unknown as MessageEventSource
    const otherSource = { postMessage: vi.fn() } as unknown as MessageEventSource
    const bridge: RoomBridge = {
      getState: vi.fn(async () => ({ value: 3 })),
      update: vi.fn(async () => ({ value: 4 })),
      subscribe: vi.fn(() => () => undefined),
    }
    const valid = {
      channel: 'meta-room',
      version: 1,
      nonce: 'load-nonce',
      type: 'bridge.request',
      requestId: 'request-1',
      method: 'getState',
    }

    await expect(handleBridgeMessage(eventWithSource(valid, otherSource), expectedSource, 'load-nonce', bridge)).resolves.toBe(false)
    await expect(handleBridgeMessage(eventWithSource({ ...valid, nonce: 'old-nonce' }, expectedSource), expectedSource, 'load-nonce', bridge)).resolves.toBe(false)
    await expect(handleBridgeMessage(eventWithSource(valid, expectedSource), expectedSource, 'load-nonce', bridge)).resolves.toBe(true)

    expect(bridge.getState).toHaveBeenCalledTimes(1)
    expect(expectedSource.postMessage).toHaveBeenCalledWith(expect.objectContaining({
      type: 'bridge.response',
      requestId: 'request-1',
      state: { value: 3 },
    }), '*')
  })

  it('rejects unexpected update payload fields', async () => {
    const source = { postMessage: vi.fn() } as unknown as MessageEventSource
    const bridge: RoomBridge = {
      getState: vi.fn(async () => ({ value: 0 })),
      update: vi.fn(async () => ({ value: 1 })),
      subscribe: vi.fn(() => () => undefined),
    }
    const event = eventWithSource({
      channel: 'meta-room',
      version: 1,
      nonce: 'nonce',
      type: 'bridge.request',
      requestId: 'request-2',
      method: 'update',
      payload: { type: 'counter.increment', amount: 100 },
    }, source)

    await expect(handleBridgeMessage(event, source, 'nonce', bridge)).resolves.toBe(false)
    expect(bridge.update).not.toHaveBeenCalled()
  })
})
