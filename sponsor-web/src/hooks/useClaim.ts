import { useCallback, useState } from "react"
import { toast } from "sonner"

import { api, errorMessage } from "@/lib/api"

export function useClaim() {
  const [claiming, setClaiming] = useState(false)
  const [loginCode, setLoginCode] = useState<string | null>(null)

  const claim = useCallback(async (orderNo: string): Promise<boolean> => {
    setClaiming(true)
    try {
      const response = await api.claim(orderNo.trim())
      setLoginCode(response.login_code)
      return true
    } catch (error) {
      toast.error(errorMessage(error))
      return false
    } finally {
      setClaiming(false)
    }
  }, [])

  const dismiss = useCallback(() => setLoginCode(null), [])

  return { claiming, loginCode, claim, dismiss }
}