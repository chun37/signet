import { useState, useEffect } from 'react'
import { useChain } from '@/hooks/useChain'
import { usePeers } from '@/hooks/usePeers'
import { usePending } from '@/hooks/usePending'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import BalanceSummary from '@/components/BalanceSummary'
import { api } from '@/api/client'

export default function Dashboard() {
  const { chain, loading: chainLoading } = useChain()
  const { peers, loading: peersLoading } = usePeers()
  const { pending, loading: pendingLoading } = usePending()
  const [myNodeName, setMyNodeName] = useState('')

  useEffect(() => {
    api.getInfo().then(info => setMyNodeName(info.node_name)).catch(() => {})
  }, [])

  if (chainLoading || peersLoading || pendingLoading) {
    return <div className="py-8 text-center text-muted-foreground">Loading...</div>
  }

  const txCount = chain.filter(b => b.payload.type === 'transaction').length
  const nodeCount = Object.keys(peers).length

  const stats = [
    { label: 'ブロック数', value: chain.length },
    { label: '立替件数', value: txCount },
    { label: '承認待ち', value: pending.length },
    { label: 'ノード数', value: nodeCount },
  ]

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">ダッシュボード</h1>

      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        {stats.map(s => (
          <Card key={s.label}>
            <CardContent className="pt-4 pb-4">
              <p className="text-xs uppercase tracking-wide text-muted-foreground">{s.label}</p>
              <p className="text-2xl font-bold">{s.value}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>精算サマリー</CardTitle>
        </CardHeader>
        <CardContent>
          <BalanceSummary chain={chain} peers={peers} myNodeName={myNodeName} />
        </CardContent>
      </Card>
    </div>
  )
}
