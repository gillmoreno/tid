import { describe, expect, it } from 'vitest'
import {
  decryptJson,
  encryptJson,
  generateRoomDataKey,
  randomToken,
  unwrapRoomDataKey,
  wrapRoomDataKey,
} from './crypto'

describe('room crypto', () => {
  it('round trips encrypted state with authenticated context', async () => {
    const key = await generateRoomDataKey()
    const payload = await encryptJson(key, { value: 42 }, 'room:test:state')

    await expect(decryptJson(key, payload, 'room:test:state')).resolves.toEqual({ value: 42 })
    await expect(decryptJson(key, payload, 'room:other:state')).rejects.toBeDefined()
  })

  it('wraps the room data key with the separately shared invitation key', async () => {
    const roomKey = await generateRoomDataKey()
    const inviteKey = randomToken()
    const envelope = await wrapRoomDataKey(roomKey, inviteKey)
    const restored = await unwrapRoomDataKey(envelope, inviteKey)
    const payload = await encryptJson(roomKey, 'secret', 'test')

    await expect(decryptJson(restored, payload, 'test')).resolves.toBe('secret')
    await expect(unwrapRoomDataKey(envelope, 'wrong-key')).rejects.toBeDefined()
  })
})
