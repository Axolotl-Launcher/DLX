import { ShieldCheck } from "lucide-react"
import { UsageHeatmap } from "@/components/UsageHeatmap"
import { buttonVariants } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { SPONSOR_URL } from "@/lib/constants"
import { formatYuan, remainingFen } from "@/lib/format"
import type { EntitlementStatus, MeResponse } from "@/lib/types"
import { cn } from "@/lib/utils"

const STATUS_META:Record<EntitlementStatus,{label:string;className:string}>={granted:{label:"已永久开通",className:"bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"},pending:{label:"尚未达标",className:"bg-amber-500/10 text-amber-600 dark:text-amber-400"},manual_review:{label:"人工审核中",className:"bg-sky-500/10 text-sky-600 dark:text-sky-400"},suspended:{label:"已暂停",className:"bg-destructive/10 text-destructive"}}
export interface StatusOverviewProps{me:MeResponse}
export function StatusOverview({me}:StatusOverviewProps){
 const status=STATUS_META[me.entitlement_status];const remaining=remainingFen(me.lifetime_paid_fen,me.threshold_fen);const percent=me.threshold_fen>0?Math.min(100,Math.round(me.lifetime_paid_fen/me.threshold_fen*100)):0
 return <div className="grid gap-6">
  <div className="grid gap-4 md:grid-cols-3">
   <Card className="md:col-span-2"><CardHeader><CardTitle className="flex items-center justify-between gap-3"><span className="flex items-center gap-2"><ShieldCheck className="size-5"/>永久资格</span><span className={cn("rounded-full px-2.5 py-0.5 text-xs font-medium",status.className)}>{status.label}</span></CardTitle></CardHeader><CardContent className="grid gap-5"><div className="flex items-end justify-between gap-4"><div><p className="text-3xl font-semibold tracking-tight">{formatYuan(me.lifetime_paid_fen)}</p><p className="mt-1 text-xs text-muted-foreground">累计支持</p></div><p className="text-sm text-muted-foreground">门槛 {formatYuan(me.threshold_fen)}</p></div><div className="grid gap-2"><div className="flex justify-between text-xs text-muted-foreground"><span>开通进度</span><span>{percent}%</span></div><div className="h-2 overflow-hidden rounded-full bg-muted" role="progressbar" aria-valuenow={percent} aria-valuemin={0} aria-valuemax={100} aria-label="累计支持进度"><div className="h-full rounded-full bg-primary transition-all duration-500" style={{width:`${percent}%`}}/></div></div>{me.entitlement_status==="pending"&&<a className={cn(buttonVariants({variant:"outline"}),"w-fit")} href={SPONSOR_URL} target="_blank" rel="noreferrer">前往支持</a>}{me.entitlement_status!=="granted"&&<p className="text-xs text-muted-foreground">{me.entitlement_status==="pending"?`还差 ${formatYuan(remaining)}`:me.entitlement_status==="manual_review"?"等待审核结果":"资格暂时不可用"}</p>}</CardContent></Card>
   <Card><CardHeader><CardTitle>账户</CardTitle></CardHeader><CardContent className="grid gap-4 text-sm"><div><p className="text-xs text-muted-foreground">永久有效</p><p className="mt-1 font-medium">Sponsor Translate</p></div><div><p className="text-xs text-muted-foreground">API Key</p><p className="mt-1 font-medium">在 API Key 中管理</p></div></CardContent></Card>
  </div>
  <UsageHeatmap/>
 </div>
}
