import type { EncryptedPayload } from './types'

const encoder = new TextEncoder()
const decoder = new TextDecoder()
const PREFIX = 'roompkg1.'

export type InvitationPackage = {
  version: 1
  inviteId: string
  inviteSecret: string
  roomId: string
  ownerDeviceId: string
  keyEnvelope: EncryptedPayload
}

function encodeBase64Url(bytes: Uint8Array): string {
  let binary = ''
  for (const byte of bytes) binary += String.fromCharCode(byte)
  return btoa(binary).replaceAll('+', '-').replaceAll('/', '_').replace(/=+$/u, '')
}

function decodeBase64Url(value: string): Uint8Array {
  const padded = value.replaceAll('-', '+').replaceAll('_', '/').padEnd(Math.ceil(value.length / 4) * 4, '=')
  const binary = atob(padded)
  return Uint8Array.from(binary, (character) => character.charCodeAt(0))
}

function isEncryptedPayload(value: unknown): value is EncryptedPayload {
  if (!value || typeof value !== 'object') return false
  const payload = value as Record<string, unknown>
  return payload.algorithm === 'AES-GCM'
    && payload.version === 1
    && typeof payload.iv === 'string'
    && payload.iv.length > 0
    && payload.iv.length <= 128
    && typeof payload.ciphertext === 'string'
    && payload.ciphertext.length > 0
    && payload.ciphertext.length <= 16_384
}

export function encodeInvitationPackage(value: InvitationPackage): string {
  return `${PREFIX}${encodeBase64Url(encoder.encode(JSON.stringify(value)))}`
}

export function parseInvitationPackage(value: string, expectedInviteId: string): InvitationPackage {
  if (!value.startsWith(PREFIX) || value.length > 32_768) throw new Error('Invalid invitation package')
  let parsed: unknown
  try {
    parsed = JSON.parse(decoder.decode(decodeBase64Url(value.slice(PREFIX.length))))
  } catch {
    throw new Error('Invalid invitation package')
  }
  if (!parsed || typeof parsed !== 'object') throw new Error('Invalid invitation package')
  const candidate = parsed as Record<string, unknown>
  if (
    candidate.version !== 1
    || candidate.inviteId !== expectedInviteId
    || typeof candidate.inviteSecret !== 'string'
    || candidate.inviteSecret.length < 16
    || typeof candidate.roomId !== 'string'
    || candidate.roomId.length < 8
    || typeof candidate.ownerDeviceId !== 'string'
    || candidate.ownerDeviceId.length < 8
    || !isEncryptedPayload(candidate.keyEnvelope)
  ) throw new Error('Invitation package does not match this invitation')
  return candidate as InvitationPackage
}
