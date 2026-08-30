import { useCallback, useState } from "react"
import { toast } from "sonner"

import { api, ApiError, errorMessage } from "@/lib/api"
import type { CdkRedeemResponse } from "@/lib/types"

const PENDING_CDK_KEY = "axl_pending_cdk"

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
        if (error instanceof ApiError && error.code === "UNAUTHENTICATED") {
          // CDK 兑换需要登录：保留输入，登录后可直接继续。
          sessionStorage.setItem(PENDING_CDK_KEY, cdk)
          toast.error("兑换 CDK 需要先登录，登录后可直接继续兑换")
        } else {
          toast.error(errorMessage(error))
        }
        return false
      } finally {
        setRedeeming(false)
      }
    },
    [onRedeemed]
  )

  return { redeeming, lastResult, redeem }
}