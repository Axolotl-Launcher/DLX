# Sponsor Web — Axolotl 赞助者中心

React 19 + TypeScript + Vite + Tailwind CSS v4 + shadcn/ui（base-rhea / b27GcrRo preset）构建的赞助者中心前端。

## 功能

- **认领订单**：`POST /afdian/claim`，核验爱发电订单并一次性展示登录码；
- **登录**：`POST /auth/recovery-login`，Cookie 会话默认 7 天；
- **账户概览**：`GET /me`，展示累计支持金额、永久资格门槛（¥9.90）进度与状态（已开通 / 未达标 / 人工审核 / 暂停）；
- **API Key**：`POST|DELETE /me/api-key`，生成 / 轮换 / 吊销，完整 Key 仅在生成后展示一次，附 Launcher 配置说明；
- 深色 / 浅色主题切换（`theme-provider`，键盘 `D` 快捷切换），中文界面，移动端响应式。

## 开发

```bash
pnpm install
pnpm dev
```

本地开发默认通过 Vite 代理把 `/sponsor/*` 转发到 `https://api.axlmc.org`（见 `vite.config.ts` 的 `server.proxy`），免 CORS 即可调试「认领订单」等无需会话的流程。

注意：生产环境会话 Cookie 的域是 `api.axlmc.org`，本地代理下浏览器不会保存该 Cookie，因此登录态无法在本地持久化；如需完整联调登录 / `/me`，请本地运行 `sponsor-gateway` 并把 `CookieDomain` 配置为本地域名，同时用 `VITE_API_ORIGIN` 指向它。

环境变量：

- `VITE_API_ORIGIN`：覆盖 API 地址。默认：开发环境代理 `/sponsor`，生产 `https://api.axlmc.org`。

## 构建

```bash
pnpm build   # tsc -b && vite build，产物在 dist/
pnpm lint
pnpm typecheck
```

生产镜像由 `Dockerfile`（构建后由 Caddy 托管）产出，域名 `sponsor.axlmc.org`，API 位于 `api.axlmc.org`。

## 目录结构

```
src/
  App.tsx                 # 会话门控骨架：加载 / 未登录(Gate) / 已登录(Dashboard)
  lib/
    api.ts                # fetch 封装、API 错误归一化（错误码 → 中文文案）
    types.ts              # 接口响应类型
    format.ts             # 分 → ¥ 格式化
    constants.ts          # 爱发电地址、登录码前缀
  hooks/
    useSession.ts         # /me 引导、登录、退出（本机状态）
    useClaim.ts           # 认领订单 → 一次性登录码
    useApiKey.ts          # 生成 / 轮换 / 吊销
  components/
    GateView.tsx          # 未登录视图（认领 + 登录）
    ClaimPanel.tsx        # 订单号提交与登录码一次性展示
    LoginPanel.tsx        # 登录码登录
    Dashboard.tsx         # 已登录 Tabs（概览 / API Key）
    StatusOverview.tsx    # 资格进度、使用与隐私、权益说明
    ApiKeyPanel.tsx       # Key 管理与 Launcher 接入说明
    OneTimeSecret.tsx     # 一次性凭据展示（复制 + 警告）
    Header.tsx / ThemeToggle.tsx
  components/ui/          # shadcn 基础组件（勿手改）
```

## 接口依赖（sponsor-gateway）

| 接口 | 说明 |
|---|---|
| `POST /afdian/claim` | 提交订单号，返回一次性 `login_code` |
| `POST /auth/recovery-login` | 登录码换会话 Cookie |
| `GET /me` | 累计金额、资格状态、门槛 |
| `POST /me/api-key` | 生成 / 轮换（一次性返回完整 Key） |
| `DELETE /me/api-key` | 吊销 Key |

用量统计展示依赖后端提供 `GET /me/usage` 之类接口，当前未实现，前端以「暂不提供」占位，不展示假数据。