import { ExternalLink, ShieldCheck } from "lucide-react"

import { Button, buttonVariants } from "@/components/ui/button"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { ThemeToggle } from "@/components/ThemeToggle"
import { SPONSOR_URL } from "@/lib/constants"
import { cn } from "@/lib/utils"

export interface HeaderProps {
  authenticated: boolean
  onLogout: () => void
}

export function Header({ authenticated, onLogout }: HeaderProps) {
  return (
    <header className="mb-8 flex flex-wrap items-center justify-between gap-4">
      <div className="flex items-center gap-3">
        <div className="grid size-10 place-items-center rounded-2xl bg-primary text-primary-foreground">
          <ShieldCheck className="size-5" />
        </div>
        <div>
          <p className="text-xs font-medium text-muted-foreground">Axolotl</p>
          <h1 className="text-xl font-semibold tracking-tight">赞助者中心 · Sponsor Translate</h1>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <ThemeToggle />
        <a
          className={cn(buttonVariants({ variant: "outline" }))}
          href={SPONSOR_URL}
          target="_blank"
          rel="noreferrer"
        >
          去爱发电支持 <ExternalLink />
        </a>
        <AlertDialog>
          <AlertDialogTrigger
            render={<Button variant="ghost" disabled={!authenticated} />}
          >
            退出登录
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>退出登录？</AlertDialogTitle>
              <AlertDialogDescription>
                退出仅清除本机的登录界面状态；服务端会话 Cookie 会在 7 天后自动过期。
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>取消</AlertDialogCancel>
              <AlertDialogAction onClick={onLogout}>确认退出</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </header>
  )
}