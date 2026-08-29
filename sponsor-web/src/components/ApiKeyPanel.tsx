import { KeyRound, ShieldAlert, ExternalLink } from "lucide-react"
import { OneTimeSecret } from "@/components/OneTimeSecret"
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from "@/components/ui/alert-dialog"
import { Button, buttonVariants } from "@/components/ui/button"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { useApiKey } from "@/hooks/useApiKey"
import { SPONSOR_URL } from "@/lib/constants"
import { formatYuan, remainingFen } from "@/lib/format"
import type { EntitlementStatus, MeResponse } from "@/lib/types"
import { cn } from "@/lib/utils"

const LOCKED:Record<Exclude<EntitlementStatus,"granted">,string>={pending:"尚未达到开通门槛",manual_review:"账户正在审核",suspended:"资格已暂停"}
export interface ApiKeyPanelProps { me:MeResponse }

export function ApiKeyPanel({me}:ApiKeyPanelProps){
  const{generating,revoking,pendingKey,generate,revoke}=useApiKey()
  if(me.entitlement_status!=="granted"){
    const remaining=remainingFen(me.lifetime_paid_fen,me.threshold_fen)
    return <Card className="animate-fade-in-up"><CardHeader><CardTitle className="flex items-center gap-2"><ShieldAlert className="size-5"/>{LOCKED[me.entitlement_status]}</CardTitle></CardHeader><CardContent className="grid gap-4">{me.entitlement_status==="pending"&&<><p className="text-sm text-muted-foreground">还差 {formatYuan(remaining)}。</p><a className={cn(buttonVariants({variant:"outline"}),"w-fit")} href={SPONSOR_URL} target="_blank" rel="noreferrer">前往支持</a></>}{me.entitlement_status!=="pending"&&<p className="text-sm text-muted-foreground">请稍后再试。</p>}</CardContent></Card>
  }
  const importUrl=pendingKey?"axolotl://import/translation?endpoint="+encodeURIComponent("https://api.axlmc.org/v1/translate")+"&api_key="+encodeURIComponent(pendingKey):null
  return <Card className="animate-fade-in-up"><CardHeader><CardTitle className="flex items-center gap-2"><KeyRound className="size-5"/>API Key 管理</CardTitle><p className="text-sm text-muted-foreground">Key 已安全保存，可随时查看、复制或导入 Launcher。</p></CardHeader><CardContent className="grid gap-5">{pendingKey?<OneTimeSecret title="当前 API Key" value={pendingKey} description="这是当前生效的 Key。吊销或轮换后立即失效。" warning="请勿分享给他人。"/>:<p className="rounded-xl border border-dashed p-4 text-sm text-muted-foreground">当前没有生效的 Key，请生成一个新的。</p>}<div className="rounded-2xl border bg-muted/40 p-4"><p className="text-sm font-medium">Launcher 配置</p><pre className="mt-3 overflow-x-auto rounded-xl bg-background/80 p-3 font-mono text-xs leading-6">{"端点   https://api.axlmc.org/v1/translate\n认证   Authorization: Bearer <你的 API Key>"}</pre>{importUrl&&<a className={cn(buttonVariants(),"mt-3 w-full sm:w-fit")} href={importUrl}><ExternalLink/>一键导入 Axolotl Launcher</a>}</div></CardContent><CardFooter className="gap-2"><Dialog><DialogTrigger render={<Button disabled={generating}/>}>生成或轮换</DialogTrigger><DialogContent><DialogHeader><DialogTitle>生成新的 API Key？</DialogTitle><DialogDescription>旧 Key 会立即失效。新 Key 生成后可随时在本页查看和复制。</DialogDescription></DialogHeader><DialogFooter><DialogClose render={<Button onClick={generate} disabled={generating}><KeyRound/>确认生成</Button>}/></DialogFooter></DialogContent></Dialog><AlertDialog><AlertDialogTrigger render={<Button variant="destructive" disabled={revoking}/>}>吊销</AlertDialogTrigger><AlertDialogContent><AlertDialogHeader><AlertDialogTitle>吊销 API Key？</AlertDialogTitle><AlertDialogDescription>该操作不可撤销，Launcher 中使用此 Key 的配置也会失效。</AlertDialogDescription></AlertDialogHeader><AlertDialogFooter><AlertDialogCancel>取消</AlertDialogCancel><AlertDialogAction onClick={revoke}>确认吊销</AlertDialogAction></AlertDialogFooter></AlertDialogContent></AlertDialog></CardFooter></Card>
}
