import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

export function Shell({ children }: { children: ReactNode }) {
  return (
    <div className="site-shell">
      <header className="masthead">
        <Link className="wordmark" to="/" aria-label="Roomworks home">
          <span className="wordmark-mark">R</span>
          <span>Roomworks</span>
        </Link>
        <div className="masthead-note">Local-first rooms · encrypted by default</div>
      </header>
      <main>{children}</main>
      <footer className="footer">
        <span>Room shell</span>
        <span>localhost:5200</span>
      </footer>
    </div>
  )
}
