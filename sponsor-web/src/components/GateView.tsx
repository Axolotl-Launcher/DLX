import { TriangleAlert } from "lucide-react"
import { CdkRedeemPanel } from "@/components/CdkRedeemPanel"
import { ClaimPanel } from "@/components/ClaimPanel"
import { LoginPanel } from "@/components/LoginPanel"
import { useClaim } from "@/hooks/useClaim"
export interface GateViewProps { meError: boolean; onLogin: (code: string) => Promise<boolean> }
export function GateView({meError,onLogin}:GateViewProps){const claim=useClaim();return <div className="animate-fade-in-up">{meError&&<div className="mb-6 flex items-start gap-2 rounded-2xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"><TriangleAlert className="mt-0.5 size-4 shrink-0" aria-hidden="true"/><p>服务暂时不可用，请稍后重试。</p></div>}<section className="mb-10 max-w-2xl"><p className="mb-3 text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">Member access</p><h2 className="text-3xl font-semibold tracking-[-0.03em] sm:text-5xl">你的翻译 API，<br className="sm:hidden"/>从这里开始。</h2><p className="mt-4 text-sm text-muted-foreground">认领订单、兑换 CDK，或使用已有登录码进入账户。</p></section><div className="grid items-stretch gap-5 md:grid-cols-3"><ClaimPanel claiming={claim.claiming} loginCode={claim.loginCode} onClaim={claim.claim} onDismiss={claim.dismiss} onLogin={onLogin}/><LoginPanel onLogin={onLogin}/><CdkRedeemPanel/></div></div>}
