import type { Block, PendingTransaction, NodeInfo, InfoResponse, ProposeRequest } from './types'

async function fetchJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init)
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error((body as { error?: string }).error || res.statusText)
  }
  return res.json()
}

export const api = {
  getChain: () => fetchJSON<Block[]>('/chain'),
  getPeers: () => fetchJSON<Record<string, NodeInfo>>('/peers'),
  getPending: () => fetchJSON<PendingTransaction[]>('/transaction/pending'),
  getInfo: () => fetchJSON<InfoResponse>('/info'),

  proposeTransaction: (data: ProposeRequest) =>
    fetchJSON<{ status: string; message: string }>('/transaction/propose', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    }),

  approveTransaction: (id: string) =>
    fetchJSON<{ status: string; block: Block }>('/transaction/approve', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id }),
    }),

  rejectTransaction: (id: string) =>
    fetchJSON<{ status: string; message: string }>('/transaction/reject', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id }),
    }),
}
