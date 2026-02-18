import { useState, useEffect, type FormEvent } from 'react'
import { usePeers } from '../hooks/usePeers'
import { api } from '../api/client'

export default function Propose() {
  const { peers, loading: peersLoading } = usePeers()
  const [nodeName, setNodeName] = useState('')
  const [to, setTo] = useState('')
  const [amount, setAmount] = useState('')
  const [title, setTitle] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  useEffect(() => {
    api.getInfo().then(info => setNodeName(info.node_name)).catch(() => {})
  }, [])

  if (peersLoading) {
    return <div className="loading">Loading...</div>
  }

  const peerList = Object.values(peers).filter(p => p.name !== nodeName)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setMessage(null)

    const amt = parseInt(amount, 10)
    if (!to || !amt || amt <= 0 || !title.trim()) return

    setSubmitting(true)
    try {
      await api.proposeTransaction({
        from: nodeName,
        to,
        amount: amt,
        title: title.trim(),
      })
      setMessage({ type: 'success', text: 'Transaction proposed successfully' })
      setTo('')
      setAmount('')
      setTitle('')
    } catch (e) {
      setMessage({ type: 'error', text: e instanceof Error ? e.message : 'Failed to propose' })
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div>
      <h1 className="page-title">Propose Transaction</h1>

      {message && (
        <div className={`alert alert-${message.type}`}>{message.text}</div>
      )}

      <div className="card" style={{ maxWidth: 480 }}>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label className="form-label">From</label>
            <input
              className="form-input"
              value={peers[nodeName]?.nick_name ?? nodeName}
              disabled
            />
          </div>

          <div className="form-group">
            <label className="form-label">To</label>
            <select
              className="form-select"
              value={to}
              onChange={e => setTo(e.target.value)}
              required
            >
              <option value="">Select recipient</option>
              {peerList.map(p => (
                <option key={p.name} value={p.name}>
                  {p.nick_name} ({p.name})
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label className="form-label">Amount</label>
            <input
              className="form-input"
              type="number"
              min="1"
              value={amount}
              onChange={e => setAmount(e.target.value)}
              placeholder="1000"
              required
            />
          </div>

          <div className="form-group">
            <label className="form-label">Title</label>
            <input
              className="form-input"
              type="text"
              maxLength={200}
              value={title}
              onChange={e => setTitle(e.target.value)}
              placeholder="e.g. Dinner split"
              required
            />
          </div>

          <button
            className="btn btn-primary"
            type="submit"
            disabled={submitting}
          >
            {submitting ? 'Submitting...' : 'Propose'}
          </button>
        </form>
      </div>
    </div>
  )
}
