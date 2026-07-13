import { roomApi } from './api'
import { createCounterBundle } from './bridge'
import {
  decryptJson,
  deriveStableToken,
  encryptJson,
  generateRoomDataKey,
  randomToken,
  unwrapRoomDataKey,
  wrapRoomDataKey,
} from './crypto'
import { encodeInvitationPackage, parseInvitationPackage } from './invitation'
import { RoomPeer } from './peer'
import {
  applyRemoteOperation,
  getDeviceIdentity,
  getRoom,
  persistIncrement,
  putRoom,
  readRoomState,
  removeOutboxItem,
  updateSyncCursors,
} from './vault'
import type {
  Checkpoint,
  ConnectionStatus,
  CounterOperation,
  CounterState,
  EncryptedPayload,
  VaultRoom,
} from './types'

export type CodeBundle = {
  version: 1
  kind: 'counter'
  title: string
}

export type CreatedRoom = {
  room: VaultRoom
  invitationUrl: string
  invitationKey: string
  checkpointQueued: boolean
}

const stateListeners = new Map<string, Set<(state: CounterState) => void>>()
const syncListeners = new Map<string, Set<() => void>>()

function notify(roomId: string, state: CounterState): void {
  for (const listener of stateListeners.get(roomId) ?? []) listener(state)
  for (const listener of syncListeners.get(roomId) ?? []) listener()
}

function checkpointContext(roomId: string): string {
  return `room:${roomId}:checkpoint`
}

function parseEncryptedEnvelope(envelope: string): EncryptedPayload {
  let value: unknown
  try {
    value = JSON.parse(envelope)
  } catch {
    throw new Error('Invalid encrypted envelope')
  }
  if (!value || typeof value !== 'object') throw new Error('Invalid encrypted envelope')
  const payload = value as Record<string, unknown>
  if (
    payload.algorithm !== 'AES-GCM'
    || payload.version !== 1
    || typeof payload.iv !== 'string'
    || typeof payload.ciphertext !== 'string'
  ) throw new Error('Invalid encrypted envelope')
  return value as EncryptedPayload
}

async function encryptCheckpoint(room: VaultRoom): Promise<string> {
  const current = (await getRoom(room.id)) ?? room
  const checkpoint: Checkpoint = {
    title: current.title,
    capacity: current.capacity,
    bundle: current.encryptedBundle,
    state: current.encryptedState,
    appliedOperationIds: current.appliedOperationIds,
    createdAt: Date.now(),
  }
  return JSON.stringify(await encryptJson(
    current.roomDataKey,
    checkpoint,
    checkpointContext(current.id),
  ))
}

export async function createRoom(
  title: string,
  capacity: number,
  creatorPermit: string,
): Promise<CreatedRoom> {
  const response = await roomApi.createRoom(capacity, creatorPermit)
  const roomDataKey = await generateRoomDataKey()
  const bundle: CodeBundle = { version: 1, kind: 'counter', title }
  const encryptedBundle = await encryptJson(roomDataKey, bundle, `room:${response.roomId}:bundle`)
  const encryptedState = await encryptJson(roomDataKey, { value: 0 }, `room:${response.roomId}:state`)
  const invitation = await roomApi.createInvite(response.roomId, response.ownerCapability)
  const keyEnvelope = await wrapRoomDataKey(roomDataKey, invitation.inviteSecret)
  const invitationPackage = encodeInvitationPackage({
    version: 1,
    inviteId: invitation.inviteId,
    inviteSecret: invitation.inviteSecret,
    roomId: response.roomId,
    ownerDeviceId: response.ownerDeviceId,
    keyEnvelope,
  })
  const invitationUrl = `${window.location.origin}/join/${encodeURIComponent(invitation.inviteId)}`
  const checkpointOutboxId = `checkpoint_${randomToken(18)}`
  let room: VaultRoom = {
    id: response.roomId,
    title,
    capacity: response.maxMembers,
    createdAt: Date.now(),
    role: 'owner',
    memberId: response.ownerMemberId,
    deviceId: response.ownerDeviceId,
    ownerDeviceId: response.ownerDeviceId,
    memberCredential: response.ownerMemberCredential,
    ownerCapability: response.ownerCapability,
    roomDataKey,
    encryptedBundle,
    encryptedState,
    appliedOperationIds: [],
    outbox: [{
      id: checkpointOutboxId,
      kind: 'checkpoint',
      payload: await encryptJson(roomDataKey, { queued: true }, `room:${response.roomId}:outbox`),
      createdAt: Date.now(),
      attempts: 0,
    }],
    operationCursor: 0,
    signalCursors: {},
    inviteId: invitation.inviteId,
    invitationPackage,
    inviteEnvelope: keyEnvelope,
    shareUrl: invitationUrl,
  }
  await putRoom(room)

  let checkpointQueued = false
  try {
    await roomApi.putCheckpoint(room.id, room.memberCredential, await encryptCheckpoint(room))
    await removeOutboxItem(room.id, checkpointOutboxId)
    room = (await getRoom(room.id)) ?? room
  } catch {
    checkpointQueued = true
  }
  return { room, invitationUrl, invitationKey: invitationPackage, checkpointQueued }
}

