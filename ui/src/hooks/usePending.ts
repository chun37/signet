import { useState, useEffect, useCallback } from 'react'
import type { PendingTransaction } from '../api/types'
import { api } from '../api/client'

export function usePending() {
  const [pending, setPending] = useState<PendingTransaction[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    try {
      const data = await api.getPending()
      setPending(data ?? [])
      setError(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to fetch pending')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { refresh() }, [refresh])

  return { pending, loading, error, refresh }
}
