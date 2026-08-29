export type EntitlementStatus =
  | "pending"
  | "granted"
  | "suspended"
  | "manual_review"

export interface MeResponse {
  lifetime_paid_fen: number
  entitlement_status: EntitlementStatus
  threshold_fen: number
}

export interface ClaimResponse {
  login_code: string
  message: string
}

export interface LoginResponse {
  status: string
}

export interface ApiKeyResponse {
  api_key: string
  message: string
}

export interface ApiErrorBody {
  code: string
  message: string
  request_id?: string
}