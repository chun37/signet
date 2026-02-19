import { useState, useEffect, useCallback } from 'react'
import type { Block } from '@/api/types'
import { api } from '@/api/client'

export function useChain() {
  const [chain, setChain] = useState<Block[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    try {
      const data = await api.getChain()
      setChain(data ?? [])
      setError(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to fetch chain')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { refresh() }, [refresh])

  return { chain, loading, error, refresh }
}
