import { ApiKeyPanel } from "@/components/ApiKeyPanel"
import { StatusOverview } from "@/components/StatusOverview"
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs"
import type { MeResponse } from "@/lib/types"

export interface DashboardProps {
  me: MeResponse
}

export function Dashboard({ me }: DashboardProps) {
  return (
    <Tabs defaultValue="overview">
      <TabsList aria-label="赞助者中心导航">
        <TabsTrigger value="overview">概览</TabsTrigger>
        <TabsTrigger value="api-key">API Key</TabsTrigger>
      </TabsList>
      <TabsContent value="overview" className="mt-5">
        <StatusOverview me={me} />
      </TabsContent>
      <TabsContent value="api-key" className="mt-5">
        <ApiKeyPanel me={me} />
      </TabsContent>
    </Tabs>
  )
}