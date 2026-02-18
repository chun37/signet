import { useState, useEffect } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { api } from '../api/client'

export default function Layout() {
  const [menuOpen, setMenuOpen] = useState(false)
  const [pendingCount, setPendingCount] = useState(0)

  useEffect(() => {
    api.getPending().then(data => setPendingCount((data ?? []).length)).catch(() => {})
    const id = setInterval(() => {
      api.getPending().then(data => setPendingCount((data ?? []).length)).catch(() => {})
    }, 10000)
    return () => clearInterval(id)
  }, [])

  const closeMenu = () => setMenuOpen(false)

  return (
    <div className="app">
      <header className="header">
        <div className="header-inner">
          <span className="logo">Signet</span>
          <button className="nav-burger" onClick={() => setMenuOpen(!menuOpen)}>
            {menuOpen ? '\u2715' : '\u2630'}
          </button>
          <nav className={`nav ${menuOpen ? 'open' : ''}`}>
            <NavLink to="/" end onClick={closeMenu}>
              Dashboard
            </NavLink>
            <NavLink to="/transactions" onClick={closeMenu}>
              Transactions
            </NavLink>
            <NavLink to="/pending" onClick={closeMenu}>
              Pending
              {pendingCount > 0 && <span className="badge">{pendingCount}</span>}
            </NavLink>
            <NavLink to="/propose" onClick={closeMenu}>
              Propose
            </NavLink>
            <NavLink to="/nodes" onClick={closeMenu}>
              Nodes
            </NavLink>
          </nav>
        </div>
      </header>
      {menuOpen && <div className="nav-overlay open" onClick={closeMenu} />}
      <main className="main">
        <Outlet />
      </main>
    </div>
  )
}
