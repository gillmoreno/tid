import { describe, expect, it } from 'vitest'
import { deterministicSessionId } from './peer'

describe('peer session rendezvous', () => {
  it('derives the same URL-safe session for both device orderings', async () => {
    const first = await deterministicSessionId('room_12345678', 'device-alpha', 'device-beta')
    const second = await deterministicSessionId('room_12345678', 'device-beta', 'device-alpha')

    expect(first).toBe(second)
    expect(first).toMatch(/^session_[A-Za-z0-9_-]{40}$/)
    await expect(deterministicSessionId('room_12345678', 'device-alpha', 'device-gamma')).resolves.not.toBe(first)
  })
})
