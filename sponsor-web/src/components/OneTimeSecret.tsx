import { useState } from "react"
import { Check, Copy, ShieldAlert } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

export interface OneTimeSecretProps {
  title: string
  value: string
  description: string
  warning: string
  action?: { label: string; onClick: () => void }
}

export function OneTimeSecret({
  title,
  value,
  description,
  warning,
  action,
}: OneTimeSecretProps) {
  const [copied, setCopied] = useState(false)

  async function copy() {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      toast.success("已复制")
      window.setTimeout(() => setCopied(false), 2000)
    } catch {
      toast.error("无法访问剪贴板，请手动复制")
    }
  }

  return (
    <div className="grid gap-3 rounded-2xl border border-border bg-muted/40 p-4">
      <div>
        <h3 className="font-semibold">{title}</h3>
        <p className="mt-1 text-sm text-muted-foreground">{description}</p>
      </div>
      <div className="flex gap-2">
        <Input
          readOnly
          aria-label={title}
          value={value}
          className="font-mono text-xs"
        />
        <Button variant="outline" onClick={copy}>
          {copied ? <Check /> : <Copy />}
          复制
        </Button>
      </div>
      <p className="flex items-start gap-2 text-xs text-amber-600 dark:text-amber-400">
        <ShieldAlert className="mt-0.5 size-3.5 shrink-0" />
        {warning}
      </p>
      {action && <Button onClick={action.onClick}>{action.label}</Button>}
    </div>
  )
}