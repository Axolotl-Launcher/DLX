import { useCallback, useEffect, useState } from "react"
import { toast } from "sonner"

import { api, errorMessage } from "@/lib/api"

export function useApiKey() {
  const [generating, setGenerating] = useState(false)
  const [revoking, setRevoking] = useState(false)
  const [pendingKey, setPendingKey] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    try { setPendingKey((await api.getApiKey()).api_key) } catch { setPendingKey(null) }
  }, [])

  useEffect(() => {
    queueMicrotask(() => void refresh())
  }, [refresh])

  const generate = useCallback(async (): Promise<boolean> => {
    setGenerating(true)
    try {
      const response = await api.generateApiKey()
      setPendingKey(response.api_key)
      toast.success("API Key 已生成，请立即复制并妥善保存")
      return true
    } catch (error) {
      toast.error(errorMessage(error))
      return false
    } finally {
      setGenerating(false)
    }
  }, [])

  const revoke = useCallback(async (): Promise<boolean> => {
    setRevoking(true)
    try {
      await api.revokeApiKey()
      setPendingKey(null)
      toast.success("API Key 已吊销")
      return true
    } catch (error) {
      toast.error(errorMessage(error))
      return false
    } finally {
      setRevoking(false)
    }
  }, [])

  return { generating, revoking, pendingKey, generate, revoke }
}