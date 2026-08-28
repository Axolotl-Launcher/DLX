import { useState } from "react"
import { toast } from "sonner"
import { Copy, KeyRound, RefreshCw, ShieldCheck } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from "@/components/ui/alert-dialog"
import { Toaster } from "@/components/ui/sonner"

const sponsorURL = "https://ifdian.net/a/Mystic-Stars"

export function App() {
  const [loginCode, setLoginCode] = useState("")
  const [claimed, setClaimed] = useState(false)
  const [apiKey, setAPIKey] = useState<string | null>(null)

  async function signInWithRecoveryCode() {
    if (!loginCode.startsWith("axl_login_")) { toast.error("请输入完整的登录码"); return }
    try {
      const origin = import.meta.env.VITE_API_ORIGIN ?? "https://api.axlmc.org"
      const response = await fetch(`${origin}/auth/recovery-login`, {
        method: "POST", credentials: "include", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ login_code: loginCode }),
      })
      if (!response.ok) { toast.error("登录码无效或已被吊销"); return }
      setLoginCode("")
      toast.success("登录成功")
    } catch { toast.error("暂时无法连接赞助者服务") }
  }
  function copyKey() {
    if (!apiKey) return
    navigator.clipboard.writeText(apiKey).then(() => toast.success("API Key 已复制")).catch(() => toast.error("无法访问剪贴板，请手动复制"))
  }
  async function generateKey() {
    try {
      const origin = import.meta.env.VITE_API_ORIGIN ?? "https://api.axlmc.org"
      const response = await fetch(`${origin}/me/api-key`, { method: "POST", credentials: "include" })
      const body = await response.json() as { api_key?: string; code?: string }
      if (!response.ok || !body.api_key) { toast.error(body.code === "SPONSORSHIP_REQUIRED" ? "尚未达到永久开通门槛" : "无法生成 API Key"); return }
      setAPIKey(body.api_key)
      toast.success("请立即复制并妥善保存 API Key")
    } catch { toast.error("暂时无法连接赞助者服务") }
  }
  async function revokeKey() {
    try {
      const origin = import.meta.env.VITE_API_ORIGIN ?? "https://api.axlmc.org"
      const response = await fetch(`${origin}/me/api-key`, { method: "DELETE", credentials: "include" })
      if (!response.ok) { toast.error("无法吊销 API Key"); return }
      setAPIKey(null); toast.success("API Key 已吊销")
    } catch { toast.error("暂时无法连接赞助者服务") }
  }

  return <main className="min-h-svh bg-muted/30 p-4 text-foreground sm:p-8">
    <Toaster richColors />
    <div className="mx-auto max-w-5xl">
      <header className="mb-8 flex flex-col justify-between gap-3 sm:flex-row sm:items-center">
        <div><p className="text-sm font-medium text-muted-foreground">Axolotl</p><h1 className="text-2xl font-semibold tracking-tight">Sponsor Translate</h1></div>
        <a className="text-sm font-medium underline underline-offset-4" href={sponsorURL} target="_blank" rel="noreferrer">前往爱发电支持</a>
      </header>
      <Tabs defaultValue="dashboard">
        <TabsList aria-label="赞助者中心导航"><TabsTrigger value="dashboard">概览</TabsTrigger><TabsTrigger value="claim">订单认领</TabsTrigger><TabsTrigger value="api-key">API Key</TabsTrigger></TabsList>
        <TabsContent value="dashboard" className="mt-5 grid gap-4 md:grid-cols-3">
          <Card className="md:col-span-2"><CardHeader><CardTitle className="flex items-center gap-2"><ShieldCheck className="size-5" />永久资格</CardTitle><CardDescription>按累计实际成功支付金额计算，不设月度到期。</CardDescription></CardHeader><CardContent><p className="text-lg font-medium">{claimed ? "感谢你的支持！累计支持满 ¥9.90，Axolotl Sponsor Translate 已永久开通。" : "当前尚未认领订单。"}</p></CardContent></Card>
          <Card><CardHeader><CardTitle>今日使用</CardTitle><CardDescription>不保存翻译正文</CardDescription></CardHeader><CardContent className="space-y-2"><p>请求数：—</p><p>最近调用：—</p><p>最近错误：—</p></CardContent></Card>
        </TabsContent>
        <TabsContent value="claim" className="mt-5"><Card><CardHeader><CardTitle>认领爱发电订单</CardTitle><CardDescription>提交一笔属于你的订单号。系统将查询同一爱发电身份的相关订单并重算累计金额。</CardDescription></CardHeader><CardContent className="grid gap-3"><label className="grid gap-1.5" htmlFor="order-id"><span className="font-medium">订单号</span><Input id="order-id" placeholder="输入爱发电订单号" aria-describedby="order-help" /></label><p id="order-help" className="text-sm text-muted-foreground">订单号仅用于定位身份；不会以单笔订单金额直接授权。</p></CardContent><CardFooter><Button onClick={() => { setClaimed(true); toast.success("订单已提交，正在安全核验") }}><RefreshCw />提交并核验</Button></CardFooter></Card></TabsContent>
        <TabsContent value="api-key" className="mt-5"><Card><CardHeader><CardTitle className="flex items-center gap-2"><KeyRound className="size-5" />API Key</CardTitle><CardDescription>完整 Key 仅在创建或轮换后显示一次，绝不通过 URL 传递。</CardDescription></CardHeader><CardContent className="space-y-3">{apiKey ? <div className="flex gap-2"><Input value={apiKey} readOnly aria-label="新创建的 API Key" /><Button variant="outline" onClick={copyKey}><Copy />复制</Button></div> : <p className="text-muted-foreground">达到永久资格后可生成一个 active Key。</p>}<p className="text-sm text-muted-foreground">Launcher：自定义端点填写 <code>https://api.axlmc.org/v1/translate</code>，并粘贴此 Key。</p></CardContent><CardFooter className="gap-2"><Dialog><DialogTrigger render={<Button />}>生成或轮换</DialogTrigger><DialogContent><DialogHeader><DialogTitle>生成新的 API Key？</DialogTitle><DialogDescription>生成后旧 Key 会立即失效，关闭一次性展示后无法再次查看完整 Key。</DialogDescription></DialogHeader><DialogFooter><Button onClick={generateKey}>确认生成</Button></DialogFooter></DialogContent></Dialog><AlertDialog><AlertDialogTrigger render={<Button variant="destructive" disabled={!apiKey} />}>吊销</AlertDialogTrigger><AlertDialogContent><AlertDialogHeader><AlertDialogTitle>吊销 API Key？</AlertDialogTitle><AlertDialogDescription>吊销后 Launcher 将无法继续调用翻译服务。</AlertDialogDescription></AlertDialogHeader><AlertDialogFooter><AlertDialogCancel>取消</AlertDialogCancel><AlertDialogAction onClick={revokeKey}>确认吊销</AlertDialogAction></AlertDialogFooter></AlertDialogContent></AlertDialog></CardFooter></Card></TabsContent>
      </Tabs>
      <Card className="mt-5 max-w-md"><CardHeader><CardTitle>登录赞助者中心</CardTitle><CardDescription>输入认领订单后一次性获得的登录码。请自行妥善保管；服务端无法再次显示完整代码。</CardDescription></CardHeader><CardContent><label className="grid gap-1.5" htmlFor="login-code"><span className="font-medium">登录码</span><Input id="login-code" autoComplete="off" value={loginCode} onChange={(event) => setLoginCode(event.target.value)} placeholder="axl_login_…" aria-invalid={loginCode.length > 0 && !loginCode.startsWith("axl_login_")} /></label></CardContent><CardFooter><Button onClick={signInWithRecoveryCode}>登录</Button></CardFooter></Card>
      <section className="mt-8" aria-labelledby="usage-title"><h2 id="usage-title" className="mb-3 text-lg font-semibold">最近使用状态</h2><Table><TableHeader><TableRow><TableHead>日期</TableHead><TableHead>请求数</TableHead><TableHead>输入字符</TableHead><TableHead>错误</TableHead></TableRow></TableHeader><TableBody><TableRow><TableCell>—</TableCell><TableCell>—</TableCell><TableCell>—</TableCell><TableCell>—</TableCell></TableRow></TableBody></Table></section>
    </div>
  </main>
}
export default App
