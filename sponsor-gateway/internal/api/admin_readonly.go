package api

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type AdminStore interface {
	AdminOverview(context.Context) (AdminOverview, error)
	ListAdminUsers(context.Context, AdminUserQuery) (AdminUserPage, error)
	GetAdminUser(context.Context, string) (AdminUserDetail, error)
	AdminUserUsage(context.Context, string, UsageQuery) (UsageSummary, error)
	ListAdminOrders(context.Context, string, AdminOrderQuery) (AdminOrderPage, error)
	ListAdminAPIKeys(context.Context, AdminAPIKeyQuery) (AdminAPIKeyPage, error)
}

type AdminOverview struct {
	Users        AdminUserCounts        `json:"users"`
	Entitlements AdminEntitlementCounts `json:"entitlements"`
	Orders       AdminOrderCounts       `json:"orders"`
	Usage        AdminUsageCounts       `json:"usage"`
	CDKs         AdminCDKCounts         `json:"cdks"`
	GeneratedAt  time.Time              `json:"generated_at"`
}
type AdminUserCounts struct {
	Total     int64 `json:"total"`
	Active    int64 `json:"active"`
	Suspended int64 `json:"suspended"`
	Blocked   int64 `json:"blocked"`
}
type AdminEntitlementCounts struct {
	Granted      int64 `json:"granted"`
	Pending      int64 `json:"pending"`
	Suspended    int64 `json:"suspended"`
	ManualReview int64 `json:"manual_review"`
}
type AdminOrderCounts struct {
	PaidCount     int64 `json:"paid_count"`
	PaidAmountFen int64 `json:"paid_amount_fen"`
	RefundedCount int64 `json:"refunded_count"`
}
type AdminUsageCounts struct {
	TodayRequestCount int64 `json:"today_request_count"`
	TodayInputChars   int64 `json:"today_input_chars"`
	TodayErrorCount   int64 `json:"today_error_count"`
}
type AdminCDKCounts struct {
	ActiveCount       int64 `json:"active_count"`
	RedeemedCount     int64 `json:"redeemed_count"`
	RedeemedAmountFen int64 `json:"redeemed_amount_fen"`
}
type AdminUserQuery struct {
	Page, PageSize                    int
	Search, Status, EntitlementStatus string
}
type AdminOrderQuery struct {
	Page, PageSize   int
	Status, From, To string
}
type UsageQuery struct{ From, To string }
type AdminUser struct {
	ID                string       `json:"id"`
	Email             *string      `json:"email"`
	Status            string       `json:"status"`
	CreatedAt         time.Time    `json:"created_at"`
	EntitlementStatus string       `json:"entitlement_status"`
	LifetimePaidFen   int64        `json:"lifetime_paid_fen"`
	GrantedAt         *time.Time   `json:"granted_at,omitempty"`
	ActiveAPIKey      *AdminAPIKey `json:"active_api_key,omitempty"`
}
type AdminAPIKey struct {
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}
type AdminUserPage struct {
	Items    []AdminUser `json:"items"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Total    int64       `json:"total"`
}
type AdminUserDetail struct {
	AdminUser
	RecalculatedAt *time.Time   `json:"recalculated_at,omitempty"`
	UsageSummary   UsageSummary `json:"usage_summary"`
}
type AdminOrder struct {
	UserID    string    `json:"user_id"`
	UserEmail *string   `json:"user_email"`
	AmountFen int64     `json:"actual_paid_fen"`
	Status    string    `json:"status"`
	SyncedAt  time.Time `json:"synced_at"`
}
type AdminOrderPage struct {
	Items    []AdminOrder `json:"items"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Total    int64        `json:"total"`
}

// AdminAPIKeyRecord is the redacted admin view of a user API key. It never
// carries the prefix, secret hash, ciphertext, or any recoverable fragment.
type AdminAPIKeyRecord struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	UserEmail  *string    `json:"user_email"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type AdminAPIKeyQuery struct {
	Page, PageSize int
	Status, Q      string
}

type AdminAPIKeyPage struct {
	Items    []AdminAPIKeyRecord `json:"items"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Total    int64               `json:"total"`
}

var ErrAdminUserNotFound = errors.New("admin user not found")

