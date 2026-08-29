import { Info, ListChecks, ShieldCheck } from "lucide-react"

import { buttonVariants } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { SPONSOR_URL } from "@/lib/constants"
import { formatYuan, remainingFen } from "@/lib/format"
import type { EntitlementStatus, MeResponse } from "@/lib/types"
import { cn } from "@/lib/utils"

const STATUS_META: Record<
  EntitlementStatus,
  { label: string; className: string }
> = {
  granted: {
    label: "已永久开通",
    className: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400",
  },
  pending: {
    label: "尚未达标",
    className: "bg-amber-500/10 text-amber-600 dark:text-amber-400",
  },
  manual_review: {
    label: "人工审核中",
    className: "bg-sky-500/10 text-sky-600 dark:text-sky-400",
  },
  suspended: {
    label: "已暂停",
    className: "bg-destructive/10 text-destructive",
  },
}

export interface StatusOverviewProps {
  me: MeResponse
}

export function StatusOverview({ me }: StatusOverviewProps) {
  const status = STATUS_META[me.entitlement_status]
  const remaining = remainingFen(me.lifetime_paid_fen, me.threshold_fen)
  const percent =
    me.threshold_fen > 0
      ? Math.min(100, Math.round((me.lifetime_paid_fen / me.threshold_fen) * 100))
      : 0

  const body =
    me.entitlement_status === "granted"
      ? "感谢你的支持！累计支持满 ¥9.90，Axolotl Sponsor Translate 已永久开通，不设月度到期。"
      : me.entitlement_status === "manual_review"
        ? "账户正在人工审核中，请耐心等待；如有疑问可通过爱发电私信联系作者。"
        : me.entitlement_status === "suspended"
          ? "你的永久资格已被暂停，如有疑问请通过爱发电私信联系作者。"
          : `当前累计支持 ${formatYuan(me.lifetime_paid_fen)}，距离永久开通还差 ${formatYuan(remaining)}。`

  return (
    <div className="grid gap-4 md:grid-cols-3">
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center justify-between gap-2">
            <span className="flex items-center gap-2">
              <ShieldCheck className="size-5" />
              永久资格
            </span>
            <span
              className={cn(
                "rounded-full px-2.5 py-0.5 text-xs font-medium",
                status.className
              )}
            >
              {status.label}
            </span>
          </CardTitle>
          <CardDescription>
            按累计实际成功支付金额计算，不设月度到期。
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4">
          <p className="text-sm">{body}</p>
          <div className="grid gap-1.5">
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <span>
                累计支持 {formatYuan(me.lifetime_paid_fen)} / 门槛{" "}
                {formatYuan(me.threshold_fen)}
              </span>
              <span>{percent}%</span>
            </div>
            <div
              className="h-2 overflow-hidden rounded-full bg-muted"
              role="progressbar"
              aria-valuenow={percent}
              aria-valuemin={0}
              aria-valuemax={100}
              aria-label="累计支持进度"
            >
              <div
                className="h-full rounded-full bg-primary transition-all"
                style={{ width: `${percent}%` }}
              />
            </div>
          </div>
          {me.entitlement_status === "pending" && (
            <a
              className={cn(buttonVariants({ variant: "outline" }), "w-fit")}
              href={SPONSOR_URL}
              target="_blank"
              rel="noreferrer"
            >
              前往爱发电支持
            </a>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Info className="size-5" />
            使用与隐私
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="grid gap-2 text-sm text-muted-foreground">
            <li>本页面暂不提供用量统计展示。</li>
            <li>服务端不保存翻译正文，仅记录无原文的日聚合用量。</li>
            <li>完整 API Key 只在生成时展示一次，服务端仅保存不可逆哈希。</li>
            <li>遇到问题请通过爱发电私信联系作者。</li>
          </ul>
        </CardContent>
      </Card>

      <Card className="md:col-span-3">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ListChecks className="size-5" />
            权益说明
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ul className="grid gap-2 text-sm text-muted-foreground md:grid-cols-2">
            <li>· 按你名下全部成功支付订单的实付金额累计计算。</li>
            <li>· 达到 ¥9.90 即永久开通，不设月度额度或到期。</li>
            <li>· 退款 / 撤销会重算净累计，低于门槛时暂停权限。</li>
            <li>· 每账户当前最多持有 1 个有效 Key，轮换后旧 Key 立即失效。</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  )
}