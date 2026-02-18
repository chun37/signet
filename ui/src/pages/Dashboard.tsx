import { useChain } from '../hooks/useChain'
import { usePeers } from '../hooks/usePeers'
import { usePending } from '../hooks/usePending'
import BalanceSummary from '../components/BalanceSummary'

export default function Dashboard() {
  const { chain, loading: chainLoading } = useChain()
  const { peers, loading: peersLoading } = usePeers()
  const { pending, loading: pendingLoading } = usePending()

  if (chainLoading || peersLoading || pendingLoading) {
    return <div className="loading">Loading...</div>
  }

  const txCount = chain.filter(b => b.payload.type === 'transaction').length
  const nodeCount = Object.keys(peers).length

  return (
    <div>
      <h1 className="page-title">Dashboard</h1>

      <div className="stat-grid">
        <div className="stat-card">
          <div className="stat-label">Blocks</div>
          <div className="stat-value">{chain.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Transactions</div>
          <div className="stat-value">{txCount}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Pending</div>
          <div className="stat-value">{pending.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Nodes</div>
          <div className="stat-value">{nodeCount}</div>
        </div>
      </div>

      <div className="card">
        <h2 style={{ fontSize: '1.1rem', marginBottom: '1rem' }}>Balance Summary</h2>
        <BalanceSummary chain={chain} peers={peers} />
      </div>
    </div>
  )
}
