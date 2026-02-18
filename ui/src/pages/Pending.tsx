import { useState } from 'react'
import { usePending } from '../hooks/usePending'
import { usePeers } from '../hooks/usePeers'
import { api } from '../api/client'

export default function Pending() {
  const { pending, loading, refresh } = usePending()
  const { nickName, loading: peersLoading } = usePeers()
  const [acting, setActing] = useState<string | null>(null)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  if (loading || peersLoading) {
    return <div className="loading">Loading...</div>
  }

  const handleApprove = async (id: string) => {
    setActing(id)
    setMessage(null)
    try {
      await api.approveTransaction(id)
      setMessage({ type: 'success', text: 'Transaction approved' })
      await refresh()
    } catch (e) {
      setMessage({ type: 'error', text: e instanceof Error ? e.message : 'Failed to approve' })
    } finally {
      setActing(null)
    }
  }

  const handleReject = async (id: string) => {
    setActing(id)
    setMessage(null)
    try {
      await api.rejectTransaction(id)
      setMessage({ type: 'success', text: 'Transaction rejected' })
      await refresh()
    } catch (e) {
      setMessage({ type: 'error', text: e instanceof Error ? e.message : 'Failed to reject' })
    } finally {
      setActing(null)
    }
  }

  return (
    <div>
      <h1 className="page-title">Pending Transactions</h1>

      {message && (
        <div className={`alert alert-${message.type}`}>{message.text}</div>
      )}

      {pending.length === 0 ? (
        <div className="empty">No pending transactions</div>
      ) : (
        <div className="card-grid">
          {pending.map(tx => (
            <div key={tx.id} className="card">
              <div style={{ marginBottom: '0.75rem' }}>
                <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.5rem' }}>
                  <span className="hash">{tx.id.slice(0, 16)}...</span>
                </div>
                <div style={{ fontSize: '1.1rem', marginBottom: '0.25rem' }}>
                  <strong>{nickName(tx.transaction.from)}</strong>
                  {' \u2192 '}
                  <strong>{nickName(tx.transaction.to)}</strong>
                </div>
                <div className="amount" style={{ fontSize: '1.5rem', margin: '0.5rem 0' }}>
                  {tx.transaction.amount.toLocaleString('ja-JP')}
                </div>
                <div style={{ color: 'var(--text-secondary)' }}>
                  {tx.transaction.title}
                </div>
              </div>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <button
                  className="btn btn-success"
                  disabled={acting === tx.id}
                  onClick={() => handleApprove(tx.id)}
                >
                  Approve
                </button>
                <button
                  className="btn btn-danger"
                  disabled={acting === tx.id}
                  onClick={() => handleReject(tx.id)}
                >
                  Reject
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
