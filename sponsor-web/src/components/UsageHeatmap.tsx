import { BarChart3, CircleHelp } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { useUsage } from "@/hooks/useUsage"

const DEEPL_RATE = 25
const LEVELS = ["var(--heat-0)", "var(--heat-1)", "var(--heat-2)", "var(--heat-3)", "var(--heat-4)"]
const ROWS = [
  { key: "request_count", label: "请求" },
  { key: "input_chars", label: "字符" },
  { key: "error_count", label: "错误" },
] as const
function compact(value: number) { return value >= 1000000 ? (value / 1000000).toFixed(1) + "M" : value >= 1000 ? (value / 1000).toFixed(1) + "K" : String(value) }
function intensity(value: number, max: number) { if (!value || !max) return 0; const ratio = value / max; return ratio > .8 ? 4 : ratio > .6 ? 3 : ratio > .35 ? 2 : ratio > .12 ? 1 : 0 }
function label(date: string) { return new Date(date + "T00:00:00").toLocaleDateString("zh-CN", { month: "numeric", day: "numeric" }) }

export function UsageHeatmap() {
  const { data, loading } = useUsage()
  if (loading) return <Card><CardContent className="p-6"><div className="h-52 animate-pulse rounded-xl bg-muted" /></CardContent></Card>
  if (!data) return null
  const days = data.days.slice(-12)
  const maxes = ROWS.map(row => Math.max(...days.map(day => day[row.key])))
  const equivalent = data.total_input_chars / 1000000 * DEEPL_RATE
  return <Card className="overflow-hidden border-border/70 shadow-none">
    <CardHeader className="gap-1 border-b bg-muted/15 px-5 py-4 sm:px-6">
      <div className="flex items-center justify-between gap-3">
        <CardTitle className="flex items-center gap-2 text-base"><BarChart3 className="size-4 text-muted-foreground" />用量概览</CardTitle>
        <span className="inline-flex items-center gap-1 text-xs text-muted-foreground"><CircleHelp className="size-3.5" />每日聚合</span>
      </div>
      <p className="text-xs text-muted-foreground">最近 12 天的翻译活动</p>
    </CardHeader>
    <CardContent className="grid gap-6 p-5 sm:p-6">
      <div className="grid grid-cols-3 divide-x rounded-xl border bg-background">
        <div className="p-3 sm:p-4"><p className="text-xl font-semibold tracking-tight sm:text-2xl">{compact(data.total_input_chars)}</p><p className="mt-1 text-[11px] text-muted-foreground">字符</p></div>
        <div className="p-3 sm:p-4"><p className="text-xl font-semibold tracking-tight sm:text-2xl">{data.total_request_count.toLocaleString()}</p><p className="mt-1 text-[11px] text-muted-foreground">请求</p></div>
        <div className="p-3 sm:p-4"><p className="text-xl font-semibold tracking-tight sm:text-2xl">{"$" + equivalent.toFixed(2)}</p><p className="mt-1 text-[11px] text-muted-foreground">等效价值</p></div>
      </div>
      <div className="overflow-x-auto pb-1">
        <div className="min-w-[680px]">
          <div className="mb-2 grid grid-cols-[44px_repeat(12,minmax(42px,1fr))] gap-1.5">
            <span />{days.map(day => <span key={day.date} className="text-center text-[10px] tabular-nums text-muted-foreground">{label(day.date)}</span>)}
          </div>
          <div className="grid gap-1.5">
            {ROWS.map((row, rowIndex) => <div key={row.key} className="grid grid-cols-[44px_repeat(12,minmax(42px,1fr))] items-center gap-1.5">
              <span className="text-[11px] font-medium text-muted-foreground">{row.label}</span>
              {days.map(day => { const value = day[row.key]; const level = intensity(value, maxes[rowIndex]); return <div key={day.date} className="group relative h-11 rounded-md transition-transform duration-150 hover:z-10 hover:scale-[1.04]" style={{ backgroundColor: LEVELS[level] }} title={day.date + " · " + row.label + " · " + value.toLocaleString()}><span className="pointer-events-none absolute inset-x-0 -bottom-7 z-20 hidden whitespace-nowrap rounded bg-foreground px-2 py-1 text-center text-[10px] text-background shadow-sm group-hover:block">{value.toLocaleString()}</span></div> })}
            </div>)}
          </div>
        </div>
      </div>
      <div className="flex items-center justify-between border-t pt-4 text-[11px] text-muted-foreground">
        <span>低使用量</span><span className="inline-flex items-center gap-1.5">{LEVELS.map((color, index) => <i key={index} className="size-3 rounded-[3px]" style={{ backgroundColor: color }} />)}</span><span>高使用量</span>
      </div>
    </CardContent>
  </Card>
}