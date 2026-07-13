export const SIGNALING_URL = import.meta.env.VITE_SIGNALING_URL || 'http://localhost:8081'

export function resolveSignalingEndpoint(signalingUrl: string, origin: string): URL {
  return new URL(signalingUrl, origin)
}

export class RoomApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(message: string, status: number, code = 'request_failed') {
    super(message)
    this.name = 'RoomApiError'
    this.status = status
    this.code = code
  }
}

type RequestOptions = {
  method?: 'GET' | 'POST' | 'PUT'
  credential?: string
  ownerCapability?: string
  creatorPermit?: string
  body?: unknown
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  let response: Response
  try {
    response = await fetch(`${SIGNALING_URL}${path}`, {
      method: options.method ?? 'GET',
      headers: {
        Accept: 'application/json',
        ...(options.body === undefined ? {} : { 'Content-Type': 'application/json' }),
        ...(options.credential ? { Authorization: `Bearer ${options.credential}` } : {}),
        ...(options.ownerCapability ? { 'X-Owner-Capability': options.ownerCapability } : {}),
        ...(options.creatorPermit ? { 'X-Room-Creator-Permit': options.creatorPermit } : {}),
      },
      body: options.body === undefined ? undefined : JSON.stringify(options.body),
    })
  } catch {
    throw new RoomApiError('The room service is unreachable. Your local rooms remain available offline.', 0, 'offline')
  }

  if (response.status === 204) return undefined as T
  const data = await response.json().catch(() => ({})) as Record<string, unknown>
  if (!response.ok) {
    const nested = data.error && typeof data.error === 'object'
      ? data.error as Record<string, unknown>
      : data
    const code = typeof nested.code === 'string' ? nested.code : 'request_failed'
    const fallback = {
      invite_expired: 'This invitation has expired.',
      invite_revoked: 'This invitation was revoked.',
      room_full: 'This room has reached its member capacity.',
      invalid_invite_secret: 'The invitation package is invalid.',
      invite_already_redeemed: 'This invitation was already redeemed by another device.',
      invalid_credential: 'The room credential was rejected.',
      checkpoint_not_found: 'The room checkpoint is not available yet.',
    }[code] ?? `The room service rejected the request (${response.status}).`
    throw new RoomApiError(
      typeof nested.message === 'string' ? nested.message : fallback,
      response.status,
      code,
    )
  }
  return data as T
}

export type CreateRoomResponse = {
  roomId: string
  maxMembers: number
  ownerCapability: string
  ownerMemberId: string
  ownerDeviceId: string
  ownerMemberCredential: string
}

export type CreateInviteResponse = {
  inviteId: string
  inviteSecret: string
  expiresAt: string
}

export type RedeemInviteRequest = {
  inviteSecret: string
  deviceId: string
  deviceIdentity: string
  memberCredential: string
  idempotencyKey: string
}

export type RedeemInviteResponse = {
  roomId: string
  memberId: string
  deviceId: string
  reconnected: boolean
}

export type RoomInfoResponse = {
  roomId: string
  maxMembers: number
  memberCount: number
  createdAt?: string
}

export type DeviceResponse = {
  devices: Array<{ deviceId: string; isOwner: boolean; isSelf: boolean }>
}

export type CheckpointResponse = {
  roomId: string
  uploaderDeviceId: string
  envelope: string
  updatedAt: string
  expiresAt: string
}

export type OperationResponse = {
  operations: Array<{
    operationId: number
    fromDeviceId: string
    toDeviceId: string
    envelope: string
    createdAt: string
    expiresAt: string
  }>
}

export type SignalKind = 'offer' | 'answer' | 'candidate'

export type SignalResponse = {
  signals: Array<{
    signalId: number
    sessionId: string
    kind: SignalKind
    fromDeviceId: string
    toDeviceId: string
    envelope: string
    createdAt: string
    expiresAt: string
  }>
}

export const roomApi = {
  createRoom(maxMembers: number, creatorPermit: string): Promise<CreateRoomResponse> {
    return request('/v2/rooms', {
      method: 'POST',
      creatorPermit,
      body: { maxMembers },
    })
  },

  createInvite(
    roomId: string,
    ownerCapability: string,
    expiresInSeconds = 86_400,
  ): Promise<CreateInviteResponse> {
    return request(`/v2/rooms/${encodeURIComponent(roomId)}/invites`, {
      method: 'POST',
      ownerCapability,
      body: { expiresInSeconds },
    })
  },

  redeemInvite(inviteId: string, input: RedeemInviteRequest): Promise<RedeemInviteResponse> {
    return request(`/v2/invites/${encodeURIComponent(inviteId)}/redeem`, {
      method: 'POST',
      body: input,
    })
  },

  getRoom(roomId: string, credential: string): Promise<RoomInfoResponse> {
    return request(`/v2/rooms/${encodeURIComponent(roomId)}`, { credential })
  },

  getDevices(roomId: string, credential: string): Promise<DeviceResponse> {
    return request(`/v2/rooms/${encodeURIComponent(roomId)}/devices`, { credential })
  },

  putCheckpoint(
    roomId: string,
    credential: string,
    envelope: string,
    expiresInSeconds = 604_800,
  ): Promise<void> {
    return request(`/v2/rooms/${encodeURIComponent(roomId)}/mailbox/checkpoint`, {
      method: 'PUT',
      credential,
      body: { envelope, expiresInSeconds },
    })
  },

  getCheckpoint(roomId: string, credential: string): Promise<CheckpointResponse> {
    return request(`/v2/rooms/${encodeURIComponent(roomId)}/mailbox/checkpoint`, { credential })
  },

  postOperation(
    roomId: string,
    recipientDeviceId: string,
    credential: string,
    envelope: string,
    expiresInSeconds = 604_800,
  ): Promise<{ operationId: number }> {
    return request(
      `/v2/rooms/${encodeURIComponent(roomId)}/mailbox/${encodeURIComponent(recipientDeviceId)}/operations`,
      { method: 'POST', credential, body: { envelope, expiresInSeconds } },
    )
  },

  getOperations(
    roomId: string,
    deviceId: string,
    credential: string,
    after: number,
  ): Promise<OperationResponse> {
    return request(
      `/v2/rooms/${encodeURIComponent(roomId)}/mailbox/${encodeURIComponent(deviceId)}/operations?after=${after}`,
      { credential },
    )
  },

  postSignal(
    roomId: string,
    sessionId: string,
    credential: string,
    input: {
      kind: SignalKind
      fromDeviceId: string
      toDeviceId: string
      envelope: string
      expiresInSeconds?: number
    },
  ): Promise<{ signalId: number }> {
    return request(
      `/v2/rooms/${encodeURIComponent(roomId)}/sessions/${encodeURIComponent(sessionId)}/signals`,
      {
        method: 'POST',
        credential,
        body: { ...input, expiresInSeconds: input.expiresInSeconds ?? 600 },
      },
    )
  },

  getSignals(
    roomId: string,
    sessionId: string,
    credential: string,
    after: number,
  ): Promise<SignalResponse> {
    return request(
      `/v2/rooms/${encodeURIComponent(roomId)}/sessions/${encodeURIComponent(sessionId)}/signals?after=${after}`,
      { credential },
    )
  },
}
