import { useState } from "react"
import type { FormEvent } from "react"
import { Gift } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { useCdk } from "@/hooks/useCdk"
import { formatYuan } from "@/lib/format"
import type { CdkRedeemResponse } from "@/lib/types"

const PENDING_CDK_KEY = "axl_pending_cdk"

export interface CdkRedeemPanelProps {
  onRedeemed?: (result: CdkRedeemResponse) => void
}

export function CdkRedeemPanel({ onRedeemed }: CdkRedeemPanelProps) {
  const [code, setCode] = useState(() => sessionStorage.getItem(PENDING_CDK_KEY) ?? "")
  const { redeeming, lastResult, redeem } = useCdk(onRedeemed)
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!code.trim()) {
      return
    }
    if (await redeem(code)) {
      sessionStorage.removeItem(PENDING_CDK_KEY)
      setCode("")
    }
  }
  return (
    <Card className="h-full animate-fade-in-up">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Gift className="size-5" />
          CDK 兑换
        </CardTitle>
        <CardDescription>兑换后金额会立即计入累计支持，帮助达成 ¥9.90 永久开通门槛。</CardDescription>
      </CardHeader>
      <form className="flex flex-1 flex-col" onSubmit={submit}>
        <CardContent className="flex-1">
          <label className="grid gap-2" htmlFor="cdk-code">
            <span className="text-sm font-medium">CDK</span>
            <div className="flex gap-2">
              <Input
                className="min-w-0 flex-1 font-mono"
                id="cdk-code"
                autoComplete="off"
                autoCapitalize="off"
                spellCheck={false}
                value={code}
                onChange={(event) => setCode(event.target.value)}
                placeholder="cdk_…"
              />
              <Button className="shrink-0" type="submit" disabled={redeeming}>
                {redeeming ? "兑换中…" : "兑换"}
              </Button>
            </div>
            {lastResult && (
              <p className="rounded-xl border bg-muted/40 px-3 py-2 text-sm text-muted-foreground">
                已兑换 <span className="font-medium text-foreground tabular-nums">{formatYuan(lastResult.amount_fen)}</span>，已计入累计支持。
              </p>
            )}
          </label>
        </CardContent>
      </form>
    </Card>
  )
}