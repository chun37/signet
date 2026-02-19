import { useState } from 'react'
import { usePending } from '@/hooks/usePending'
import { usePeers } from '@/hooks/usePeers'
import { api } from '@/api/client'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'

export default function Pending() {
  const { pending, loading, refresh } = usePending()
  const { nickName, loading: peersLoading } = usePeers()
  const [acting, setActing] = useState<string | null>(null)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  if (loading || peersLoading) {
    return <div className="py-8 text-center text-muted-foreground">Loading...</div>
  }

  const handleApprove = async (id: string) => {
    setActing(id)
    setMessage(null)
    try {
      await api.approveTransaction(id)
      setMessage({ type: 'success', text: '承認しました' })
      await refresh()
    } catch (e) {
      setMessage({ type: 'error', text: e instanceof Error ? e.message : '承認に失敗しました' })
    } finally {
      setActing(null)
    }
  }

  const handleReject = async (id: string) => {
    setActing(id)
    setMessage(null)
    try {
      await api.rejectTransaction(id)
      setMessage({ type: 'success', text: '却下しました' })
      await refresh()
    } catch (e) {
      setMessage({ type: 'error', text: e instanceof Error ? e.message : '却下に失敗しました' })
    } finally {
      setActing(null)
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">承認待ちの立替</h1>

      {message && (
        <Alert variant={message.type === 'error' ? 'destructive' : 'default'}>
          <AlertDescription>{message.text}</AlertDescription>
        </Alert>
      )}

      {pending.length === 0 ? (
        <p className="py-8 text-center text-muted-foreground">承認待ちの立替はありません</p>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {pending.map(tx => (
            <Card key={tx.id}>
              <CardContent className="space-y-3 pt-4 pb-4">
                <p className="font-mono text-xs text-muted-foreground">
                  {tx.id.slice(0, 16)}...
                </p>
                <p className="text-lg">
                  <span className="font-medium">{nickName(tx.transaction.from)}</span>
                  {' \u2192 '}
                  <span className="font-medium">{nickName(tx.transaction.to)}</span>
                </p>
                <p className="font-mono text-2xl font-bold">
                  {tx.transaction.amount.toLocaleString('ja-JP')} 円
                </p>
                <p className="text-sm text-muted-foreground">{tx.transaction.title}</p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    className="border-green-600 text-green-500 hover:bg-green-500/10"
                    disabled={acting === tx.id}
                    onClick={() => handleApprove(tx.id)}
                  >
                    承認
                  </Button>
                  <Button
                    variant="destructive"
                    disabled={acting === tx.id}
                    onClick={() => handleReject(tx.id)}
                  >
                    却下
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
