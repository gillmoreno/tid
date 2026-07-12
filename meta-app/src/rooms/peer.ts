import { roomApi, type SignalKind } from './api'
import { decryptJson, deriveStableToken, encryptJson } from './crypto'
import type { ConnectionStatus, EncryptedPayload, VaultRoom } from './types'

type PeerCallbacks = {
  onPayload: (payload: EncryptedPayload) => void
  onStatus: (status: ConnectionStatus) => void
  getSignalCursor: (sessionId: string) => number
  onSignalCursor: (sessionId: string, cursor: number) => void
}

type SignalPayload = {
  description?: RTCSessionDescriptionInit
  candidate?: RTCIceCandidateInit
}

export async function deterministicSessionId(
  roomId: string,
  firstDeviceId: string,
  secondDeviceId: string,
): Promise<string> {
  const devices = [firstDeviceId, secondDeviceId].sort().join(':')
  return `session_${(await deriveStableToken(roomId, `devices:${devices}`)).slice(0, 40)}`
}

export class RoomPeer {
  private connection?: RTCPeerConnection
  private channel?: RTCDataChannel
  private sessionId?: string
  private signalCursor = 0
  private pollTimer?: number
  private pendingCandidates: RTCIceCandidateInit[] = []
  private readonly room: VaultRoom
  private readonly remoteDeviceId: string
  private readonly callbacks: PeerCallbacks
  private initiator = false

  constructor(room: VaultRoom, remoteDeviceId: string, callbacks: PeerCallbacks) {
    this.room = room
    this.remoteDeviceId = remoteDeviceId
    this.callbacks = callbacks
  }

  async connect(): Promise<void> {
    if (!navigator.onLine) {
      this.callbacks.onStatus('offline')
      return
    }
    if (typeof RTCPeerConnection === 'undefined') {
      this.callbacks.onStatus('unavailable')
      return
    }
    this.callbacks.onStatus('connecting')
    try {
      this.sessionId = await deterministicSessionId(
        this.room.id,
        this.room.deviceId,
        this.remoteDeviceId,
      )
      this.signalCursor = this.callbacks.getSignalCursor(this.sessionId)
      this.initiator = this.room.deviceId.localeCompare(this.remoteDeviceId) < 0
      this.connection = new RTCPeerConnection({
        iceServers: [{ urls: ['stun:stun.l.google.com:19302', 'stun:stun.cloudflare.com:3478'] }],
      })
      this.connection.onconnectionstatechange = () => {
        const state = this.connection?.connectionState
        if (state === 'connected') this.callbacks.onStatus('p2p')
        if (state === 'failed' || state === 'disconnected' || state === 'closed') {
          this.callbacks.onStatus(navigator.onLine ? 'mailbox' : 'offline')
        }
      }
      this.connection.onicecandidate = (event) => {
        if (event.candidate) void this.postSignal('candidate', { candidate: event.candidate.toJSON() })
      }
      if (this.initiator) {
        this.attachChannel(this.connection.createDataChannel('meta-room-operations', { ordered: true }))
        const offer = await this.connection.createOffer()
        await this.connection.setLocalDescription(offer)
        await this.postSignal('offer', { description: offer })
      } else {
        this.connection.ondatachannel = (event) => this.attachChannel(event.channel)
      }
      this.pollTimer = window.setInterval(() => void this.pollSignals(), 1_500)
      void this.pollSignals()
    } catch {
      this.closeConnection()
      this.callbacks.onStatus(navigator.onLine ? 'mailbox' : 'offline')
    }
  }

  send(payload: EncryptedPayload): boolean {
    if (this.channel?.readyState !== 'open') return false
    this.channel.send(JSON.stringify({ type: 'encrypted-operation', payload }))
    return true
  }

  close(): void {
    this.closeConnection()
    this.callbacks.onStatus(navigator.onLine ? 'mailbox' : 'offline')
  }

  private signalContext(): string {
    return `room:${this.room.id}:signal:${this.sessionId}`
  }

  private async postSignal(kind: SignalKind, payload: SignalPayload): Promise<void> {
    if (!this.sessionId) return
    const envelope = JSON.stringify(await encryptJson(this.room.roomDataKey, payload, this.signalContext()))
    await roomApi.postSignal(this.room.id, this.sessionId, this.room.memberCredential, {
      kind,
      fromDeviceId: this.room.deviceId,
      toDeviceId: this.remoteDeviceId,
      envelope,
    })
  }

  private async pollSignals(): Promise<void> {
    if (!this.connection || !this.sessionId) return
    try {
      const response = await roomApi.getSignals(
        this.room.id,
        this.sessionId,
        this.room.memberCredential,
        this.signalCursor,
      )
      for (const signal of response.signals) {
        const payload = await decryptJson<SignalPayload>(
          this.room.roomDataKey,
          parseEncryptedPayload(signal.envelope),
          this.signalContext(),
        )
        if (signal.kind === 'offer' && !this.initiator && payload.description && !this.connection.remoteDescription) {
          await this.connection.setRemoteDescription(payload.description)
          const answer = await this.connection.createAnswer()
          await this.connection.setLocalDescription(answer)
          await this.postSignal('answer', { description: answer })
          await this.flushCandidates()
        }
        if (signal.kind === 'answer' && this.initiator && payload.description && !this.connection.remoteDescription) {
          await this.connection.setRemoteDescription(payload.description)
          await this.flushCandidates()
        }
        if (signal.kind === 'candidate' && payload.candidate) {
          if (this.connection.remoteDescription) await this.connection.addIceCandidate(payload.candidate)
          else this.pendingCandidates.push(payload.candidate)
        }
        this.signalCursor = Math.max(this.signalCursor, signal.signalId)
      }
      this.callbacks.onSignalCursor(this.sessionId, this.signalCursor)
    } catch {
      // Keep polling; mailbox delivery remains available while signaling recovers.
    }
  }

  private async flushCandidates(): Promise<void> {
    if (!this.connection?.remoteDescription) return
    for (const candidate of this.pendingCandidates.splice(0)) {
      await this.connection.addIceCandidate(candidate)
    }
  }

  private attachChannel(channel: RTCDataChannel): void {
    this.channel = channel
    channel.onopen = () => this.callbacks.onStatus('p2p')
    channel.onclose = () => this.callbacks.onStatus(navigator.onLine ? 'mailbox' : 'offline')
    channel.onmessage = (event) => {
      if (typeof event.data !== 'string' || event.data.length > 128_000) return
      try {
        const message = JSON.parse(event.data) as { type?: unknown; payload?: unknown }
        if (message.type !== 'encrypted-operation' || !isEncryptedPayload(message.payload)) return
        this.callbacks.onPayload(message.payload)
      } catch {
        // Ignore malformed peer data.
      }
    }
  }

  private closeConnection(): void {
    if (this.pollTimer) window.clearInterval(this.pollTimer)
    this.pollTimer = undefined
    this.channel?.close()
    this.connection?.close()
    this.channel = undefined
    this.connection = undefined
  }
}

function parseEncryptedPayload(value: string): EncryptedPayload {
  const parsed = JSON.parse(value) as unknown
  if (!isEncryptedPayload(parsed)) throw new Error('Invalid signal envelope')
  return parsed
}

function isEncryptedPayload(value: unknown): value is EncryptedPayload {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Record<string, unknown>
  return candidate.algorithm === 'AES-GCM'
    && candidate.version === 1
    && typeof candidate.iv === 'string'
    && typeof candidate.ciphertext === 'string'
    && candidate.iv.length <= 128
    && candidate.ciphertext.length <= 128_000
}
