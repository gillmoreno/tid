import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { RoomApiError, SIGNALING_URL } from './api'
import { createRoom, type CreatedRoom } from './roomService'
import type { VaultRoom } from './types'
import { listRooms } from './vault'

async function copy(value: string, setCopied: (label: string) => void, label: string) {
  await navigator.clipboard.writeText(value)
  setCopied(label)
  window.setTimeout(() => setCopied(''), 1_500)
}

export function HomePage() {
  const [rooms, setRooms] = useState<VaultRoom[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [title, setTitle] = useState('A shared counter')
  const [capacity, setCapacity] = useState(2)
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState('')
  const [created, setCreated] = useState<CreatedRoom>()
  const [copied, setCopied] = useState('')

  useEffect(() => {
    void listRooms().then(setRooms).finally(() => setLoading(false))
  }, [])

  async function submit(event: React.FormEvent) {
    event.preventDefault()
    setError('')
    setCreating(true)
    try {
      const result = await createRoom(title.trim(), capacity)
      setCreated(result)
      setRooms((current) => [result.room, ...current.filter((room) => room.id !== result.room.id)])
      setShowCreate(false)
    } catch (caught) {
      setError(caught instanceof RoomApiError ? caught.message : 'The encrypted room could not be created.')
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="home">
      <section className="hero-panel">
        <div className="hero-copy">
          <div className="kicker">Private collaborative software</div>
          <h1>Make a room.<br />Keep the keys.</h1>
          <p>
            A durable, encrypted workspace that continues locally when the network disappears.
            Your application data key never leaves this browser.
          </p>
          <button className="primary-button" onClick={() => setShowCreate(true)}>Create a room</button>
        </div>
        <div className="hero-diagram" aria-hidden="true">
          <div className="diagram-node diagram-you">YOU</div>
          <div className="diagram-line" />
          <div className="diagram-vault">AES<br /><span>256</span></div>
          <div className="diagram-line" />
          <div className="diagram-node">PEER</div>
          <div className="diagram-caption">P2P when possible<br />mailbox when necessary</div>
        </div>
      </section>

      <section className="rooms-section">
        <div className="section-heading">
          <div>
            <div className="kicker">This device</div>
            <h2>Your rooms</h2>
          </div>
          <span className="endpoint-pill" title={SIGNALING_URL}>service · {new URL(SIGNALING_URL).port || '443'}</span>
        </div>
        {loading && <div className="empty-state">Opening the local vault…</div>}
        {!loading && rooms.length === 0 && (
          <div className="empty-state">
            <span>01</span>
            No rooms live in this browser yet.
          </div>
        )}
        <div className="room-list">
          {rooms.map((room, index) => (
            <Link className="room-row" to={`/rooms/${encodeURIComponent(room.id)}`} key={room.id}>
              <span className="room-index">{String(index + 1).padStart(2, '0')}</span>
              <span className="room-main">
                <strong>{room.title}</strong>
                <small>{room.role} · capacity {room.capacity} unique members</small>
              </span>
              <span className="room-arrow">↗</span>
            </Link>
          ))}
        </div>
      </section>

      {showCreate && (
        <div className="modal-backdrop" role="presentation">
          <form className="modal-card" onSubmit={submit} aria-label="Create room">
            <button className="modal-close" type="button" onClick={() => setShowCreate(false)} aria-label="Close">×</button>
            <div className="kicker">New encrypted room</div>
            <h2>Start with a small instrument.</h2>
            <label>
              Room purpose
              <input value={title} onChange={(event) => setTitle(event.target.value)} maxLength={80} required />
            </label>
            <label>
              Unique member capacity
              <input
                type="number"
                min="2"
                max="20"
                value={capacity}
                onChange={(event) => setCapacity(Math.min(20, Math.max(2, Number(event.target.value))))}
                required
              />
            </label>
            <p className="form-note">Reopening from an admitted device does not consume another seat.</p>
            {error && <div className="error-banner" role="alert">{error}</div>}
            <button className="primary-button full" disabled={creating || !title.trim()}>
              {creating ? 'Creating durable room…' : 'Create room & invitation'}
            </button>
          </form>
        </div>
      )}

      {created && (
        <div className="modal-backdrop" role="presentation">
          <section className="modal-card share-card" aria-label="Room invitation">
            <div className="kicker">Room ready</div>
            <h2>Send these separately.</h2>
            <p className="form-note">The public link identifies the invitation. The private package carries its one-time secret and encrypted key envelope.</p>
            {created.checkpointQueued && (
              <div className="error-banner" role="status">
                The encrypted checkpoint is queued locally. Open the room while online before the other member joins.
              </div>
            )}
            <div className="copy-block">
              <span>Public share link</span>
              <code>{created.invitationUrl}</code>
              <button type="button" onClick={() => void copy(created.invitationUrl, setCopied, 'link')}>
                {copied === 'link' ? 'Copied' : 'Copy link'}
              </button>
            </div>
            <div className="copy-block secret">
              <span>Private invitation package</span>
              <code>{created.invitationKey}</code>
              <button type="button" onClick={() => void copy(created.invitationKey, setCopied, 'key')}>
                {copied === 'key' ? 'Copied' : 'Copy key'}
              </button>
            </div>
            <Link className="primary-button full centered" to={`/rooms/${encodeURIComponent(created.room.id)}`}>
              Enter room
            </Link>
            <button className="text-button" type="button" onClick={() => setCreated(undefined)}>Close</button>
          </section>
        </div>
      )}
    </div>
  )
}
