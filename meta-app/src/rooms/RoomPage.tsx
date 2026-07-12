import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { RoomFrame } from './RoomFrame'
import { loadBundle, RoomSync, type CodeBundle } from './roomService'
import type { ConnectionStatus, VaultRoom } from './types'
import { getRoom } from './vault'

const statusCopy: Record<ConnectionStatus, string> = {
  offline: 'Offline · changes queued',
  mailbox: 'Mailbox fallback · syncing',
  connecting: 'Opening secure peer channel',
  p2p: 'Live peer channel',
  unavailable: 'P2P unavailable · local mode',
}

export function RoomPage() {
  const { roomId = '' } = useParams()
  const [room, setRoom] = useState<VaultRoom>()
  const [bundle, setBundle] = useState<CodeBundle>()
  const [status, setStatus] = useState<ConnectionStatus>(navigator.onLine ? 'connecting' : 'offline')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [copied, setCopied] = useState('')

  useEffect(() => {
    let active = true
    void getRoom(roomId)
      .then(async (stored) => {
        if (!active || !stored) return
        setRoom(stored)
        setBundle(await loadBundle(stored))
      })
      .catch(() => setError('The encrypted room data could not be opened on this device.'))
      .finally(() => setLoading(false))
    return () => { active = false }
  }, [roomId])

  useEffect(() => {
    if (!room) return
    const sync = new RoomSync(room, setStatus)
    sync.start()
    return () => sync.stop()
  }, [room])

  async function copy(value: string, label: string) {
    await navigator.clipboard.writeText(value)
    setCopied(label)
    window.setTimeout(() => setCopied(''), 1_500)
  }

  if (loading) return <div className="room-loading">Opening encrypted room…</div>
  if (!room || !bundle) {
    return (
      <section className="missing-room">
        <div className="kicker">Room unavailable</div>
        <h1>This room is not in this device’s vault.</h1>
        <p>{error || 'Use the original invitation URL and key to join. No placeholder room was created.'}</p>
        <Link className="primary-button centered" to="/">Return home</Link>
      </section>
    )
  }

  return (
    <div className="room-page">
      <aside className="room-sidebar">
        <Link className="back-link" to="/">← All rooms</Link>
        <div>
          <div className="kicker">{room.role} room</div>
          <h1>{room.title}</h1>
          <p className="room-id">{room.id}</p>
        </div>
        <div className={`connection-state state-${status}`}>
          <span />
          <div>
            <strong>{statusCopy[status]}</strong>
            <small>Encrypted operations persist before send.</small>
          </div>
        </div>
        <dl className="room-facts">
          <div><dt>Seats</dt><dd>{room.capacity} unique members</dd></div>
          <div><dt>State</dt><dd>AES-GCM vault</dd></div>
          <div><dt>Transport</dt><dd>WebRTC + mailbox</dd></div>
        </dl>
        {room.role === 'owner' && room.shareUrl && room.invitationPackage && (
          <section className="invite-panel">
            <h2>Invitation</h2>
            <button type="button" onClick={() => void copy(room.shareUrl!, 'link')}>
              {copied === 'link' ? 'Link copied' : 'Copy public link'}
            </button>
            <button type="button" onClick={() => void copy(room.invitationPackage!, 'key')}>
              {copied === 'key' ? 'Package copied' : 'Copy invitation package'}
            </button>
          </section>
        )}
      </aside>
      <section className="room-stage">
        <div className="stage-label">
          <span>Sandboxed application</span>
          <span>network blocked</span>
        </div>
        <RoomFrame room={room} bundle={bundle} />
      </section>
    </div>
  )
}
