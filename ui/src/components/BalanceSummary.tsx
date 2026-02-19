import type { Block, NodeInfo } from '@/api/types'

function formatAmount(n: number): string {
  return n.toLocaleString('ja-JP')
}

interface Props {
  chain: Block[]
  peers: Record<string, NodeInfo>
  myNodeName: string
}

interface SettlementEntry {
  peer: string
  amount: number // 正=相手に貸し(受け取り), 負=相手に借り(支払い)
}

export default function BalanceSummary({ chain, peers, myNodeName }: Props) {
  const balanceMap = new Map<string, number>()

  for (const block of chain) {
    if (block.payload.type !== 'transaction' || !block.payload.transaction) continue
    const { from, to, amount } = block.payload.transaction

    if (from === myNodeName) {
      balanceMap.set(to, (balanceMap.get(to) ?? 0) + amount)
    } else if (to === myNodeName) {
      balanceMap.set(from, (balanceMap.get(from) ?? 0) - amount)
    }
  }

  const settlements: SettlementEntry[] = []
  for (const [peer, net] of balanceMap) {
    if (net === 0) continue
    settlements.push({ peer, amount: net })
  }
  settlements.sort((a, b) => b.amount - a.amount)

  const nick = (name: string) => peers[name]?.nick_name ?? name

  if (settlements.length === 0) {
    return <p className="py-8 text-center text-muted-foreground">精算データがありません</p>
  }

  const total = settlements.reduce((sum, s) => sum + s.amount, 0)

  return (
    <div className="space-y-1">
      {settlements.map((s, i) => (
        <div key={i} className="flex items-center justify-between py-3 border-b last:border-0">
          <span className="font-medium">{nick(s.peer)}</span>
          {s.amount > 0 ? (
            <span className="font-mono font-semibold text-green-500">
              +{formatAmount(s.amount)} 円（受け取り）
            </span>
          ) : (
            <span className="font-mono font-semibold text-red-500">
              -{formatAmount(Math.abs(s.amount))} 円（支払い）
            </span>
          )}
        </div>
      ))}
      {settlements.length > 1 && (
        <div className="flex items-center justify-between pt-3 border-t">
          <span className="font-medium text-muted-foreground">合計</span>
          <span className={`font-mono font-bold ${total > 0 ? 'text-green-500' : total < 0 ? 'text-red-500' : 'text-muted-foreground'}`}>
            {total > 0 ? '+' : total < 0 ? '-' : ''}{formatAmount(Math.abs(total))} 円
          </span>
        </div>
      )}
    </div>
  )
}
