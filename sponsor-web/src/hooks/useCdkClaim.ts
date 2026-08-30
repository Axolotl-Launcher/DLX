import { useCallback, useState } from "react"
import { toast } from "sonner"

import { api, errorMessage } from "@/lib/api"

export interface CdkClaimResult {
  loginCode: string
  amountFen: number
}

export function useCdkClaim() {
  const [claiming, setClaiming] = useState(false)
  const [result, setResult] = useState<CdkClaimResult | null>(null)

  const claim = useCallback(async (cdk: string): Promise<boolean> => {
    setClaiming(true)
    try {
      const response = await api.claimCdk(cdk.trim())
      setResult({ loginCode: response.login_code, amountFen: response.amount_fen })
      return true
    } catch (error) {
      toast.error(errorMessage(error))
      return false
    } finally {
      setClaiming(false)
    }
  }, [])

  const dismiss = useCallback(() => setResult(null), [])

  return { claiming, result, claim, dismiss }
}