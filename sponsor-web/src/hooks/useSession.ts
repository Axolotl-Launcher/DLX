import { useCallback, useEffect, useState } from "react"
import { toast } from "sonner"

import { api, ApiError, errorMessage } from "@/lib/api"
import type { MeResponse } from "@/lib/types"

export type SessionState =
  | { status: "loading" }
  | { status: "anonymous"; meError: boolean }
  | { status: "authenticated"; me: MeResponse }

export function useSession() {
  const [state, setState] = useState<SessionState>({ status: "loading" })

  useEffect(() => {
    let cancelled = false
    api.me().then(
      (me) => {
        if (!cancelled) {
          setState({ status: "authenticated", me })
        }
      },
      (error: unknown) => {
        if (cancelled) {
          return
        }
        if (error instanceof ApiError && error.status === 401) {
          setState({ status: "anonymous", meError: false })
        } else {
          // 服务不可达时也按未登录展示，但提示服务异常而非引导登录。
          setState({ status: "anonymous", meError: true })
        }
      }
    )
    return () => {
      cancelled = true
    }
  }, [])

  const login = useCallback(
    async (loginCode: string): Promise<boolean> => {
      try {
        await api.login(loginCode)
        const me = await api.me()
        setState({ status: "authenticated", me })
        return true
      } catch (error) {
        toast.error(errorMessage(error))
        return false
      }
    },
    []
  )

  const logout = useCallback(() => {
    setState({ status: "anonymous", meError: false })
  }, [])

  return { state, login, logout }
}