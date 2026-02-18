import type { Block, NodeInfo } from '../api/types'

function formatAmount(n: number): string {
  return n.toLocaleString('ja-JP')
}

interface Props {
  chain: Block[]
  peers: Record<string, NodeInfo>
}

interface Balance {
  from: string
  to: string
  amount: number
}

export default function BalanceSummary({ chain, peers }: Props) {
  // Calculate net balances between each pair
  const pairMap = new Map<string, number>()

  for (const block of chain) {
    if (block.payload.type !== 'transaction' || !block.payload.transaction) continue
    const { from, to, amount } = block.payload.transaction
    // Canonical key: sorted pair
    const key = from < to ? `${from}|${to}` : `${to}|${from}`
    const sign = from < to ? 1 : -1
    pairMap.set(key, (pairMap.get(key) ?? 0) + amount * sign)
  }

  const balances: Balance[] = []
  for (const [key, net] of pairMap) {
    if (net === 0) continue
    const [a, b] = key.split('|')
    // Positive means a owes b (a -> b transactions exceeded b -> a)
    if (net > 0) {
      balances.push({ from: a, to: b, amount: net })
    } else {
      balances.push({ from: b, to: a, amount: -net })
    }
  }

  const nick = (name: string) => peers[name]?.nick_name ?? name

  if (balances.length === 0) {
    return <p className="empty">No balances to display</p>
  }

  return (
    <div className="table-wrap">
      <table className="balance-table">
        <thead>
          <tr>
            <th>From</th>
            <th>To</th>
            <th style={{ textAlign: 'right' }}>Net Amount</th>
          </tr>
        </thead>
        <tbody>
          {balances.map((b, i) => (
            <tr key={i}>
              <td>{nick(b.from)}</td>
              <td>{nick(b.to)}</td>
              <td style={{ textAlign: 'right' }}>
                <span className="amount">{formatAmount(b.amount)}</span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
