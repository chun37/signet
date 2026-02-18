import { usePeers } from '../hooks/usePeers'

export default function Nodes() {
  const { peers, loading } = usePeers()

  if (loading) {
    return <div className="loading">Loading...</div>
  }

  const nodes = Object.values(peers)

  return (
    <div>
      <h1 className="page-title">Nodes</h1>

      {nodes.length === 0 ? (
        <div className="empty">No nodes registered</div>
      ) : (
        <div className="card-grid">
          {nodes.map(node => (
            <div key={node.name} className="card">
              <div style={{ fontSize: '1.1rem', fontWeight: 700, marginBottom: '0.25rem' }}>
                {node.nick_name}
              </div>
              <div style={{ color: 'var(--text-secondary)', marginBottom: '0.75rem' }}>
                {node.name}
              </div>
              <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
                <div style={{ marginBottom: '0.25rem' }}>
                  Address: <span style={{ color: 'var(--text-primary)' }}>{node.address}</span>
                </div>
                <div>
                  Public Key: <span className="hash">{node.public_key.slice(0, 16)}...</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
