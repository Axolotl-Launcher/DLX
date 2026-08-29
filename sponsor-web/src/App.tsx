import { useState } from "react"
import { Loader2 } from "lucide-react"
import { Dashboard } from "@/components/Dashboard"
import { GateView } from "@/components/GateView"
import { Header } from "@/components/Header"
import { SiteFooter } from "@/components/SiteFooter"
import { Toaster } from "@/components/ui/sonner"
import { useSession } from "@/hooks/useSession"
export function App(){const{state,login,logout}=useSession();const[tab,setTab]=useState<"overview"|"api-key">("overview");const authenticated=state.status==="authenticated";return <main className="min-h-svh bg-background text-foreground"><Toaster richColors position="top-center"/><Header authenticated={authenticated} activeTab={tab} onTabChange={setTab} onLogout={logout}/><div className="mx-auto w-full max-w-6xl px-4 pb-16 pt-10 sm:px-6 sm:pt-14 lg:px-8">{state.status==="loading"?<div className="grid min-h-[55vh] place-items-center"><Loader2 className="size-5 animate-spin text-muted-foreground" aria-hidden="true"/><span className="sr-only">正在加载</span></div>:state.status==="anonymous"?<GateView meError={state.meError} onLogin={login}/>:<Dashboard me={state.me} activeTab={tab}/>}</div><SiteFooter/></main>}
export default App
