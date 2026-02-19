import { useChain } from '@/hooks/useChain'
import { usePeers } from '@/hooks/usePeers'
import {
  Table, TableHeader, TableBody, TableRow, TableHead, TableCell,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'

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
    return <div className="py-8 text-center text-muted-foreground">Loading...</div>
  }

  const blocks = [...chain].reverse()

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">立替履歴</h1>

      {blocks.length === 0 ? (
        <p className="py-8 text-center text-muted-foreground">ブロックがありません</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">#</TableHead>
              <TableHead className="w-16">種別</TableHead>
              <TableHead>内容</TableHead>
              <TableHead className="text-right">金額</TableHead>
              <TableHead>日時</TableHead>
              <TableHead>ハッシュ</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {blocks.map(block => (
              <TableRow key={block.header.index}>
                <TableCell>{block.header.index}</TableCell>
                <TableCell>
                  <Badge variant={block.payload.type === 'transaction' ? 'default' : 'secondary'}>
                    {block.payload.type === 'transaction' ? '立替' : 'ノード'}
                  </Badge>
                </TableCell>
                <TableCell>
                  {block.payload.type === 'transaction' && block.payload.transaction ? (
                    <div>
                      <span className="font-medium">{nickName(block.payload.transaction.from)}</span>
                      {' \u2192 '}
                      <span className="font-medium">{nickName(block.payload.transaction.to)}</span>
                      <p className="text-sm text-muted-foreground">{block.payload.transaction.title}</p>
                    </div>
                  ) : block.payload.add_node ? (
                    <div>
                      <span className="font-medium">{block.payload.add_node.nick_name}</span>
                      <span className="text-muted-foreground"> ({block.payload.add_node.node_name})</span>
                    </div>
                  ) : '-'}
                </TableCell>
                <TableCell className="text-right font-mono font-semibold">
                  {block.payload.transaction ? `${formatAmount(block.payload.transaction.amount)} 円` : '-'}
                </TableCell>
                <TableCell className="text-nowrap">
                  {formatDate(block.header.created_at)}
                </TableCell>
                <TableCell>
                  <span className="font-mono text-xs text-muted-foreground">
                    {block.header.hash.slice(0, 12)}...
                  </span>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  )
}
