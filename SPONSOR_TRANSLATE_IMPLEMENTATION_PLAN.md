# Axolotl 赞助者翻译服务：完整改造计划

## 1. 已确认的产品规则

- 用户在爱发电对 Axolotl 的**累计实际成功支付金额**达到或超过 **¥9.90（990 分）**，永久获得翻译 API 权限。
- 不设套餐、档位、月度额度或续费失效机制；停止月度赞助后仍可使用。
- 退款、撤销订单或管理员封禁时，重新计算净累计金额；若低于 990 分，可暂停 Key 并进入审计/人工处理流程。
- 爱发电的“买一杯咖啡 ¥9.90/月”仅是推荐赞助方式，不是授权有效期。

用户主流程：爱发电赞助 → 赞助者中心登录 → 提交订单号认领 → 服务端汇总累计金额 → 生成 API Key → 在 Launcher 粘贴并测试 → 翻译。

## 2. 现状与架构决策

### 当前 DLX

- Go/Gin 服务，默认监听 1188；接口包括 /translate、/v1/translate、/v2/translate。
- service/service.go 的 TOKEN 是单一全局令牌，不能用于用户 API Key、爱发电订单或授权管理。
- compose.yaml 当前暴露 1188；生产环境必须停止对公网暴露该端口。

### 当前 Launcher

源代码位于 D:\Code\Minecraft\axolotl-launcher。

- Rust 翻译适配：packages/app-lib/src/api/translation.rs。
- 设置 UI：apps/app-frontend/src/components/ui/settings/TranslationSettings.vue。
- 现有“DeepL 自定义端点”对非官方端点发送 Authorization: Bearer <key>，并发送含字符串 text、source_lang、target_lang 的 JSON；这与 DLX 兼容。
- 因此不要重写 Launcher 翻译调用链。新网关应实现同样的请求/响应兼容层。
- 当前代码有请求正文、文本预览和 Key 前缀的 tracing/debug 日志。上线前必须删除，翻译文本和凭据不得写日志或上传遥测。

### 目标拓扑

Launcher -- HTTPS + 用户 API Key --> api.axlmc.org (Sponsor Gateway) --> 内网 DLX --> 翻译上游。

sponsor.axlmc.org 的赞助者中心与 api.axlmc.org 的 Gateway 共用数据库和认证服务。

公网仅开放 80/443；DLX:1188、PostgreSQL 和 Redis 只能在 Docker 内网访问。用户 Key 只验证于 Gateway；Gateway 用独立内部 Token 调用 DLX。

## 3. 新增 Sponsor Gateway

在 DLX 仓库中新建 sponsor-gateway/，不要把支付和用户系统塞入 service/service.go。

建议目录：

- sponsor-gateway/cmd/server：HTTP 服务入口；
- sponsor-gateway/internal/api：路由、会话、中间件、错误响应；
- sponsor-gateway/internal/auth：邮箱验证码、用户会话、API Key；
- sponsor-gateway/internal/afdian：爱发电 client、签名、Webhook、同步；
- sponsor-gateway/internal/entitlement：累计金额与永久资格计算；
- sponsor-gateway/internal/translate：DLX client 和错误映射；
- sponsor-gateway/internal/usage：Redis 限流与日用量；
- sponsor-gateway/internal/store：PostgreSQL 查询与迁移；
- sponsor-gateway/internal/worker：同步和失败重试。

推荐继续使用 Go，降低运维复杂度。Gateway 负责业务，DLX 只负责翻译。

## 4. 数据库设计

创建下列最小表：