export async function redeemInvitation(inviteId: string, packageValue: string): Promise<VaultRoom> {
  const invitation = parseInvitationPackage(packageValue, inviteId)
  const device = await getDeviceIdentity()
  const memberCredential = await deriveStableToken(device.identity, `member:${inviteId}`)
  const idempotencyKey = await deriveStableToken(device.identity, `redeem:${inviteId}`)
  const response = await roomApi.redeemInvite(inviteId, {
    inviteSecret: invitation.inviteSecret,
    deviceId: device.id,
    deviceIdentity: device.identity,
    memberCredential,
    idempotencyKey,
  })
  if (response.roomId !== invitation.roomId || response.deviceId !== device.id) {
    throw new Error('Invitation redemption did not match its package')
  }

  const [roomInfo, checkpointResponse, deviceResponse] = await Promise.all([
    roomApi.getRoom(response.roomId, memberCredential),
    roomApi.getCheckpoint(response.roomId, memberCredential),
    roomApi.getDevices(response.roomId, memberCredential),
  ])
  if (roomInfo.roomId !== invitation.roomId || checkpointResponse.roomId !== invitation.roomId) {
    throw new Error('Room checkpoint does not match its invitation')
  }
  const owner = deviceResponse.devices.find((device) => device.isOwner)
  if (!owner || owner.deviceId !== invitation.ownerDeviceId) {
    throw new Error('Room owner does not match its invitation package')
  }
  const roomDataKey = await unwrapRoomDataKey(invitation.keyEnvelope, invitation.inviteSecret)
  const checkpoint = await decryptJson<Checkpoint>(
    roomDataKey,
    parseEncryptedEnvelope(checkpointResponse.envelope),
    checkpointContext(invitation.roomId),
  )
  const bundle = await decryptJson<CodeBundle>(
    roomDataKey,
    checkpoint.bundle,
    `room:${invitation.roomId}:bundle`,
  )
  await decryptJson<CounterState>(
    roomDataKey,
    checkpoint.state,
    `room:${invitation.roomId}:state`,
  )
  if (
    bundle.version !== 1
    || bundle.kind !== 'counter'
    || typeof bundle.title !== 'string'
    || checkpoint.title !== bundle.title
    || checkpoint.capacity !== roomInfo.maxMembers
    || !Array.isArray(checkpoint.appliedOperationIds)
  ) throw new Error('Room checkpoint metadata is invalid')

  const room: VaultRoom = {
    id: response.roomId,
    title: checkpoint.title,
    capacity: roomInfo.maxMembers,
    createdAt: checkpoint.createdAt,
    role: 'member',
    memberId: response.memberId,
    deviceId: response.deviceId,
    ownerDeviceId: invitation.ownerDeviceId,
    memberCredential,
    roomDataKey,
    encryptedBundle: checkpoint.bundle,
    encryptedState: checkpoint.state,
    appliedOperationIds: checkpoint.appliedOperationIds,
    outbox: [],
    operationCursor: 0,
    signalCursors: {},
    inviteId,
  }
  await putRoom(room)
  return room
}

export async function loadBundle(room: VaultRoom): Promise<CodeBundle> {
  return decryptJson(room.roomDataKey, room.encryptedBundle, `room:${room.id}:bundle`)
}

export async function incrementCounter(room: VaultRoom): Promise<CounterState> {
  const operation: CounterOperation = {
    id: `op_${randomToken(18)}`,
    kind: 'counter.increment',
    delta: 1,
    memberId: room.memberId,
    deviceId: room.deviceId,
    createdAt: Date.now(),
  }
  const result = await persistIncrement(room.id, operation)
  notify(room.id, result.state)
  return result.state
}

export async function applyEncryptedOperation(room: VaultRoom, payload: EncryptedPayload): Promise<CounterState> {
  const operation = await decryptJson<CounterOperation>(
    room.roomDataKey,
    payload,
    `room:${room.id}:operation`,
  )
  if (
    operation.kind !== 'counter.increment'
    || operation.delta !== 1
    || typeof operation.id !== 'string'
    || operation.id.length > 128
    || typeof operation.memberId !== 'string'
    || typeof operation.deviceId !== 'string'
    || !Number.isSafeInteger(operation.createdAt)
  ) throw new Error('Invalid counter operation')
  const state = await applyRemoteOperation(room.id, operation)
  notify(room.id, state)
  return state
}

export function subscribeToRoom(roomId: string, listener: (state: CounterState) => void): () => void {
  const listeners = stateListeners.get(roomId) ?? new Set()
  listeners.add(listener)
  stateListeners.set(roomId, listeners)
  return () => {
    listeners.delete(listener)
    if (listeners.size === 0) stateListeners.delete(roomId)
  }
}

