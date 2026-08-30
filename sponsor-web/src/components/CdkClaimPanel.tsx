import { useState } from "react"
import type { FormEvent } from "react"
import { Gift } from "lucide-react"
import { toast } from "sonner"
import { OneTimeSecret } from "@/components/OneTimeSecret"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { useCdkClaim } from "@/hooks/useCdkClaim"
import { formatYuan } from "@/lib/format"
export interface CdkClaimPanelProps { onLogin: (code: string) => Promise<boolean> }
export function CdkClaimPanel({ onLogin }: CdkClaimPanelProps) {
  const { claiming, result, claim, dismiss } = useCdkClaim()
  const [cdk, setCdk] = useState("")
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!cdk.trim()) {
      toast.error("请输入 CDK")
      return
    }
    if (await claim(cdk)) setCdk("")
  }
  if (result) {
    return (
      <Card className="h-full animate-fade-in-up">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Gift className="size-5 text-emerald-600 dark:text-emerald-400" />
            兑换成功
          </CardTitle>
          <CardDescription>{formatYuan(result.amountFen)} 已计入累计支持，登录码已生成，仅显示一次。</CardDescription>
        </CardHeader>
        <CardContent>
          <OneTimeSecret
            title="登录码"
            value={result.loginCode}
            description=""
            warning="请保存好登录码，不要分享给他人。"
            action={{ label: "立即登录", onClick: () => onLogin(result.loginCode) }}
          />
        </CardContent>
        <CardFooter>
          <Button variant="ghost" size="sm" onClick={dismiss}>兑换其它 CDK</Button>
        </CardFooter>
      </Card>
    )
  }
  return (
    <Card className="h-full transition-shadow duration-300 hover:shadow-md">
      <CardHeader>
        <CardTitle>CDK 兑换</CardTitle>
        <CardDescription>输入 CDK 兑换</CardDescription>
      </CardHeader>
      <form className="flex flex-1 flex-col" onSubmit={submit}>
        <CardContent className="flex-1">
          <label className="grid gap-2" htmlFor="cdk-code">
            <span className="text-sm font-medium">CDK</span>
            <div className="flex gap-2">
              <Input className="min-w-0 flex-1 font-mono" id="cdk-code" autoComplete="off" autoCapitalize="off" spellCheck={false} value={cdk} onChange={e => setCdk(e.target.value)} placeholder="cdk_…" />
              <Button className="shrink-0" type="submit" disabled={claiming}>{claiming ? "兑换中…" : "兑换"}</Button>
            </div>
          </label>
        </CardContent>
      </form>
    </Card>
  )
}