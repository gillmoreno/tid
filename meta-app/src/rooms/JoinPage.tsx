import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { RoomApiError } from './api'
import { redeemInvitation } from './roomService'

export function JoinPage() {
  const { inviteId = '' } = useParams()
  const navigate = useNavigate()
  const [key, setKey] = useState('')
  const [joining, setJoining] = useState(false)
  const [error, setError] = useState('')

  async function submit(event: React.FormEvent) {
    event.preventDefault()
    if (!key.trim()) return
    setJoining(true)
    setError('')
    try {
      const room = await redeemInvitation(inviteId, key.trim())
      navigate(`/rooms/${encodeURIComponent(room.id)}`, { replace: true })
    } catch (caught) {
      if (caught instanceof RoomApiError) setError(caught.message)
      else if (caught instanceof DOMException) setError('The invitation package could not decrypt this room.')
      else setError('This invitation could not be verified. Check the package and try again.')
    } finally {
      setJoining(false)
    }
  }

  return (
    <div className="join-page">
      <section className="join-intro">
        <Link className="back-link" to="/">← Roomworks</Link>
        <div>
          <div className="kicker">Private invitation</div>
          <h1>A room is waiting.</h1>
          <p>
            This public URL identifies the invitation, but contains no room secret.
            Paste the private package you received separately to redeem your membership.
          </p>
        </div>
        <div className="join-id">
          <span>Invitation</span>
          <code>{inviteId}</code>
        </div>
      </section>
      <form className="join-form" onSubmit={submit}>
        <div className="join-step">01 / 01</div>
        <h2>Unlock the envelope</h2>
        <label>
          Invitation package
          <input
            autoFocus
            type="password"
            value={key}
            onChange={(event) => setKey(event.target.value)}
            placeholder="roompkg1.…"
            autoComplete="off"
            spellCheck="false"
          />
        </label>
        <p className="form-note">
          Only its one-time invite secret is redeemed. The wrapped room key stays in this browser.
        </p>
        {error && <div className="error-banner" role="alert">{error}</div>}
        <button className="primary-button full" disabled={joining || !key.trim()}>
          {joining ? 'Verifying & downloading checkpoint…' : 'Join encrypted room'}
        </button>
        <div className="join-safety">A failed, expired, revoked, or full invitation is never saved to this device.</div>
      </form>
    </div>
  )
}