export async function roomState(room: VaultRoom): Promise<CounterState> {
  const current = await getRoom(room.id)
  if (!current) throw new Error('Room is unavailable')
  return readRoomState(current)
}

export class RoomSync {
  private peer?: RoomPeer
  private peerDeviceId?: string
  private peerStatus: ConnectionStatus = 'mailbox'
  private lastPeerAttempt = 0
  private mailboxTimer?: number
  private stopped = false
  private syncing = false
  private room: VaultRoom
  private readonly onStatus: (status: ConnectionStatus) => void

  constructor(room: VaultRoom, onStatus: (status: ConnectionStatus) => void) {
    this.room = room
    this.onStatus = onStatus
  }

  start(): void {
    const listeners = syncListeners.get(this.room.id) ?? new Set()
    listeners.add(this.wake)
    syncListeners.set(this.room.id, listeners)
    this.mailboxTimer = window.setInterval(this.wake, 4_000)
    this.wake()
  }

  stop(): void {
    this.stopped = true
    if (this.mailboxTimer) window.clearInterval(this.mailboxTimer)
    const listeners = syncListeners.get(this.room.id)
    listeners?.delete(this.wake)
    if (listeners?.size === 0) syncListeners.delete(this.room.id)
    this.peer?.close()
  }

  private readonly wake = () => {
    if (!this.syncing) void this.syncMailbox()
  }

  private async discoverPeer(otherDeviceIds: string[]): Promise<void> {
    const peerDeviceId = [...otherDeviceIds].sort()[0]
    if (!peerDeviceId) {
      this.peer?.close()
      this.peer = undefined
      this.peerDeviceId = undefined
      return
    }
    const shouldRetry = this.peerStatus !== 'p2p' && Date.now() - this.lastPeerAttempt > 10_000
    if (this.peer && this.peerDeviceId === peerDeviceId && !shouldRetry) return
    this.peer?.close()
    this.peerDeviceId = peerDeviceId
    this.lastPeerAttempt = Date.now()
    this.peer = new RoomPeer(this.room, peerDeviceId, {
      onPayload: (payload) => {
        void applyEncryptedOperation(this.room, payload).catch(() => undefined)
      },
      onStatus: (status) => {
        this.peerStatus = status
        this.onStatus(status)
      },
      getSignalCursor: (sessionId) => this.room.signalCursors?.[sessionId] ?? 0,
      onSignalCursor: (sessionId, cursor) => {
        void updateSyncCursors(this.room.id, undefined, { sessionId, cursor })
      },
    })
    void this.peer.connect()
  }

  async syncMailbox(): Promise<void> {
    this.syncing = true
    try {
      if (this.stopped || !navigator.onLine) {
        this.onStatus('offline')
        return
      }
      let current = await getRoom(this.room.id)
      if (!current) return
      this.room = current
      const { devices } = await roomApi.getDevices(current.id, current.memberCredential)
      const recipients = devices
        .filter((device) => !device.isSelf)
        .map((device) => device.deviceId)
      await this.discoverPeer(recipients)

      let stateChanged = false
      const processedOutboxIds: string[] = []
      for (const item of current.outbox) {
        if (item.kind === 'operation') {
          this.peer?.send(item.payload)
          const envelope = JSON.stringify(item.payload)
          for (const recipientDeviceId of recipients) {
            await roomApi.postOperation(
              current.id,
              recipientDeviceId,
              current.memberCredential,
              envelope,
            )
          }
          stateChanged = true
        } else {
          stateChanged = true
        }
        processedOutboxIds.push(item.id)
      }

      const operations = await roomApi.getOperations(
        current.id,
        current.deviceId,
        current.memberCredential,
        current.operationCursor ?? 0,
      )
      let cursor = current.operationCursor ?? 0
      for (const operation of operations.operations) {
        await applyEncryptedOperation(current, parseEncryptedEnvelope(operation.envelope))
        cursor = Math.max(cursor, operation.operationId)
        stateChanged = true
      }
      if (stateChanged) {
        await roomApi.putCheckpoint(current.id, current.memberCredential, await encryptCheckpoint(current))
        for (const itemId of processedOutboxIds) await removeOutboxItem(current.id, itemId)
        if (cursor !== (current.operationCursor ?? 0)) {
          current = (await updateSyncCursors(current.id, cursor)) ?? current
          this.room = current
        }
      }
      if (this.peerStatus !== 'p2p') this.onStatus('mailbox')
    } catch {
      this.onStatus(navigator.onLine ? 'mailbox' : 'offline')
    } finally {
      this.syncing = false
    }
  }
}

export function renderCounterBundle(bundle: CodeBundle): { src: string; nonce: string } {
  if (bundle.version !== 1 || bundle.kind !== 'counter') throw new Error('Unsupported room bundle')
  return createCounterBundle(bundle.title)
}
