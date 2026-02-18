import { useChain } from '../hooks/useChain'
import { usePeers } from '../hooks/usePeers'

function formatDate(unix: number): string {
  if (unix === 0) return '-'
  return new Date(unix * 1000).toLocaleString('ja-JP')
}

function formatAmount(n: number): string {
  return n.toLocaleString('ja-JP')
}

export default function Transactions() {
  const { chain, loading: chainLoading } = useChain()
  const { loading: peersLoading, nickName } = usePeers()

  if (chainLoading || peersLoading) {
    return <div className="loading">Loading...</div>
  }

  // Reverse to show newest first, skip genesis
  const blocks = [...chain].reverse()

  return (
    <div>
      <h1 className="page-title">Transactions</h1>

      {blocks.length === 0 ? (
        <div className="empty">No blocks in chain</div>
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>#</th>
                <th>Type</th>
                <th>Details</th>
                <th style={{ textAlign: 'right' }}>Amount</th>
                <th>Date</th>
                <th>Hash</th>
              </tr>
            </thead>
            <tbody>
              {blocks.map(block => (
                <tr key={block.header.index}>
                  <td>{block.header.index}</td>
                  <td>
                    {block.payload.type === 'transaction' ? 'TX' : 'Node'}
                  </td>
                  <td>
                    {block.payload.type === 'transaction' && block.payload.transaction ? (
                      <>
                        <strong>{nickName(block.payload.transaction.from)}</strong>
                        {' \u2192 '}
                        <strong>{nickName(block.payload.transaction.to)}</strong>
                        <br />
                        <span style={{ color: 'var(--text-secondary)', fontSize: '0.85rem' }}>
                          {block.payload.transaction.title}
                        </span>
                      </>
                    ) : block.payload.add_node ? (
                      <>
                        <strong>{block.payload.add_node.nick_name}</strong>
                        <span style={{ color: 'var(--text-secondary)' }}>
                          {' '}({block.payload.add_node.node_name})
                        </span>
                      </>
                    ) : '-'}
                  </td>
                  <td style={{ textAlign: 'right' }}>
                    {block.payload.transaction ? (
                      <span className="amount">
                        {formatAmount(block.payload.transaction.amount)}
                      </span>
                    ) : '-'}
                  </td>
                  <td style={{ whiteSpace: 'nowrap' }}>
                    {formatDate(block.header.created_at)}
                  </td>
                  <td>
                    <span className="hash">{block.header.hash.slice(0, 12)}...</span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
