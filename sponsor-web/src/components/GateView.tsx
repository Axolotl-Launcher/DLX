import { TriangleAlert } from "lucide-react"

import { ClaimPanel } from "@/components/ClaimPanel"
import { LoginPanel } from "@/components/LoginPanel"
import { useClaim } from "@/hooks/useClaim"

export interface GateViewProps {
  meError: boolean
  onLogin: (code: string) => Promise<boolean>
}

export function GateView({ meError, onLogin }: GateViewProps) {
  const claim = useClaim()

  return (
    <div className="grid gap-6">
      {meError && (
        <div className="flex items-start gap-2 rounded-2xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          <TriangleAlert className="mt-0.5 size-4 shrink-0" />
          <p>
            暂时无法连接赞助者服务。如果已经登录，请检查网络后刷新页面重试；否则仍可尝试认领订单。
          </p>
        </div>
      )}

      <section className="grid gap-2">
        <h2 className="text-2xl font-semibold tracking-tight">欢迎使用赞助者中心</h2>
        <p className="max-w-xl text-sm text-muted-foreground">
          在爱发电累计支持满 ¥9.90 即可永久开通 Sponsor Translate 翻译 API。
          认领一笔订单即可生成登录码；已有登录码可直接登录。
        </p>
      </section>

      <div className="grid gap-4 md:grid-cols-2">
        <ClaimPanel
          claiming={claim.claiming}
          loginCode={claim.loginCode}
          onClaim={claim.claim}
          onDismiss={claim.dismiss}
          onLogin={onLogin}
        />
        <LoginPanel onLogin={onLogin} />
      </div>
    </div>
  )
}