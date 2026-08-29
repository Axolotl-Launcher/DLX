import { Monitor, Moon, Sun } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useTheme } from "@/components/theme-provider"
const modes=[{value:"light",label:"浅色模式",icon:Sun},{value:"dark",label:"深色模式",icon:Moon},{value:"system",label:"跟随系统",icon:Monitor}] as const
export function ThemeToggle(){const{theme,setTheme}=useTheme();return <div className="flex items-center gap-0.5 rounded-full border border-border bg-background p-0.5" aria-label="明暗主题">{modes.map(mode=><Button key={mode.value} variant="ghost" size="icon-xs" aria-label={mode.label} aria-pressed={theme===mode.value} title={mode.label} className="rounded-full text-muted-foreground hover:text-foreground aria-pressed:bg-muted aria-pressed:text-foreground" onClick={()=>setTheme(mode.value)}><mode.icon /></Button>)}</div>}
