import { afterEach, describe, expect, it, vi } from 'vitest'
import { parseInvitationPackage } from './invitation'
import { createRoom, loadBundle, redeemInvitation, roomState } from './roomService'
import { clearVaultForTests, getRoom, listRooms } from './vault'

afterEach(async () => {
  vi.unstubAllGlobals()
  await clearVaultForTests()
})

describe('room create and offline-owner redemption', () => {
  it('packages client key material and bootstraps a clean vault from the durable checkpoint', async () => {
    let storedCheckpoint = ''
    const redemptionBodies: unknown[] = []
    let redemptionAttempt = 0
    const fetchMock = vi.fn(async (input: string | URL | Request, init?: RequestInit) => {
      const url = String(input)
      const path = new URL(url).pathname
      const method = init?.method ?? 'GET'
      const body = init?.body ? JSON.parse(String(init.body)) as Record<string, unknown> : {}

      if (method === 'POST' && path === '/v2/rooms') {
        const headers = init?.headers as Record<string, string> | undefined
        expect(headers?.['X-Room-Creator-Permit']).toBe('room-creator-permit')
        return json({
          roomId: 'room_12345678',
          maxMembers: 2,
          ownerCapability: 'owner_capability_secret_value',
          ownerMemberId: 'member_owner_123456',
          ownerDeviceId: 'device_owner_123456',
          ownerMemberCredential: 'owner_member_credential_secret',
        }, 201)
      }
      if (method === 'POST' && path === '/v2/rooms/room_12345678/invites') {
        const headers = init?.headers as Record<string, string> | undefined
        expect(headers?.['X-Owner-Capability']).toBe('owner_capability_secret_value')
        expect(body).toEqual({ expiresInSeconds: 86_400 })
        return json({
          inviteId: 'invite_12345678',
          inviteSecret: 'backend_invite_secret_value',
          expiresAt: '2026-07-13T00:00:00Z',
        }, 201)
      }
      if (method === 'PUT' && path === '/v2/rooms/room_12345678/mailbox/checkpoint') {
        storedCheckpoint = String(body.envelope)
        return new Response(null, { status: 204 })
      }
      if (method === 'POST' && path === '/v2/invites/invite_12345678/redeem') {
        redemptionBodies.push(body)
        redemptionAttempt += 1
        if (redemptionAttempt === 1) {
          return json({ error: { code: 'temporarily_unavailable', message: 'retry redemption' } }, 503)
        }
        return json({
          roomId: 'room_12345678',
          memberId: 'member_new_123456',
          deviceId: body.deviceId,
          reconnected: false,
        }, 201)
      }
      if (method === 'GET' && path === '/v2/rooms/room_12345678') {
        return json({ roomId: 'room_12345678', maxMembers: 2, memberCount: 2 })
      }
      if (method === 'GET' && path === '/v2/rooms/room_12345678/mailbox/checkpoint') {
        return json({
          roomId: 'room_12345678',
          uploaderDeviceId: 'device_owner_123456',
          envelope: storedCheckpoint,
          updatedAt: '2026-07-12T12:00:00Z',
          expiresAt: '2026-07-19T12:00:00Z',
        })
      }
      if (method === 'GET' && path === '/v2/rooms/room_12345678/devices') {
        return json({
          roomId: 'room_12345678',
          devices: [
            { deviceId: 'device_owner_123456', isOwner: true, isSelf: false },
            { deviceId: String((redemptionBodies[1] as Record<string, unknown>).deviceId), isOwner: false, isSelf: true },
          ],
        })
      }
      throw new Error(`Unexpected request: ${method} ${path}`)
    })
    vi.stubGlobal('fetch', fetchMock)

    const created = await createRoom('Offline shared counter', 2, 'room-creator-permit')
    const invitation = parseInvitationPackage(created.invitationKey, 'invite_12345678')
    expect(invitation).toMatchObject({
      inviteId: 'invite_12345678',
      inviteSecret: 'backend_invite_secret_value',
      roomId: 'room_12345678',
      ownerDeviceId: 'device_owner_123456',
    })
    expect(storedCheckpoint).not.toContain('Offline shared counter')
    expect(storedCheckpoint).not.toContain('backend_invite_secret_value')

    // Simulate a brand-new browser. The owner makes no further request.
    await clearVaultForTests()
    await expect(redeemInvitation('invite_12345678', created.invitationKey)).rejects.toThrow('retry redemption')
    expect(await listRooms()).toEqual([])
    const joined = await redeemInvitation('invite_12345678', created.invitationKey)

    expect(redemptionBodies).toHaveLength(2)
    expect(redemptionBodies[0]).toEqual(redemptionBodies[1])
    expect(redemptionBodies[1]).toMatchObject({
      inviteSecret: 'backend_invite_secret_value',
    })
    expect(redemptionBodies[1]).not.toHaveProperty('keyEnvelope')
    expect(redemptionBodies[1]).not.toHaveProperty('roomDataKey')
    expect(joined.role).toBe('member')
    expect(await getRoom('room_12345678')).toBeDefined()
    await expect(loadBundle(joined)).resolves.toEqual({
      version: 1,
      kind: 'counter',
      title: 'Offline shared counter',
    })
    await expect(roomState(joined)).resolves.toEqual({ value: 0 })
  })
})

function json(value: unknown, status = 200): Response {
  return new Response(JSON.stringify(value), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}
