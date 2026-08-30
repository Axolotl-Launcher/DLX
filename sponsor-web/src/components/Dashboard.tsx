import { ApiKeyPanel } from "@/components/ApiKeyPanel"
import { StatusOverview } from "@/components/StatusOverview"
import type { CdkRedeemResponse, MeResponse } from "@/lib/types"
export interface DashboardProps { me:MeResponse; activeTab:"overview"|"api-key"; onRedeemed?: (result: CdkRedeemResponse) => void }
export function Dashboard({me,activeTab,onRedeemed}:DashboardProps){return <div className="animate-fade-in-up">{activeTab==="overview"?<StatusOverview me={me} onRedeemed={onRedeemed}/>:<ApiKeyPanel me={me}/>}</div>}