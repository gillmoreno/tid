import { decryptJson, encryptJson, randomToken } from './crypto'
import type { CounterOperation, CounterState, DeviceIdentity, OutboxItem, VaultRoom } from './types'

const DATABASE_NAME = 'meta-room-vault'
const DATABASE_VERSION = 1
const ROOMS_STORE = 'rooms'
const SETTINGS_STORE = 'settings'

let databasePromise: Promise<IDBDatabase> | undefined
const roomQueues = new Map<string, Promise<void>>()

async function withRoomLock<T>(roomId: string, operation: () => Promise<T>): Promise<T> {
  const previous = roomQueues.get(roomId) ?? Promise.resolve()
  let release: () => void = () => undefined
  const gate = new Promise<void>((resolve) => { release = resolve })
  const tail = previous.then(() => gate)
  roomQueues.set(roomId, tail)
  await previous
  try {
    return await operation()
  } finally {
    release()
    if (roomQueues.get(roomId) === tail) roomQueues.delete(roomId)
  }
}

function requestResult<T>(request: IDBRequest<T>): Promise<T> {
  return new Promise((resolve, reject) => {
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error ?? new Error('IndexedDB request failed'))
  })
}

function transactionDone(transaction: IDBTransaction): Promise<void> {
  return new Promise((resolve, reject) => {
    transaction.oncomplete = () => resolve()
    transaction.onerror = () => reject(transaction.error ?? new Error('IndexedDB transaction failed'))
    transaction.onabort = () => reject(transaction.error ?? new Error('IndexedDB transaction aborted'))
  })
}

function openVault(): Promise<IDBDatabase> {
  databasePromise ??= new Promise((resolve, reject) => {
    const request = indexedDB.open(DATABASE_NAME, DATABASE_VERSION)
    request.onupgradeneeded = () => {
      const database = request.result
      if (!database.objectStoreNames.contains(ROOMS_STORE)) {
        database.createObjectStore(ROOMS_STORE, { keyPath: 'id' })
      }
      if (!database.objectStoreNames.contains(SETTINGS_STORE)) {
        database.createObjectStore(SETTINGS_STORE, { keyPath: 'key' })
      }
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error ?? new Error('Unable to open room vault'))
  })
  return databasePromise
}

export async function putRoom(room: VaultRoom): Promise<void> {
  const database = await openVault()
  const transaction = database.transaction(ROOMS_STORE, 'readwrite')
  transaction.objectStore(ROOMS_STORE).put(room)
  await transactionDone(transaction)
}

export async function getRoom(roomId: string): Promise<VaultRoom | undefined> {
  const database = await openVault()
  const transaction = database.transaction(ROOMS_STORE, 'readonly')
  return requestResult(transaction.objectStore(ROOMS_STORE).get(roomId)) as Promise<VaultRoom | undefined>
}

export async function listRooms(): Promise<VaultRoom[]> {
  const database = await openVault()
  const transaction = database.transaction(ROOMS_STORE, 'readonly')
  const rooms = await requestResult(transaction.objectStore(ROOMS_STORE).getAll()) as VaultRoom[]
  return rooms.sort((left, right) => right.createdAt - left.createdAt)
}

export async function getDeviceIdentity(): Promise<DeviceIdentity> {
  const database = await openVault()
  const readTransaction = database.transaction(SETTINGS_STORE, 'readonly')
  const existing = await requestResult(readTransaction.objectStore(SETTINGS_STORE).get('device')) as
    | { key: 'device'; value: DeviceIdentity }
    | undefined
  if (existing?.value.identity) return existing.value

  const identity: DeviceIdentity = {
    id: existing?.value.id ?? `device_${randomToken(18)}`,
    identity: randomToken(32),
    label: existing?.value.label ?? navigator.userAgent.slice(0, 120),
  }
  const writeTransaction = database.transaction(SETTINGS_STORE, 'readwrite')
  writeTransaction.objectStore(SETTINGS_STORE).put({ key: 'device', value: identity })
  await transactionDone(writeTransaction)
  return identity
}

export async function persistIncrement(roomId: string, operation: CounterOperation): Promise<{
  room: VaultRoom
  state: CounterState
}> {
  return withRoomLock(roomId, async () => {
    const room = await getRoom(roomId)
    if (!room) throw new Error('Room is not available in this vault')
    const state = await decryptJson<CounterState>(room.roomDataKey, room.encryptedState, `room:${room.id}:state`)
    if (room.appliedOperationIds.includes(operation.id)) return { room, state }

    const nextState = { value: state.value + operation.delta }
    const encryptedState = await encryptJson(room.roomDataKey, nextState, `room:${room.id}:state`)
    const payload = await encryptJson(room.roomDataKey, operation, `room:${room.id}:operation`)
    const outboxItem: OutboxItem = {
      id: operation.id,
      kind: 'operation',
      payload,
      createdAt: operation.createdAt,
      attempts: 0,
    }
    const updated: VaultRoom = {
      ...room,
      encryptedState,
      appliedOperationIds: [...room.appliedOperationIds, operation.id],
      outbox: [...room.outbox, outboxItem],
    }
    await putRoom(updated)
    return { room: updated, state: nextState }
  })
}

export async function applyRemoteOperation(roomId: string, operation: CounterOperation): Promise<CounterState> {
  return withRoomLock(roomId, async () => {
    const room = await getRoom(roomId)
    if (!room) throw new Error('Room is not available in this vault')
    const state = await decryptJson<CounterState>(room.roomDataKey, room.encryptedState, `room:${room.id}:state`)
    if (room.appliedOperationIds.includes(operation.id)) return state

    const nextState = { value: state.value + operation.delta }
    await putRoom({
      ...room,
      encryptedState: await encryptJson(room.roomDataKey, nextState, `room:${room.id}:state`),
      appliedOperationIds: [...room.appliedOperationIds, operation.id],
    })
    return nextState
  })
}

export async function removeOutboxItem(roomId: string, operationId: string): Promise<void> {
  await withRoomLock(roomId, async () => {
    const room = await getRoom(roomId)
    if (!room) return
    await putRoom({ ...room, outbox: room.outbox.filter((item) => item.id !== operationId) })
  })
}

export async function updateSyncCursors(
  roomId: string,
  operationCursor?: number,
  signal?: { sessionId: string; cursor: number },
): Promise<VaultRoom | undefined> {
  return withRoomLock(roomId, async () => {
    const room = await getRoom(roomId)
    if (!room) return undefined
    const updated: VaultRoom = {
      ...room,
      operationCursor: operationCursor ?? room.operationCursor,
      signalCursors: signal
        ? { ...room.signalCursors, [signal.sessionId]: signal.cursor }
        : room.signalCursors,
    }
    await putRoom(updated)
    return updated
  })
}

export async function readRoomState(room: VaultRoom): Promise<CounterState> {
  return decryptJson(room.roomDataKey, room.encryptedState, `room:${room.id}:state`)
}

export async function clearVaultForTests(): Promise<void> {
  const database = await databasePromise
  database?.close()
  databasePromise = undefined
  roomQueues.clear()
  await new Promise<void>((resolve, reject) => {
    const request = indexedDB.deleteDatabase(DATABASE_NAME)
    request.onsuccess = () => resolve()
    request.onerror = () => reject(request.error ?? new Error('Unable to clear vault'))
    request.onblocked = () => resolve()
  })
}
