import type { EncryptedPayload } from './types'

const encoder = new TextEncoder()
const decoder = new TextDecoder()

function bytesToBase64Url(bytes: Uint8Array): string {
  let binary = ''
  for (const byte of bytes) binary += String.fromCharCode(byte)
  return btoa(binary).replaceAll('+', '-').replaceAll('/', '_').replace(/=+$/u, '')
}

function base64UrlToBytes(value: string): Uint8Array<ArrayBuffer> {
  const padded = value.replaceAll('-', '+').replaceAll('_', '/').padEnd(Math.ceil(value.length / 4) * 4, '=')
  const binary = atob(padded)
  const bytes = new Uint8Array(binary.length)
  for (let index = 0; index < binary.length; index += 1) bytes[index] = binary.charCodeAt(index)
  return bytes
}

export function randomToken(byteLength = 32): string {
  return bytesToBase64Url(crypto.getRandomValues(new Uint8Array(byteLength)))
}

export async function deriveStableToken(secret: string, context: string): Promise<string> {
  const digest = await crypto.subtle.digest('SHA-256', encoder.encode(`${context}\u0000${secret}`))
  return bytesToBase64Url(new Uint8Array(digest))
}

export async function generateRoomDataKey(): Promise<CryptoKey> {
  return crypto.subtle.generateKey({ name: 'AES-GCM', length: 256 }, true, ['encrypt', 'decrypt'])
}

export async function deriveInviteKey(invitationKey: string): Promise<CryptoKey> {
  const digest = await crypto.subtle.digest('SHA-256', encoder.encode(invitationKey))
  return crypto.subtle.importKey('raw', digest, { name: 'AES-GCM' }, false, ['encrypt', 'decrypt'])
}

export async function encryptBytes(
  key: CryptoKey,
  value: Uint8Array<ArrayBuffer>,
  context: string,
): Promise<EncryptedPayload> {
  const iv = crypto.getRandomValues(new Uint8Array(12))
  const ciphertext = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv, additionalData: encoder.encode(context) },
    key,
    value,
  )
  return {
    algorithm: 'AES-GCM',
    iv: bytesToBase64Url(iv),
    ciphertext: bytesToBase64Url(new Uint8Array(ciphertext)),
    version: 1,
  }
}

export async function decryptBytes(
  key: CryptoKey,
  payload: EncryptedPayload,
  context: string,
): Promise<Uint8Array<ArrayBuffer>> {
  if (payload.algorithm !== 'AES-GCM' || payload.version !== 1) throw new Error('Unsupported encrypted payload')
  const plaintext = await crypto.subtle.decrypt(
    {
      name: 'AES-GCM',
      iv: base64UrlToBytes(payload.iv),
      additionalData: encoder.encode(context),
    },
    key,
    base64UrlToBytes(payload.ciphertext),
  )
  return new Uint8Array(plaintext)
}

export async function encryptJson<T>(key: CryptoKey, value: T, context: string): Promise<EncryptedPayload> {
  return encryptBytes(key, encoder.encode(JSON.stringify(value)), context)
}

export async function decryptJson<T>(key: CryptoKey, payload: EncryptedPayload, context: string): Promise<T> {
  const bytes = await decryptBytes(key, payload, context)
  return JSON.parse(decoder.decode(bytes)) as T
}

export async function wrapRoomDataKey(roomDataKey: CryptoKey, invitationKey: string): Promise<EncryptedPayload> {
  const rawKey = new Uint8Array(await crypto.subtle.exportKey('raw', roomDataKey))
  return encryptBytes(await deriveInviteKey(invitationKey), rawKey, 'meta-room-data-key:v1')
}

export async function unwrapRoomDataKey(
  envelope: EncryptedPayload,
  invitationKey: string,
): Promise<CryptoKey> {
  const rawKey = await decryptBytes(await deriveInviteKey(invitationKey), envelope, 'meta-room-data-key:v1')
  return crypto.subtle.importKey('raw', rawKey, { name: 'AES-GCM' }, true, ['encrypt', 'decrypt'])
}