| 表 | 关键字段 | 责任 |
|---|---|---|
| users | id, email, status, created_at | 赞助者中心账户 |
| afdian_identities | user_id, afdian_user_id, verified_at | 本地账户与爱发电身份绑定 |
| afdian_orders | out_trade_no, afdian_user_id, actual_paid_fen, status, raw_payload, synced_at | 不可变订单账本 |
| entitlements | user_id, lifetime_paid_fen, status, granted_at, recalculated_at | 永久资格状态 |
| api_keys | id, user_id, prefix, secret_hash, status, created_at, last_used_at | 用户 Key |
| usage_daily | user_id, date, request_count, input_chars, error_count | 无原文日聚合 |
| webhook_events | provider_event_key, payload, received_at, processed_at, result | 回调幂等 |
| audit_logs | actor, action, target, metadata, created_at | 管理员和用户操作审计 |

约束：

- afdian_orders.out_trade_no 唯一；
- 一个 afdian_user_id 只能绑定一个 users.id；冲突进入人工处理；
- 金额统一用整数分；
- api_keys 只存 HMAC-SHA-256 或 Argon2id 哈希；不能存明文；
- 每个用户首版最多一个 active Key；
- 翻译原文不入库。

## 5. 爱发电授权和订单认领

### 配置

在受限服务器 Secret 中配置：AFDIAN_USER_ID、AFDIAN_TOKEN、AFDIAN_WEBHOOK_SECRET、AFDIAN_MIN_LIFETIME_PAID_FEN=990。

爱发电开发者入口：https://afdian.com/dashboard/dev

需要调用的开放接口：

- POST https://afdian.com/api/open/query-order；
- POST https://afdian.com/api/open/query-sponsor。

实际签名规则、字段含义、Webhook 契约必须以团队自己的爱发电开发者后台文档和真实小额订单联调结果为准。可参考公开 SDK：https://github.com/DearLicy/AfdianClient。

### 资格算法

net_paid_fen = 所有属于同一 afdian_user_id、属于 Axolotl、支付成功且有效订单的实际支付金额之和 - 已退款/撤销金额。

eligible = net_paid_fen >= 990。

必须使用订单实际支付金额 total_amount，不能使用可能为折扣前金额的 show_amount。订单号只用于定位爱发电 user_id，不能只依据用户提交的单笔订单金额决定资格。

### 用户认领流程

1. 用户在 sponsor.axlmc.org 输入邮箱，获取并验证验证码；
2. 登录后提交一笔属于自己的爱发电订单号；
3. Gateway 用 query-order 查询并确认支付成功、归属于 Axolotl；
4. 取得 afdian_user_id，查询/同步该身份的所有相关订单；
5. 在事务中写入身份、订单和 entitlement；
6. 净累计达到 990 分时显示“永久开通”，允许生成 Key；未达到时显示还差金额；
7. 同一订单或爱发电身份被另一账户绑定时不泄露原账户信息，只提示联系客服。

### Webhook 和同步

Webhook 为实时路径：验证回调 Secret/签名 → 先持久化 webhook_events 并以事件键幂等 → 根据订单号 query-order 二次核验 → upsert 订单 → 重算权益。

同步为兜底路径：每日同步近期订单；用户面板允许每 10 分钟手动同步一次自己的订单；网络失败指数退避。退款、金额异常和状态不明进入管理员队列。无需实现月度过期、宽限期或月度额度任务。

## 6. 用户 API Key 和翻译 API

Key 格式：axl_live_<key_id>_<高熵随机秘密>。

行为：

- 首次达到资格后，用户在面板点击“生成 API Key”；
- 完整 Key 仅在创建/轮换后显示一次；
- 轮换立即使旧 Key 失效；
- 停止月度赞助不失效；
- Key 只以 Authorization: Bearer <key> 发送，绝不通过 URL；
- 日志仅可记录 request_id 和内部 key id，不得记录 Key 或可识别前缀。

公开接口：

- POST /v1/translate：Launcher 翻译；
- GET /v1/account：Key 状态与赞助资格；
- POST /auth/request-code：发送邮箱验证码；
- POST /auth/verify-code：登录；
- POST /me/afdian/claim：提交订单号认领；
- POST /me/afdian/sync：限频同步；
- POST /me/api-key：创建或轮换；
- DELETE /me/api-key：吊销；
- POST /webhooks/afdian：回调；
- GET /healthz 和 GET /readyz：健康检查。

