import { useState, useEffect, useCallback } from 'react'
import type { NodeInfo } from '../api/types'
import { api } from '../api/client'

export function usePeers() {
  const [peers, setPeers] = useState<Record<string, NodeInfo>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    try {
      const data = await api.getPeers()
      setPeers(data ?? {})
      setError(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to fetch peers')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { refresh() }, [refresh])

  const nickName = useCallback((nodeName: string) => {
    return peers[nodeName]?.nick_name ?? nodeName
  }, [peers])

  return { peers, loading, error, refresh, nickName }
}
