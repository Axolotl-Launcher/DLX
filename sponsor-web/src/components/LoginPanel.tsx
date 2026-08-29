import { useState } from "react"
import type { FormEvent } from "react"
import { LogIn } from "lucide-react"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { LOGIN_CODE_PREFIX } from "@/lib/constants"
export interface LoginPanelProps { onLogin:(code:string)=>Promise<boolean> }
export function LoginPanel({onLogin}:LoginPanelProps){const[code,setCode]=useState("");const[busy,setBusy]=useState(false);const invalid=code.length>0&&!code.startsWith(LOGIN_CODE_PREFIX);async function submit(event:FormEvent<HTMLFormElement>){event.preventDefault();const trimmed=code.trim();if(!trimmed.startsWith(LOGIN_CODE_PREFIX)){toast.error("请输入正确格式的登录码");return}setBusy(true);if(await onLogin(trimmed))setCode("");setBusy(false)}return <Card className="h-full transition-shadow duration-300 hover:shadow-md"><CardHeader><CardTitle>登录</CardTitle><CardDescription>使用已保存的登录码继续。</CardDescription></CardHeader><form className="flex flex-1 flex-col" onSubmit={submit}><CardContent className="flex-1"><label className="grid gap-2" htmlFor="login-code"><span className="text-sm font-medium">登录码</span><div className="flex gap-2"><Input className="min-w-0 flex-1" id="login-code" autoComplete="off" autoCapitalize="off" spellCheck={false} value={code} onChange={e=>setCode(e.target.value)} placeholder={LOGIN_CODE_PREFIX+"…"} aria-invalid={invalid||undefined} aria-describedby={invalid?"login-code-error":undefined}/><Button className="shrink-0" type="submit" disabled={busy}><LogIn data-icon="inline-start"/>{busy?"登录中…":"登录"}</Button></div>{invalid&&<p id="login-code-error" className="mt-2 text-xs text-destructive">格式不正确</p>}</label></CardContent></form></Card>}
