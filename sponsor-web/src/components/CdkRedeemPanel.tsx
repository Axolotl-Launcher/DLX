import { useState } from "react"
import { Gift, Loader2 } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { useCdk } from "@/hooks/useCdk"
import { formatYuan } from "@/lib/format"
import type { CdkRedeemResponse } from "@/lib/types"

export interface CdkRedeemPanelProps {
  onRedeemed?: (result: CdkRedeemResponse) => void
}

export function CdkRedeemPanel({ onRedeemed }: CdkRedeemPanelProps) {
  const [code, setCode] = useState("")
  const { redeeming, lastResult, redeem } = useCdk(onRedeemed)
  return (
    <Card className="animate-fade-in-up">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Gift className="size-5" />
          CDK 兑换
        </CardTitle>
        <CardDescription>兑换后金额会立即计入累计支持，帮助达成 ¥9.90 永久开通门槛。</CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4">
        <form
          className="flex flex-col gap-2 sm:flex-row"
          onSubmit={(event) => {
            event.preventDefault()
            void redeem(code)
          }}
        >
          <Input
            value={code}
            onChange={(event) => setCode(event.target.value)}
            placeholder="请输入 CDK，如 cdk_…"
            aria-label="CDK"
            autoComplete="off"
            spellCheck={false}
          />
          <Button type="submit" disabled={redeeming} className="shrink-0">
            {redeeming && <Loader2 className="size-4 animate-spin" />}
            兑换
          </Button>
        </form>
        {lastResult && (
          <p className="rounded-xl border bg-muted/40 px-3 py-2 text-sm text-muted-foreground">
            已兑换 <span className="font-medium text-foreground tabular-nums">{formatYuan(lastResult.amount_fen)}</span>，已计入累计支持。
          </p>
        )}
      </CardContent>
    </Card>
  )
}