import { describe, expect, it } from 'vitest'
import { RoomApiError } from './api'
import { creatorPermitCapacityHint, isCreatorPermitError } from './creationPermit'

function shapedPermit(claims: Record<string, unknown>): string {
  const payload = btoa(JSON.stringify(claims))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')
  return `rwp1.${payload}.signature-not-checked-by-browser`
}

const baseClaims = {
  v: 1,
  purpose: 'create_room',
  jti: 'public-token-id',
  maxMembers: 11,
  iat: 1,
  exp: 2,
}

describe('creator permit hints', () => {
  it('reads a capacity hint without treating the signature as authorization', () => {
    expect(creatorPermitCapacityHint(shapedPermit(baseClaims))).toBe(11)
    expect(creatorPermitCapacityHint(shapedPermit({ ...baseClaims, maxMembers: 10 }))).toBe(10)
  })

  it.each([
    '',
    'rwp1.only-two-parts',
    shapedPermit({ ...baseClaims, v: 2 }),
    shapedPermit({ ...baseClaims, purpose: 'anything_else' }),
    shapedPermit({ ...baseClaims, maxMembers: 1 }),
    shapedPermit({ ...baseClaims, maxMembers: 51 }),
    shapedPermit({ ...baseClaims, maxMembers: 3.5 }),
  ])('rejects malformed or out-of-range hints', (permit) => {
    expect(creatorPermitCapacityHint(permit)).toBeUndefined()
  })

  it.each([
    'creator_permit_required',
    'invalid_creator_permit',
    'creator_permit_expired',
    'creator_permit_not_yet_valid',
    'creator_permit_capacity_mismatch',
    'creator_permit_used',
  ])('classifies %s as a permit error', (code) => {
    expect(isCreatorPermitError(new RoomApiError('rejected', 403, code))).toBe(true)
  })

  it('does not clear permits for unrelated failures', () => {
    expect(isCreatorPermitError(new RoomApiError('offline', 0, 'offline'))).toBe(false)
    expect(isCreatorPermitError(new RoomApiError('bad capacity', 400, 'invalid_max_members'))).toBe(false)
  })
})