Launcher 翻译请求必须兼容：

POST /v1/translate
Authorization: Bearer axl_live_...
Content-Type: application/json

{ "text": "Hello", "source_lang": "EN", "target_lang": "ZH" }

返回：

{ "translations": [{ "text": "你好", "detected_source_language": "EN" }] }

错误必须包含稳定 code：INVALID_API_KEY、SPONSORSHIP_REQUIRED、RATE_LIMITED、UPSTREAM_BUSY、SERVICE_UNAVAILABLE、UPSTREAM_TIMEOUT。

请求处理顺序：校验请求体和文本长度 → Key 哈希验证 → entitlement 为 granted → Redis 原子限流 → 生成 request_id → 调用内网 DLX → 记录无原文 usage_daily → 返回标准响应。

统一公平使用保护（不是分档）：每 Key/用户/IP 限流、文本和请求体大小限制、全局 Gateway→DLX 并发限制、上游连续 429/403/5xx 时熔断退避、可人工封禁滥用 Key。

## 7. Sponsor Web 前端

### 强制规范

任何新建的赞助者中心前端必须：

- 使用 shadcn/ui；
- shadcn preset 必须是 b27GcrRo；
- 初始化该 preset 后再添加组件，禁止改用其他 preset；
- 使用 React + TypeScript + Tailwind；
- 使用 shadcn 的 Button、Input、Card、Dialog/AlertDialog、Toast、Table、Tabs 等基础组件，而不是另造冲突的基础组件；
- 必须响应式、键盘可操作，并正确提供 label、错误文案和 aria 状态。

若无既有 Web 技术栈约束，使用 Vite + React + TypeScript + shadcn/ui（preset=b27GcrRo），静态产物由 Caddy/Nginx 托管。

### 首版页面

1. 登录页：邮箱、验证码、登录；
2. Dashboard：资格状态、累计金额、Key 状态、最近使用时间和引导按钮；
3. 订单认领：输入订单号、订单号帮助、验证进度、未达标差额、前往 https://ifdian.net/a/Mystic-Stars 的链接；
4. API Key 卡片：生成、一次性展示、复制、轮换、吊销、Launcher 接入说明；
5. 使用状态卡片：今日请求数、最近调用时间、最近错误；首版不做复杂图表；
6. 最小管理员页：订单同步失败、身份冲突、退款异常、手动重算权益、暂停/恢复 Key；所有操作记 audit_logs。

文案：

- 已开通：感谢你的支持！累计支持满 ¥9.90，Axolotl Sponsor Translate 已永久开通。
- 未达标：当前累计支持 ¥x.xx，距离永久开通还差 ¥x.xx。
- Key：请立即复制并妥善保存 API Key；关闭此页面后无法再次查看完整 Key。

## 8. Launcher 范围（本阶段不改造）

本阶段**不修改** Axolotl Launcher 的任何代码、界面、默认 Endpoint、日志或本地凭据存储。Launcher 继续使用现有的自定义 DeepL 兼容服务配置能力，由用户手动填写 Gateway 地址和 API Key。

Gateway 必须主动兼容当前 Launcher 行为：

- 使用 Authorization: Bearer <API Key>；
- 接收字符串 text、source_lang、target_lang 的 JSON；
- 返回 translations 数组；
- 兼容当前客户端的并发批处理和错误处理。

后续 Launcher 改造作为独立版本规划，届时再处理 Axolotl 服务预设、面板跳转、错误提示、本地凭据保护及敏感日志清理。当前阶段不能以 Launcher 改造为上线依赖。

## 9. 部署、安全与运维

新增 compose.production.yaml，服务为 caddy、sponsor-gateway、sponsor-web、worker（或 Gateway worker 模式）、dlx、postgres、redis。

要求：