func adminGET(s *Server, w http.ResponseWriter, r *http.Request, fn func(AdminStore) error) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", "")
		return
	}
	if !s.adminAuthorized(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin authorization required", "")
		return
	}
	store, ok := s.Store.(AdminStore)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "admin service unavailable", "")
		return
	}
	if err := fn(store); err != nil {
		if errors.Is(err, ErrAdminUserNotFound) {
			writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found", "")
			return
		}
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "admin service unavailable", "")
	}
}
func (s *Server) adminOverview(w http.ResponseWriter, r *http.Request) {
	adminGET(s, w, r, func(store AdminStore) error {
		v, err := store.AdminOverview(r.Context())
		if err == nil {
			writeJSON(w, http.StatusOK, v)
		}
		return err
	})
}
func (s *Server) adminUsers(w http.ResponseWriter, r *http.Request) {
	q, err := parseAdminUserQuery(r.URL.Query())
	if err != nil {
		writeError(w, 400, "INVALID_QUERY", "invalid query", "")
		return
	}
	adminGET(s, w, r, func(store AdminStore) error {
		v, e := store.ListAdminUsers(r.Context(), q)
		if e == nil {
			writeJSON(w, 200, v)
		}
		return e
	})
}
func (s *Server) adminOrders(w http.ResponseWriter, r *http.Request) {
	q, err := parseAdminOrderQuery(r.URL.Query(), false)
	if err != nil {
		writeError(w, 400, "INVALID_QUERY", "invalid query", "")
		return
	}
	adminGET(s, w, r, func(store AdminStore) error {
		v, e := store.ListAdminOrders(r.Context(), "", q)
		if e == nil {
			writeJSON(w, 200, v)
		}
		return e
	})
}
func validAdminID(id string) bool {
	if len(id) != 36 {
		return false
	}
	for i, c := range id {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
			continue
		}
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
func (s *Server) adminUserPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/users/")
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" || len(parts) > 2 {
		writeError(w, 404, "NOT_FOUND", "not found", "")
		return
	}
	id := parts[0]
	if !validAdminID(id) {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "invalid user id", "")
		return
	}
	if len(parts) == 1 {
		adminGET(s, w, r, func(store AdminStore) error {
			v, e := store.GetAdminUser(r.Context(), id)
			if e == nil {
				writeJSON(w, 200, v)
			}
			return e
		})
		return
	}
	switch parts[1] {
	case "usage":
		q, e := parseUsageQuery(r.URL.Query())
		if e != nil {
			writeError(w, 400, "INVALID_QUERY", "invalid query", "")
			return
		}
		adminGET(s, w, r, func(store AdminStore) error {
			v, e := store.AdminUserUsage(r.Context(), id, q)
			if e == nil {
				writeJSON(w, 200, v)
			}
			return e
		})
	case "orders":
		q, e := parseAdminOrderQuery(r.URL.Query(), true)
		if e != nil {
			writeError(w, 400, "INVALID_QUERY", "invalid query", "")
			return
		}
		adminGET(s, w, r, func(store AdminStore) error {
			v, e := store.ListAdminOrders(r.Context(), id, q)
			if e == nil {
				writeJSON(w, 200, v)
			}
			return e
		})
	default:
		writeError(w, 404, "NOT_FOUND", "not found", "")
	}
}
func parsePage(v url.Values, allowed map[string]bool) (int, int, error) {
	for k := range v {
		if !allowed[k] || len(v[k]) != 1 {
			return 0, 0, errors.New("invalid query")
		}
	}
	page, size := 1, 25
	var err error
	if v.Get("page") != "" {
		page, err = strconv.Atoi(v.Get("page"))
		if err != nil || page < 1 {
			return 0, 0, errors.New("invalid page")
		}
	}
	if v.Get("page_size") != "" {
		size, err = strconv.Atoi(v.Get("page_size"))
		if err != nil || size < 1 || size > 100 {
			return 0, 0, errors.New("invalid page size")
		}
	}
	return page, size, nil
}
func parseAdminUserQuery(v url.Values) (AdminUserQuery, error) {
	p, s, e := parsePage(v, map[string]bool{"page": true, "page_size": true, "q": true, "status": true, "entitlement_status": true})
	if e != nil {
		return AdminUserQuery{}, e
	}
	if len(v.Get("q")) > 100 {
		return AdminUserQuery{}, errors.New("q")
	}
	if x := v.Get("status"); x != "" && x != "active" && x != "suspended" && x != "blocked" {
		return AdminUserQuery{}, errors.New("status")
	}
	if x := v.Get("entitlement_status"); x != "" && !map[string]bool{"pending": true, "granted": true, "suspended": true, "manual_review": true}[x] {
		return AdminUserQuery{}, errors.New("entitlement")
	}
	return AdminUserQuery{p, s, v.Get("q"), v.Get("status"), v.Get("entitlement_status")}, nil
}
func parseDateRange(v url.Values, allowed map[string]bool) (string, string, error) {
	for k := range v {
		if !allowed[k] || len(v[k]) != 1 {
			return "", "", errors.New("query")
		}
	}
	from, to := v.Get("from"), v.Get("to")
	if from == "" {
		from = time.Now().UTC().AddDate(0, 0, -29).Format("2006-01-02")
	}
	if to == "" {
		to = time.Now().UTC().Format("2006-01-02")
	}
	f, e := time.Parse("2006-01-02", from)
	if e != nil {
		return "", "", e
	}
	t, e := time.Parse("2006-01-02", to)
	if e != nil || t.Before(f) || t.Sub(f) > 366*24*time.Hour {
		return "", "", errors.New("date range")
	}
	return from, to, nil
}
func parseUsageQuery(v url.Values) (UsageQuery, error) {
	f, t, e := parseDateRange(v, map[string]bool{"from": true, "to": true})
	return UsageQuery{f, t}, e
}
func parseAdminOrderQuery(v url.Values, user bool) (AdminOrderQuery, error) {
	allowed := map[string]bool{"page": true, "page_size": true, "status": true, "from": true, "to": true}
	p, s, e := parsePage(v, allowed)
	if e != nil {
		return AdminOrderQuery{}, e
	}
	f, t, e := parseDateRange(v, allowed)
	if e != nil {
		return AdminOrderQuery{}, e
	}
	if x := v.Get("status"); x != "" && !map[string]bool{"paid": true, "success": true, "pending": true, "refunded": true, "revoked": true, "cancelled": true, "unknown": true}[x] {
		return AdminOrderQuery{}, errors.New("status")
	}
	return AdminOrderQuery{p, s, v.Get("status"), f, t}, nil
}
