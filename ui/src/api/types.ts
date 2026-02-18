export interface BlockHeader {
  index: number
  created_at: number
  prev_hash: string
  hash: string
}

export interface TransactionData {
  from: string
  to: string
  amount: number
  title: string
}

export interface AddNodeData {
  public_key: string
  node_name: string
  nick_name: string
  address: string
}

export interface BlockPayload {
  type: 'transaction' | 'add_node'
  transaction?: TransactionData
  add_node?: AddNodeData
  from_signature: string
  to_signature: string
}

export interface Block {
  header: BlockHeader
  payload: BlockPayload
}

export interface PendingTransaction {
  id: string
  from_sig: string
  transaction: TransactionData
}

export interface NodeInfo {
  name: string
  nick_name: string
  address: string
  public_key: string
}

export interface InfoResponse {
  node_name: string
}

export interface ProposeRequest {
  from: string
  to: string
  amount: number
  title: string
}
