import { useState, useEffect } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { Badge } from '@/components/ui/badge'
import { api } from '@/api/client'

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

  const linkClass = ({ isActive }: { isActive: boolean }) =>
    `px-3 py-2 rounded-md text-sm transition-colors ${
      isActive
        ? 'text-foreground bg-accent'
        : 'text-muted-foreground hover:text-foreground hover:bg-accent'
    }`

  return (
    <div className="flex min-h-screen flex-col">
      <header className="sticky top-0 z-50 border-b bg-card">
        <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4">
          <span className="text-lg font-bold tracking-tight">Signet</span>
          <button
            className="text-xl sm:hidden"
            onClick={() => setMenuOpen(!menuOpen)}
          >
            {menuOpen ? '\u2715' : '\u2630'}
          </button>
          <nav className={`${menuOpen ? 'flex' : 'hidden'} fixed inset-x-0 top-14 bottom-0 z-50 flex-col gap-1 border-t bg-card p-4 sm:static sm:flex sm:flex-row sm:border-0 sm:p-0`}>
            <NavLink to="/" end className={linkClass} onClick={closeMenu}>ダッシュボード</NavLink>
            <NavLink to="/transactions" className={linkClass} onClick={closeMenu}>立替履歴</NavLink>
            <NavLink to="/pending" className={linkClass} onClick={closeMenu}>
              承認待ち
              {pendingCount > 0 && (
                <Badge variant="destructive" className="ml-1.5 text-[10px] px-1.5 py-0">
                  {pendingCount}
                </Badge>
              )}
            </NavLink>
            <NavLink to="/propose" className={linkClass} onClick={closeMenu}>立替を記録</NavLink>
            <NavLink to="/nodes" className={linkClass} onClick={closeMenu}>ノード</NavLink>
          </nav>
        </div>
      </header>
      {menuOpen && <div className="fixed inset-0 top-14 z-40 sm:hidden" onClick={closeMenu} />}
      <main className="mx-auto w-full max-w-5xl flex-1 p-4">
        <Outlet />
      </main>
    </div>
  )
}
