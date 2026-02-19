import { usePeers } from '@/hooks/usePeers'
import { Card, CardContent } from '@/components/ui/card'

export default function Nodes() {
  const { peers, loading } = usePeers()

  if (loading) {
    return <div className="py-8 text-center text-muted-foreground">Loading...</div>
  }

  const nodes = Object.values(peers)

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Nodes</h1>

      {nodes.length === 0 ? (
        <p className="py-8 text-center text-muted-foreground">No nodes registered</p>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {nodes.map(node => (
            <Card key={node.name}>
              <CardContent className="space-y-2 pt-4 pb-4">
                <p className="text-lg font-bold">{node.nick_name}</p>
                <p className="text-sm text-muted-foreground">{node.name}</p>
                <div className="space-y-1 text-sm text-muted-foreground">
                  <p>Address: <span className="text-foreground">{node.address}</span></p>
                  <p>Public Key: <span className="font-mono text-xs">{node.public_key.slice(0, 16)}...</span></p>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
