import { BarChart3 } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { useUsage } from "@/hooks/useUsage"

const DEEPL_RATE = 25
const HEAT = ["var(--heat-0)", "var(--heat-1)", "var(--heat-2)", "var(--heat-3)", "var(--heat-4)"]
const WEEKDAYS = ["", "周一", "", "周三", "", "周五", ""]
function compact(value: number) { return value >= 1000000 ? (value / 1000000).toFixed(1) + "M" : value >= 1000 ? (value / 1000).toFixed(1) + "K" : String(value) }
function intensity(value: number, max: number) { if (!value || !max) return 0; const ratio = value / max; return ratio > .8 ? 4 : ratio > .6 ? 3 : ratio > .3 ? 2 : ratio > .08 ? 1 : 0 }
function dayKey(date: Date) { return date.toISOString().slice(0, 10) }

export function UsageHeatmap() {
  const { data, loading } = useUsage()
  if (loading) return <Card><CardContent className="p-6"><div className="h-36 animate-pulse rounded-xl bg-muted" /></CardContent></Card>
  if (!data) return null
  const today = new Date(); today.setHours(0, 0, 0, 0)
  const start = new Date(today); start.setDate(today.getDate() - 364 - today.getDay())
  const cells = Array.from({ length: 371 }, (_, index) => { const date = new Date(start); date.setDate(start.getDate() + index); const key = dayKey(date); return { date, key, item: data.days.find(day => day.date === key) } })
  const weeks = Array.from({ length: 53 }, (_, index) => cells.slice(index * 7, index * 7 + 7))
  const max = Math.max(...cells.map(cell => cell.item?.input_chars ?? 0))
  const equivalent = data.total_input_chars / 1000000 * DEEPL_RATE
  return <Card className="animate-fade-in shadow-none">
    <CardHeader className="gap-1 pb-4 sm:flex-row sm:items-end sm:justify-between">
      <div><CardTitle className="flex items-center gap-2 text-base"><BarChart3 className="size-4 text-muted-foreground" />翻译活动</CardTitle><p className="mt-1 text-xs text-muted-foreground">{data.total_request_count.toLocaleString()} 次请求 · {compact(data.total_input_chars)} 字符 · 等效 {"$" + equivalent.toFixed(2)}</p></div>
      <span className="text-xs text-muted-foreground">最近一年</span>
    </CardHeader>
    <CardContent className="px-4 pb-5 sm:px-6 sm:pb-6">
      <div className="overflow-x-auto pb-2"><div className="grid min-w-[760px] grid-cols-[32px_minmax(0,1fr)] justify-start gap-x-2">
        <div />
        <div className="grid grid-cols-[repeat(53,minmax(12px,1fr))] gap-1 text-[10px] text-muted-foreground">{weeks.map((week, index) => { const month = week.find(cell => cell.date.getDate() <= 7); return <span key={index}>{month ? month.date.toLocaleString("zh-CN", { month: "short" }) : ""}</span> })}</div>
        <div className="grid grid-rows-7 gap-1 text-[10px] leading-3 text-muted-foreground">{WEEKDAYS.map((name, index) => <span key={index} className="h-3">{name}</span>)}</div>
        <div className="grid grid-flow-col grid-cols-[repeat(53,minmax(12px,1fr))] grid-rows-7 gap-1">{cells.map(cell => { const chars = cell.item?.input_chars ?? 0; const requests = cell.item?.request_count ?? 0; const text = cell.key + " · " + compact(chars) + " 字符 · " + requests + " 次请求"; return <span key={cell.key} className="size-3 rounded-[3px] ring-1 ring-inset ring-black/[.04] transition-transform hover:z-10 hover:scale-125 dark:ring-white/[.06]" style={{ backgroundColor: HEAT[intensity(chars, max)] }} title={text} aria-label={text} /> })}</div>
      </div></div>
      <div className="mt-3 flex items-center justify-end gap-1.5 text-[11px] text-muted-foreground">少 <span className="inline-flex gap-1">{HEAT.map((color, index) => <i key={index} className="size-3 rounded-[3px] ring-1 ring-inset ring-black/[.04] dark:ring-white/[.06]" style={{ backgroundColor: color }} />)}</span> 多</div>
    </CardContent>
  </Card>
}