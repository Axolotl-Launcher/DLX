import { KeyRound, ShieldAlert } from "lucide-react"

import { OneTimeSecret } from "@/components/OneTimeSecret"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { Button, buttonVariants } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { useApiKey } from "@/hooks/useApiKey"
import { SPONSOR_URL } from "@/lib/constants"
import { formatYuan, remainingFen } from "@/lib/format"
import type { EntitlementStatus, MeResponse } from "@/lib/types"
import { cn } from "@/lib/utils"

const LOCKED_COPY: Record<
  Exclude<EntitlementStatus, "granted">,
  { title: string; body: string }
> = {
  pending: {
    title: "尚未达到永久开通门槛",
    body: "",
  },
  manual_review: {
    title: "账户正在人工审核",
    body: "审核通过后将自动获得 API Key 生成权限，请耐心等待。",
  },
  suspended: {
    title: "资格已暂停",
    body: "请通过爱发电私信联系作者处理。",
  },
}

export interface ApiKeyPanelProps {
  me: MeResponse
}

export function ApiKeyPanel({ me }: ApiKeyPanelProps) {
  const { generating, revoking, pendingKey, generate, revoke } = useApiKey()
  const status = me.entitlement_status

  if (status !== "granted") {
    const locked = LOCKED_COPY[status]
    const remaining = remainingFen(me.lifetime_paid_fen, me.threshold_fen)
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ShieldAlert className="size-5" />
            {locked.title}
          </CardTitle>
          <CardDescription>
            {status === "pending"
              ? `当前累计支持 ${formatYuan(me.lifetime_paid_fen)}，还差 ${formatYuan(remaining)}。达标后即可在此生成 API Key。`
              : locked.body}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {status === "pending" && (
            <a
              className={cn(buttonVariants({ variant: "outline" }))}
              href={SPONSOR_URL}
              target="_blank"
              rel="noreferrer"
            >
              前往爱发电支持
            </a>
          )}
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <KeyRound className="size-5" />
          API Key
        </CardTitle>
        <CardDescription>
          完整 Key 仅在创建或轮换后展示一次；服务端只保存不可逆哈希，无法找回。
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4">
        {pendingKey ? (
          <OneTimeSecret
            title="新的 API Key"
            value={pendingKey}
            description="请立即复制并粘贴到 Launcher 的翻译端点设置中。"
            warning="关闭或刷新页面后将无法再次查看完整 Key，请勿分享给他人。"
          />
        ) : (
          <p className="text-sm text-muted-foreground">
            当前没有可显示的 Key。点击「生成或轮换」创建新的 API Key，轮换后旧 Key 立即失效。
          </p>
        )}

        <div className="rounded-2xl border border-border bg-muted/40 p-4 text-sm">
          <h3 className="font-semibold">在 Launcher 中配置</h3>
          <pre className="mt-2 overflow-x-auto rounded-xl bg-background/80 p-3 font-mono text-xs">
            {`端点   https://api.axlmc.org/v1/translate
认证   Authorization: Bearer <你的 API Key>`}
          </pre>
          <p className="mt-2 text-xs text-muted-foreground">
            Launcher 使用 DeepL 兼容的自定义端点，填写上方地址并粘贴 Key 即可使用。
          </p>
        </div>
      </CardContent>
      <CardFooter className="gap-2">
        <Dialog>
          <DialogTrigger render={<Button disabled={generating} />}>
            生成或轮换
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>生成新的 API Key？</DialogTitle>
              <DialogDescription>
                生成后旧 Key 会立即失效；新 Key 仅在本次展示，关闭或刷新页面后无法再次查看完整
                Key。
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <DialogClose
                render={
                  <Button onClick={generate} disabled={generating}>
                    <KeyRound />
                    确认生成
                  </Button>
                }
              />
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <AlertDialog>
          <AlertDialogTrigger render={<Button variant="destructive" disabled={revoking} />}>
            吊销
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>吊销 API Key？</AlertDialogTitle>
              <AlertDialogDescription>
                吊销后 Launcher 将无法继续调用翻译服务，且该操作不可撤销。
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>取消</AlertDialogCancel>
              <AlertDialogAction onClick={revoke}>确认吊销</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </CardFooter>
    </Card>
  )
}