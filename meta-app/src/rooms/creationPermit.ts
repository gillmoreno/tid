import { RoomApiError } from './api'

const CREATOR_PERMIT_ERRORS = new Set([
  'creator_permit_required',
  'invalid_creator_permit',
  'creator_permit_expired',
  'creator_permit_not_yet_valid',
  'creator_permit_capacity_mismatch',
  'creator_permit_used',
])

export function isCreatorPermitError(error: unknown): error is RoomApiError {
  return error instanceof RoomApiError && CREATOR_PERMIT_ERRORS.has(error.code)
}

// This is only a display hint. The Go service verifies the signature and capacity.
export function creatorPermitCapacityHint(permit: string): number | undefined {
  const parts = permit.split('.')
  if (parts.length !== 3 || parts[0] !== 'rwp1') return undefined
  try {
    const encoded = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const padded = encoded.padEnd(Math.ceil(encoded.length / 4) * 4, '=')
    const bytes = Uint8Array.from(atob(padded), (character) => character.charCodeAt(0))
    const claims = JSON.parse(new TextDecoder().decode(bytes)) as Record<string, unknown>
    const capacity = claims.maxMembers
    if (
      claims.v !== 1
      || claims.purpose !== 'create_room'
      || typeof capacity !== 'number'
      || !Number.isInteger(capacity)
      || capacity < 2
      || capacity > 50
    ) return undefined
    return capacity
  } catch {
    return undefined
  }
}
