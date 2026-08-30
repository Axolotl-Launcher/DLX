import { useCallback, useState } from "react"
import { toast } from "sonner"

import { api, errorMessage } from "@/lib/api"
import type { CdkRedeemResponse } from "@/lib/types"

export function useCdk(onRedeemed?: (result: CdkRedeemResponse) => void) {
  const [redeeming, setRedeeming] = useState(false)
  const [lastResult, setLastResult] = useState<CdkRedeemResponse | null>(null)

  const redeem = useCallback(
    async (code: string): Promise<boolean> => {
      const cdk = code.trim()
      if (!cdk) {
        toast.error("请输入 CDK")
        return false
      }
      setRedeeming(true)
      try {
        const result = await api.redeemCdk(cdk)
        setLastResult(result)
        toast.success("兑换成功，金额已计入累计支持")
        onRedeemed?.(result)
        return true
      } catch (error) {
        toast.error(errorMessage(error))
        return false
      } finally {
        setRedeeming(false)
      }
    },
    [onRedeemed]
  )

  return { redeeming, lastResult, redeem }
}