export type EntitlementStatus = "pending" | "granted" | "suspended" | "manual_review"
export interface MeResponse { lifetime_paid_fen:number; entitlement_status:EntitlementStatus; threshold_fen:number }
export interface UsageDay { date:string; request_count:number; input_chars:number; error_count:number }
export interface UsageResponse { days:UsageDay[]; total_request_count:number; total_input_chars:number; total_error_count:number }
export interface ClaimResponse { login_code:string; message:string }
export interface LoginResponse { status:string }
export interface ApiKeyResponse { api_key:string; message:string; created_at?:string; last_used_at?:string | null }
export interface ApiErrorBody { code:string; message:string; request_id?:string }
