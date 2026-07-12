export type RoomRole = 'owner' | 'member'

export type ConnectionStatus =
  | 'offline'
  | 'mailbox'
  | 'connecting'
  | 'p2p'
  | 'unavailable'

export type EncryptedPayload = {
  algorithm: 'AES-GCM'
  iv: string
  ciphertext: string
  version: 1
}

export type CounterOperation = {
  id: string
  kind: 'counter.increment'
  delta: 1
  memberId: string
  deviceId: string
  createdAt: number
}

export type CounterState = {
  value: number
}

export type OutboxItem = {
  id: string
  kind: 'operation' | 'checkpoint'
  payload: EncryptedPayload
  createdAt: number
  attempts: number
}

export type VaultRoom = {
  id: string
  title: string
  capacity: number
  createdAt: number
  role: RoomRole
  memberId: string
  deviceId: string
  memberCredential: string
  ownerDeviceId: string
  ownerCapability?: string
  roomDataKey: CryptoKey
  encryptedBundle: EncryptedPayload
  encryptedState: EncryptedPayload
  appliedOperationIds: string[]
  outbox: OutboxItem[]
  operationCursor: number
  signalCursors: Record<string, number>
  inviteId?: string
  invitationPackage?: string
  inviteEnvelope?: EncryptedPayload
  shareUrl?: string
}

export type RoomSummary = Pick<
  VaultRoom,
  'id' | 'title' | 'capacity' | 'createdAt' | 'role' | 'inviteId' | 'shareUrl'
>

export type DeviceIdentity = {
  id: string
  identity: string
  label: string
}

export type Checkpoint = {
  title: string
  capacity: number
  bundle: EncryptedPayload
  state: EncryptedPayload
  appliedOperationIds: string[]
  createdAt: number
}

export type RoomDevice = {
  deviceId: string
  isOwner: boolean
  isSelf: boolean
}