- dlx、postgres、redis 不配置宿主机 ports；
- 镜像固定版本/digest，生产禁止盲用 latest；
- 每服务有 healthcheck；
- Secrets 放入受限权限文件、Docker Secret 或 Secret Manager，禁止提交 Git；
- Caddy/Nginx 强制 HTTPS，限制请求体和超时；
- PostgreSQL 每日加密异地备份，每月执行恢复演练；
- 监控请求量、P95 延迟、401/403/429/5xx、DLX 超时、Webhook 失败、同步延迟、Redis/DB 健康；
- 日志保留 14–30 天，不得含 Key、Authorization、翻译文本、爱发电 Token 或订单敏感个人信息。

上线前需确认 DLX 的上游调用与对外提供服务符合相关服务条款；准备隐私政策，明确不保存翻译正文以及订单数据用途。

## 10. 实施顺序

### A. 基础 Gateway 和私有 DLX

1. 将 DLX 改为内网服务，补 healthcheck；
2. 建 Sponsor Gateway、数据库迁移、Redis 限流和 Dockerfile；
3. 实现测试用户的 Key 验证和 Gateway→DLX 转发；
4. 实现 POST /v1/translate 兼容测试；
5. 使用 curl/Postman 或独立测试客户端完成 Gateway→DLX 真实联调；不修改 Launcher。

验收：有 Key 可在 Launcher 翻译；无 Key、无资格和高频请求得到正确错误。

### B. 爱发电资格与赞助者中心

1. 实现邮箱验证码登录；
2. 实现 query-order、订单认领、订单账本和累计金额重算；
3. 用真实小额订单验证“单笔 9.90”“4.90+5.00”“未达标”“重复绑定”；
4. 用 shadcn/ui preset=b27GcrRo 实现最小面板；
5. 实现 Key 的创建、一次性展示、复制和轮换。

验收：提交订单号后可正确汇总全部订单；累计达到 990 分可创建 Key；不可盗领。

### C. 回调、同步与管理员处理

1. 配置爱发电 Webhook；
2. 实现验签、落库、幂等、订单二次核验；
3. 实现每日同步和用户手动限频同步；
4. 实现退款/异常订单队列及管理员处理；
5. 加入审计、告警、备份。

验收：重复/漏失回调均不会造成错发或漏发；异常可追踪和人工处理。

### D. 服务端安全、灰度和开放

1. 验证 Gateway 与现有 Launcher 请求/响应格式的兼容性，但不提交 Launcher 改动；
2. 确认 Gateway、Panel、Worker 的敏感日志与遥测脱敏；
3. 验证文本长度、并发批处理、限流和上游熔断；
4. 邀请少量已达标赞助者使用手动配置方式灰度；
5. 基于稳定性调整统一公平使用阈值；
6. 全量开放。

验收：用户可在 5 分钟内完成“订单号认领 → 获取 Key → 手动填入兼容客户端并测试连接”；DLX 不存在公网端口；Gateway、Panel 和 Worker 的日志不出现 Key 和翻译原文。

## 11. 测试清单

- 金额精度：4.90 + 5.00 = 9.90；
- 成功、失败、退款、撤销订单的净累计计算；
- Webhook、订单同步、订单认领的幂等与并发；
- 同一订单/爱发电身份的盗领防护；
- Key 创建、哈希验证、轮换和吊销；
- 限流、熔断和 DLX 超时；
- DLX 请求兼容和 translations 响应兼容；
- Launcher 对 401/403/429/503/504 的中文提示；
- 日志/遥测中不存在 Key、Authorization、翻译正文。

## 12. 首版非目标

- 不做套餐、档位、月度配额和到期停用；
- 不做多 Key、多设备 Key、团队/组织账户；
- 不做积分、余额、发票、复杂支付；
- 不把爱发电账号密码/Token 传入 Launcher；
- 不公开 DLX 1188；
- 不保存翻译原文。
