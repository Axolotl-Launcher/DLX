import { ApiKeyPanel } from "@/components/ApiKeyPanel"
import { StatusOverview } from "@/components/StatusOverview"
import type { MeResponse } from "@/lib/types"
export interface DashboardProps { me:MeResponse; activeTab:"overview"|"api-key" }
export function Dashboard({me,activeTab}:DashboardProps){return <div className="animate-fade-in-up">{activeTab==="overview"?<StatusOverview me={me}/>:<ApiKeyPanel me={me}/>}</div>}
