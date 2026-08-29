import { Loader2 } from "lucide-react"

import { Dashboard } from "@/components/Dashboard"
import { GateView } from "@/components/GateView"
import { Header } from "@/components/Header"
import { Toaster } from "@/components/ui/sonner"
import { useSession } from "@/hooks/useSession"

export function App() {
  const { state, login, logout } = useSession()

  return (
    <main className="min-h-svh bg-muted/30 text-foreground">
      <Toaster richColors />
      <div className="mx-auto max-w-5xl px-4 py-6 sm:px-8 sm:py-10">
        <Header
          authenticated={state.status === "authenticated"}
          onLogout={logout}
        />
        {state.status === "loading" && (
          <div className="grid place-items-center py-28">
            <Loader2 className="size-6 animate-spin text-muted-foreground" />
            <span className="sr-only">正在加载账户状态</span>
          </div>
        )}
        {state.status === "anonymous" && (
          <GateView meError={state.meError} onLogin={login} />
        )}
        {state.status === "authenticated" && <Dashboard me={state.me} />}
      </div>
    </main>
  )
}

export default App