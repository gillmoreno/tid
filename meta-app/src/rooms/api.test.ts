import { afterEach, describe, expect, it, vi } from 'vitest'
import { roomApi, SIGNALING_URL } from './api'

function jsonResponse(value: unknown, status = 200): Response {
  return new Response(JSON.stringify(value), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

function requestDetails(mock: ReturnType<typeof vi.fn>, index: number) {
  const [url, init] = mock.mock.calls[index] as [string, RequestInit]
  return {
    url,
    method: init.method,
    headers: init.headers as Record<string, string>,
    body: init.body ? JSON.parse(String(init.body)) : undefined,
  }
}

afterEach(() => vi.unstubAllGlobals())

describe('signaling v2 API contract', () => {
  it('creates rooms and owner-authenticated invitations with canonical payloads', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(jsonResponse({
        roomId: 'room-1',
        maxMembers: 3,
        ownerCapability: 'owner-capability',
        ownerMemberId: 'member-owner',
        ownerDeviceId: 'device-owner',
        ownerMemberCredential: 'owner-member-credential',
      }, 201))
      .mockResolvedValueOnce(jsonResponse({
        inviteId: 'invite-1',
        inviteSecret: 'invite-secret-value',
        expiresAt: '2026-07-13T00:00:00Z',
      }, 201))
    vi.stubGlobal('fetch', fetchMock)

    await roomApi.createRoom(3, 'room-creator-permit')
    await roomApi.createInvite('room-1', 'owner-capability', 86_400)

    expect(requestDetails(fetchMock, 0)).toEqual({
      url: `${SIGNALING_URL}/v2/rooms`,
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
        'X-Room-Creator-Permit': 'room-creator-permit',
      },
      body: { maxMembers: 3 },
    })
    expect(requestDetails(fetchMock, 1)).toEqual({
      url: `${SIGNALING_URL}/v2/rooms/room-1/invites`,
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
        'X-Owner-Capability': 'owner-capability',
      },
      body: { expiresInSeconds: 86_400 },
    })
  })

  it.each([
    ['creator_permit_required', 401, 'room creator permit is required'],
    ['invalid_creator_permit', 403, 'room creator permit is invalid'],
  ])('surfaces creator gate failure %s', async (code, status, message) => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse({
      error: { code, message },
    }, status)))

    await expect(roomApi.createRoom(2, code === 'creator_permit_required' ? '' : 'wrong-permit'))
      .rejects.toMatchObject({ status, code, message })
  })

  it('redeems with client retry credentials and uses bearer room reads', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(jsonResponse({
        roomId: 'room-1',
        memberId: 'member-2',
        deviceId: 'device-2',
        reconnected: false,
      }, 201))
      .mockResolvedValueOnce(jsonResponse({ roomId: 'room-1', maxMembers: 2, memberCount: 2 }))
      .mockResolvedValueOnce(jsonResponse({ devices: [{ deviceId: 'device-2', isOwner: false, isSelf: true }] }))
    vi.stubGlobal('fetch', fetchMock)
    const redemption = {
      inviteSecret: 'invite-secret-value',
      deviceId: 'device-2',
      deviceIdentity: 'stable-device-identity',
      memberCredential: 'stable-member-credential',
      idempotencyKey: 'stable-idempotency-key',
    }

    await roomApi.redeemInvite('invite-1', redemption)
    await roomApi.getRoom('room-1', 'stable-member-credential')
    await roomApi.getDevices('room-1', 'stable-member-credential')

    expect(requestDetails(fetchMock, 0)).toEqual({
      url: `${SIGNALING_URL}/v2/invites/invite-1/redeem`,
      method: 'POST',
      headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
      body: redemption,
    })
    expect(requestDetails(fetchMock, 1)).toEqual({
      url: `${SIGNALING_URL}/v2/rooms/room-1`,
      method: 'GET',
      headers: { Accept: 'application/json', Authorization: 'Bearer stable-member-credential' },
      body: undefined,
    })
    expect(requestDetails(fetchMock, 2).url).toBe(`${SIGNALING_URL}/v2/rooms/room-1/devices`)
  })

  it('uses canonical checkpoint and per-device operation mailboxes', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(new Response(null, { status: 204 }))
      .mockResolvedValueOnce(jsonResponse({
        roomId: 'room-1',
        uploaderDeviceId: 'device-owner',
        envelope: 'checkpoint-ciphertext',
        updatedAt: 'now',
        expiresAt: 'later',
      }))
      .mockResolvedValueOnce(jsonResponse({ operationId: 9 }, 201))
      .mockResolvedValueOnce(jsonResponse({ operations: [] }))
    vi.stubGlobal('fetch', fetchMock)

    await roomApi.putCheckpoint('room-1', 'credential', 'checkpoint-ciphertext')
    await roomApi.getCheckpoint('room-1', 'credential')
    await roomApi.postOperation('room-1', 'device-2', 'credential', 'operation-ciphertext')
    await roomApi.getOperations('room-1', 'device-2', 'credential', 8)

    expect(requestDetails(fetchMock, 0)).toEqual({
      url: `${SIGNALING_URL}/v2/rooms/room-1/mailbox/checkpoint`,
      method: 'PUT',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
        Authorization: 'Bearer credential',
      },
      body: { envelope: 'checkpoint-ciphertext', expiresInSeconds: 604_800 },
    })
    expect(requestDetails(fetchMock, 1).url).toBe(`${SIGNALING_URL}/v2/rooms/room-1/mailbox/checkpoint`)
    expect(requestDetails(fetchMock, 2).url).toBe(`${SIGNALING_URL}/v2/rooms/room-1/mailbox/device-2/operations`)
    expect(requestDetails(fetchMock, 2).body).toEqual({
      envelope: 'operation-ciphertext',
      expiresInSeconds: 604_800,
    })
    expect(requestDetails(fetchMock, 3).url).toBe(`${SIGNALING_URL}/v2/rooms/room-1/mailbox/device-2/operations?after=8`)
  })

  it('uses unified addressed signaling and parses nested API errors', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(jsonResponse({ signalId: 4 }, 201))
      .mockResolvedValueOnce(jsonResponse({ signals: [] }))
      .mockResolvedValueOnce(jsonResponse({
        error: { code: 'room_full', message: 'room has reached its member capacity' },
      }, 409))
    vi.stubGlobal('fetch', fetchMock)

    await roomApi.postSignal('room-1', 'session_12345678', 'credential', {
      kind: 'offer',
      fromDeviceId: 'device-1',
      toDeviceId: 'device-2',
      envelope: 'encrypted-offer',
    })
    await roomApi.getSignals('room-1', 'session_12345678', 'credential', 3)
    const rejected = roomApi.redeemInvite('invite-full', {
      inviteSecret: 'secret',
      deviceId: 'device',
      deviceIdentity: 'identity',
      memberCredential: 'credential',
      idempotencyKey: 'idempotency',
    })

    expect(requestDetails(fetchMock, 0)).toEqual({
      url: `${SIGNALING_URL}/v2/rooms/room-1/sessions/session_12345678/signals`,
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
        Authorization: 'Bearer credential',
      },
      body: {
        kind: 'offer',
        fromDeviceId: 'device-1',
        toDeviceId: 'device-2',
        envelope: 'encrypted-offer',
        expiresInSeconds: 600,
      },
    })
    expect(requestDetails(fetchMock, 1).url).toBe(
      `${SIGNALING_URL}/v2/rooms/room-1/sessions/session_12345678/signals?after=3`,
    )
    await expect(rejected).rejects.toMatchObject({
      code: 'room_full',
      message: 'room has reached its member capacity',
      status: 409,
    })
  })
})
