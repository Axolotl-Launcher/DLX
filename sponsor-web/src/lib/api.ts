import type {
  ApiErrorBody,
  ApiKeyResponse,
  ClaimResponse,
  LoginResponse,
  MeResponse,
} from "./types"

// 本地开发走 Vite 代理（/sponsor/* → https://api.axlmc.org），
// 生产直接请求完整域名；VITE_API_ORIGIN 可覆盖二者。
const API_BASE =
  import.meta.env.VITE_API_ORIGIN ??
  (import.meta.env.DEV ? "/sponsor" : "https://api.axlmc.org")

export class ApiError extends Error {
  readonly code: string
  readonly status: number
  readonly requestId?: string

  constructor(code: string, status: number, message: string, requestId?: string) {
    super(message)
    this.name = "ApiError"
    this.code = code
    this.status = status
    this.requestId = requestId
  }
}

const ERROR_MESSAGES: Record<string, string> = {
  UNAUTHENTICATED: "登录已失效，请重新登录",
  INVALID_LOGIN_CODE: "登录码无效或已被吊销",
  INVALID_ORDER: "订单号格式不正确",
  ORDER_VERIFICATION_FAILED: "订单无法核验，请确认订单号属于 Axolotl 且已支付成功",
  ORDER_ALREADY_CLAIMED: "该订单已被其他账户认领，请通过爱发电私信联系作者",
  SPONSORSHIP_REQUIRED: "尚未达到 ¥9.90 永久开通门槛",
  RATE_LIMITED: "操作过于频繁，请稍后再试",
  UPSTREAM_BUSY: "翻译服务目前繁忙，请稍后重试",
  UPSTREAM_TIMEOUT: "翻译服务响应超时，请稍后重试",
  SERVICE_UNAVAILABLE: "赞助者服务暂时不可用，请稍后重试",
}

export const NETWORK_ERROR_MESSAGE = "暂时无法连接赞助者服务，请检查网络后重试"

export function errorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return ERROR_MESSAGES[error.code] ?? error.message
  }
  return NETWORK_ERROR_MESSAGE
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response
  try {
    response = await fetch(`${API_BASE}${path}`, {
      credentials: "include",
      headers: { "Content-Type": "application/json", ...init?.headers },
      ...init,
    })
  } catch {
    throw new ApiError("NETWORK_ERROR", 0, NETWORK_ERROR_MESSAGE)
  }

  if (response.status === 204) {
    return undefined as T
  }

  const body: unknown = await response.json().catch(() => null)
  if (!response.ok) {
    const error = (body as ApiErrorBody | null) ?? {
      code: "UNKNOWN",
      message: `请求失败（HTTP ${response.status}）`,
    }
    throw new ApiError(error.code, response.status, error.message, error.request_id)
  }

  return body as T
}

export const api = {
  me: () => request<MeResponse>("/me"),

  claim: (orderNo: string) =>
    request<ClaimResponse>("/afdian/claim", {
      method: "POST",
      body: JSON.stringify({ order_no: orderNo }),
    }),

  login: (loginCode: string) =>
    request<LoginResponse>("/auth/recovery-login", {
      method: "POST",
      body: JSON.stringify({ login_code: loginCode }),
    }),

  generateApiKey: () => request<ApiKeyResponse>("/me/api-key", { method: "POST" }),

  revokeApiKey: () => request<void>("/me/api-key", { method: "DELETE" }),
}