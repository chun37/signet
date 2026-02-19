import { useState, useEffect, type FormEvent } from 'react'
import { usePeers } from '@/hooks/usePeers'
import { api } from '@/api/client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Alert, AlertDescription } from '@/components/ui/alert'

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
    return <div className="py-8 text-center text-muted-foreground">Loading...</div>
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
      setMessage({ type: 'success', text: '立替を記録しました' })
      setTo('')
      setAmount('')
      setTitle('')
    } catch (e) {
      setMessage({ type: 'error', text: e instanceof Error ? e.message : '記録に失敗しました' })
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">立替を記録</h1>

      {message && (
        <Alert variant={message.type === 'error' ? 'destructive' : 'default'}>
          <AlertDescription>{message.text}</AlertDescription>
        </Alert>
      )}

      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>新しい立替</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-sm text-muted-foreground">立替えた人</label>
              <Input value={peers[nodeName]?.nick_name ?? nodeName} disabled />
            </div>

            <div className="space-y-1.5">
              <label className="text-sm text-muted-foreground">請求先</label>
              <select
                className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                value={to}
                onChange={e => setTo(e.target.value)}
                required
              >
                <option value="">請求先を選択</option>
                {peerList.map(p => (
                  <option key={p.name} value={p.name}>
                    {p.nick_name} ({p.name})
                  </option>
                ))}
              </select>
            </div>

            <div className="space-y-1.5">
              <label className="text-sm text-muted-foreground">金額</label>
              <Input
                type="number"
                min="1"
                value={amount}
                onChange={e => setAmount(e.target.value)}
                placeholder="1000"
                required
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-sm text-muted-foreground">内容</label>
              <Input
                type="text"
                maxLength={200}
                value={title}
                onChange={e => setTitle(e.target.value)}
                placeholder="例: 飲み会代"
                required
              />
            </div>

            <Button type="submit" disabled={submitting} className="w-full">
              {submitting ? '送信中...' : '送信'}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
