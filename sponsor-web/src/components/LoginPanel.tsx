import { useState } from "react"
import type { FormEvent } from "react"
import { LogIn } from "lucide-react"
import { toast } from "sonner"

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
import { LOGIN_CODE_PREFIX } from "@/lib/constants"

export interface LoginPanelProps {
  onLogin: (code: string) => Promise<boolean>
}

export function LoginPanel({ onLogin }: LoginPanelProps) {
  const [code, setCode] = useState("")
  const [busy, setBusy] = useState(false)
  const invalid = code.length > 0 && !code.startsWith(LOGIN_CODE_PREFIX)

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const trimmed = code.trim()
    if (!trimmed.startsWith(LOGIN_CODE_PREFIX)) {
      toast.error(`请输入完整的登录码（${LOGIN_CODE_PREFIX} 开头）`)
      return
    }
    setBusy(true)
    const ok = await onLogin(trimmed)
    setBusy(false)
    if (ok) {
      setCode("")
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>已有登录码</CardTitle>
        <CardDescription>
          输入认领订单后获得的一次性登录码。服务端无法再次显示完整登录码，请妥善保存。
        </CardDescription>
      </CardHeader>
      <form onSubmit={submit}>
        <CardContent>
          <label className="grid gap-1.5" htmlFor="login-code">
            <span className="font-medium">登录码</span>
            <Input
              id="login-code"
              autoComplete="off"
              autoCapitalize="off"
              spellCheck={false}
              value={code}
              onChange={(event) => setCode(event.target.value)}
              placeholder={`${LOGIN_CODE_PREFIX}…`}
              aria-invalid={invalid || undefined}
              aria-describedby={invalid ? "login-code-error" : undefined}
            />
          </label>
          {invalid && (
            <p id="login-code-error" className="mt-2 text-xs text-destructive">
              登录码应以 {LOGIN_CODE_PREFIX} 开头
            </p>
          )}
        </CardContent>
        <CardFooter>
          <Button type="submit" disabled={busy}>
            <LogIn />
            登录
          </Button>
        </CardFooter>
      </form>
    </Card>
  )
}