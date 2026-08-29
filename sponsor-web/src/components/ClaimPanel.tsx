import { useState } from "react"
import type { FormEvent } from "react"
import { ShieldCheck } from "lucide-react"
import { toast } from "sonner"

import { OneTimeSecret } from "@/components/OneTimeSecret"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { SPONSOR_URL } from "@/lib/constants"

export interface ClaimPanelProps {
  claiming: boolean
  loginCode: string | null
  onClaim: (orderNo: string) => Promise<boolean>
  onDismiss: () => void
  onLogin: (code: string) => Promise<boolean>
}

export function ClaimPanel({
  claiming,
  loginCode,
  onClaim,
  onDismiss,
  onLogin,
}: ClaimPanelProps) {
  const [orderNo, setOrderNo] = useState("")

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!orderNo.trim()) {
      toast.error("请输入爱发电订单号")
      return
    }
    const ok = await onClaim(orderNo)
    if (ok) {
      setOrderNo("")
    }
  }

  if (loginCode) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ShieldCheck className="size-5 text-emerald-600 dark:text-emerald-400" />
            认领成功
          </CardTitle>
          <CardDescription>订单已核验，已为你生成一次性登录码。</CardDescription>
        </CardHeader>
        <CardContent>
          <OneTimeSecret
            title="登录码"
            value={loginCode}
            description="请立即复制并妥善保存，此后登录赞助者中心都需要用到它。"
            warning="登录码等同于账户凭据，请勿分享给他人；服务端无法再次显示完整登录码。"
            action={{ label: "使用此登录码登录", onClick: () => onLogin(loginCode) }}
          />
        </CardContent>
        <CardFooter>
          <Button variant="ghost" onClick={onDismiss}>
            重新提交其它订单
          </Button>
        </CardFooter>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>认领订单</CardTitle>
        <CardDescription>
          提交一笔属于你的爱发电订单号，系统将核验并关联同一爱发电身份的全部订单。
        </CardDescription>
      </CardHeader>
      <form onSubmit={submit}>
        <CardContent className="grid gap-3">
          <label className="grid gap-1.5" htmlFor="claim-order-no">
            <span className="font-medium">订单号</span>
            <Input
              id="claim-order-no"
              value={orderNo}
              onChange={(event) => setOrderNo(event.target.value)}
              placeholder="输入爱发电订单号"
              aria-describedby="claim-order-help"
              autoComplete="off"
            />
          </label>
          <p id="claim-order-help" className="text-xs text-muted-foreground">
            订单号仅用于核验与绑定你的爱发电身份，系统会汇总该身份下的全部订单计算累计金额。
            还没有订单？
            <a
              className="ml-1 font-medium underline underline-offset-4"
              href={SPONSOR_URL}
              target="_blank"
              rel="noreferrer"
            >
              前往爱发电支持
            </a>
            。
          </p>
        </CardContent>
        <CardFooter>
          <Button type="submit" disabled={claiming}>
            提交并核验
          </Button>
        </CardFooter>
      </form>
    </Card>
  )
}