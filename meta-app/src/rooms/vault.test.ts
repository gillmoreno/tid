import { afterEach, describe, expect, it } from 'vitest'
import { encryptJson, generateRoomDataKey } from './crypto'
import type { CounterOperation, VaultRoom } from './types'
import {
  applyRemoteOperation,
  clearVaultForTests,
  getRoom,
  listRooms,
  persistIncrement,
  putRoom,
  readRoomState,
} from './vault'

async function roomFixture(id: string): Promise<VaultRoom> {
  const roomDataKey = await generateRoomDataKey()
  return {
    id,
    title: 'Test counter',
    capacity: 2,
    createdAt: Date.now(),
    role: 'owner',
    memberId: 'member-owner',
    deviceId: 'device-a',
    ownerDeviceId: 'device-a',
    memberCredential: 'member-credential',
    roomDataKey,
    encryptedBundle: await encryptJson(roomDataKey, { version: 1, kind: 'counter' }, `room:${id}:bundle`),
    encryptedState: await encryptJson(roomDataKey, { value: 0 }, `room:${id}:state`),
    appliedOperationIds: [],
    outbox: [],
    operationCursor: 0,
    signalCursors: {},
  }
}

function operation(id: string, deviceId: string): CounterOperation {
  return {
    id,
    kind: 'counter.increment',
    delta: 1,
    memberId: `member-${deviceId}`,
    deviceId,
    createdAt: 1,
  }
}

afterEach(clearVaultForTests)

describe('IndexedDB room vault', () => {
  it('restores encrypted room records after persistence', async () => {
    const room = await roomFixture('room-persisted')
    await putRoom(room)

    const restored = await getRoom(room.id)
    expect(restored?.memberCredential).toBe('member-credential')
    await expect(readRoomState(restored!)).resolves.toEqual({ value: 0 })
    expect(await listRooms()).toHaveLength(1)
  })

  it('persists a unique increment and its outbox item atomically', async () => {
    const room = await roomFixture('room-outbox')
    await putRoom(room)
    const result = await persistIncrement(room.id, operation('op-local', 'device-a'))

    expect(result.state.value).toBe(1)
    expect(result.room.appliedOperationIds).toEqual(['op-local'])
    expect(result.room.outbox.map((item) => item.id)).toEqual(['op-local'])
  })

  it('converges for duplicate operations arriving in different orders', async () => {
    const first = await roomFixture('room-first')
    const second = await roomFixture('room-second')
    await putRoom(first)
    await putRoom(second)
    const operationA = operation('op-a', 'device-a')
    const operationB = operation('op-b', 'device-b')

    await Promise.all([
      applyRemoteOperation(first.id, operationA),
      applyRemoteOperation(first.id, operationB),
      applyRemoteOperation(first.id, operationA),
    ])
    await applyRemoteOperation(second.id, operationB)
    await applyRemoteOperation(second.id, operationA)
    await applyRemoteOperation(second.id, operationB)

    await expect(readRoomState((await getRoom(first.id))!)).resolves.toEqual({ value: 2 })
    await expect(readRoomState((await getRoom(second.id))!)).resolves.toEqual({ value: 2 })
  })
})
