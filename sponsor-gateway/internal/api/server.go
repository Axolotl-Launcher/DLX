package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/OwO-Network/DLX/sponsor-gateway/internal/afdian"
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/auth"
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/entitlement"
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/translate"
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/usage"
)

type KeyRecord struct {
	ID, UserID, Hash, Status string
	LastUsedAt               time.Time
}
type Store interface {
	FindKey(hash string) (KeyRecord, bool)
	UserStatus(userID string) (paidFen int64, status string, ok bool)
	TouchKey(id string, at time.Time) error
}
type UsageRecorder interface {
	RecordUsage(userID string, inputChars int, errorOccurred bool, at time.Time) error
}
type RecoveryUserFinder interface {
	FindUserByRecoveryHash(ctx context.Context, recoveryHash string) (string, bool)
}
type ClaimCreator interface {
	CreateVerifiedClaim(ctx context.Context, userID string, recoveryHash string, order afdian.VerifiedOrder, thresholdFen int64) error
}
type KeyManager interface {
	RotateKey(ctx context.Context, userID, id, prefix, hash string) error
	RevokeKey(ctx context.Context, userID string) error
}
type Server struct {
	Store         Store
	DLX           *translate.Client
	Hasher        *auth.Hasher
	Sessions      *auth.Sessions
	Afdian        afdian.Verifier
	CookieDomain  string
	Limiter       usage.Limiter
	ReadyCheck    func(context.Context) error
	UpstreamSlots chan struct{}
	ThresholdFen  int64
	MaxTextChars  int
}
type errorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, 405, "METHOD_NOT_ALLOWED", "method not allowed", "")
			return
		}
		writeJSON(w, 200, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/readyz", s.ready)
	mux.HandleFunc("/auth/recovery-login", s.recoveryLogin)
	mux.HandleFunc("/afdian/claim", s.claimOrder)
	mux.HandleFunc("/me/api-key", s.apiKey)
	mux.HandleFunc("/me", s.me)
	mux.HandleFunc("/v1/translate", s.handleTranslate)
	mux.HandleFunc("/v1/account", s.account)
	return sponsorCORS(mux)
}

func sponsorCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") == "https://sponsor.axlmc.org" {
			w.Header().Set("Access-Control-Allow-Origin", "https://sponsor.axlmc.org")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) sessionUser(r *http.Request) (string, bool) {
	if s.Sessions == nil {
		return "", false
	}
	cookie, err := r.Cookie("axl_sponsor_session")
	if err != nil {
		return "", false
	}
	userID, err := s.Sessions.Validate(cookie.Value, time.Now().UTC())
	return userID, err == nil
}
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", "")
		return
	}
	userID, ok := s.sessionUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "login required", "")
		return
	}
	paid, status, ok := s.Store.UserStatus(userID)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "account unavailable", "")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"lifetime_paid_fen": paid, "entitlement_status": status, "threshold_fen": s.threshold()})
}

func (s *Server) apiKey(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.sessionUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "login required", "")
		return
	}
	manager, ok := s.Store.(KeyManager)
	if !ok || s.Hasher == nil {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "key service unavailable", "")
		return
	}
	switch r.Method {
	case http.MethodPost:
		id, err := auth.NewID()
		if err != nil {
			writeError(w, 503, "SERVICE_UNAVAILABLE", "key service unavailable", "")
			return
		}
		key, digest, err := auth.NewAPIKey(id, s.Hasher)
		if err != nil {
			writeError(w, 503, "SERVICE_UNAVAILABLE", "key service unavailable", "")
			return
		}
		if err = manager.RotateKey(r.Context(), userID, id, "axl_live_"+id, digest); err != nil {
			writeError(w, 403, "SPONSORSHIP_REQUIRED", "permanent sponsorship is required", "")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"api_key": key, "message": "Copy and store this API Key now; it will not be shown again."})
	case http.MethodDelete:
		if err := manager.RevokeKey(r.Context(), userID); err != nil {
			writeError(w, 503, "SERVICE_UNAVAILABLE", "key service unavailable", "")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", "")
	}
}

func (s *Server) claimOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", "")
		return
	}
	if s.Store == nil || s.Hasher == nil || s.Afdian == nil {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "claim service is not ready", "")
		return
	}
	creator, ok := s.Store.(ClaimCreator)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "claim service is not ready", "")
		return
	}
	var request struct {
		OrderNo string `json:"order_no"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&request); err != nil || strings.TrimSpace(request.OrderNo) == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ORDER", "invalid order number", "")
		return
	}
	order, err := s.Afdian.VerifyOrder(r.Context(), strings.TrimSpace(request.OrderNo))
	if err != nil || order.OutTradeNo != strings.TrimSpace(request.OrderNo) || order.AfdianUserID == "" || order.ActualPaidFen < 0 || (order.Status != "paid" && order.Status != "success") {
		writeError(w, http.StatusForbidden, "ORDER_VERIFICATION_FAILED", "order could not be verified", "")
		return
	}
	userID, err := auth.NewID()
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "claim service unavailable", "")
		return
	}
	code, digest, err := auth.NewRecoveryCode(s.Hasher)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "claim service unavailable", "")
		return
	}
	if err = creator.CreateVerifiedClaim(r.Context(), userID, digest, order, s.threshold()); err != nil {
		writeError(w, http.StatusConflict, "ORDER_ALREADY_CLAIMED", "please contact support", "")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"login_code": code, "message": "Copy and store this login code now; it will not be shown again."})
}

func (s *Server) recoveryLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed", "")
		return
	}
	if s.Store == nil || s.Hasher == nil || s.Sessions == nil {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "login service is not ready", "")
		return
	}
	finder, ok := s.Store.(RecoveryUserFinder)
	if !ok {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "login service is not ready", "")
		return
	}
	var request struct {
		LoginCode string `json:"login_code"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_LOGIN_CODE", "invalid login code", "")
		return
	}
	code, err := auth.ParseRecoveryCode(request.LoginCode)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_LOGIN_CODE", "invalid login code", "")
		return
	}
	userID, found := finder.FindUserByRecoveryHash(r.Context(), s.Hasher.Hash("recovery:"+code))
	if !found {
		writeError(w, http.StatusUnauthorized, "INVALID_LOGIN_CODE", "invalid login code", "")
		return
	}
	token, err := s.Sessions.Issue(userID, time.Now().UTC())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "login service unavailable", "")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "axl_sponsor_session", Value: token, Path: "/", Domain: s.CookieDomain, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, MaxAge: 7 * 24 * 60 * 60})
	writeJSON(w, http.StatusOK, map[string]string{"status": "authenticated"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, 405, "METHOD_NOT_ALLOWED", "method not allowed", "")
		return
	}
	if s.Store == nil || s.DLX == nil || s.Hasher == nil {
		writeError(w, 503, "SERVICE_UNAVAILABLE", "service is not ready", "")
		return
	}
	if s.ReadyCheck != nil && s.ReadyCheck(r.Context()) != nil {
		writeError(w, 503, "SERVICE_UNAVAILABLE", "service dependencies are not ready", "")
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ready"})
}
func (s *Server) authenticate(r *http.Request) (KeyRecord, string) {
	key, err := auth.ParseBearer(r.Header.Get("Authorization"))
	if err != nil || s.Store == nil || s.Hasher == nil {
		return KeyRecord{}, ""
	}
	record, ok := s.Store.FindKey(s.Hasher.Hash(key))
	if !ok || record.Status != "active" || !s.Hasher.Equal(record.Hash, key) {
		return KeyRecord{}, ""
	}
	return record, record.UserID
}
func (s *Server) handleTranslate(w http.ResponseWriter, r *http.Request) {
	requestID := newRequestID()
	w.Header().Set("X-Request-ID", requestID)
	if r.Method != http.MethodPost {
		writeError(w, 405, "METHOD_NOT_ALLOWED", "method not allowed", requestID)
		return
	}
	var payload translate.Request
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20))
	if err := decoder.Decode(&payload); err != nil || strings.TrimSpace(payload.Text) == "" || payload.TargetLang == "" {
		writeError(w, 400, "INVALID_REQUEST", "invalid request payload", requestID)
		return
	}
	maxChars := s.MaxTextChars
	if maxChars == 0 {
		maxChars = 10000
	}
	if len([]rune(payload.Text)) > maxChars {
		writeError(w, 413, "INVALID_REQUEST", "text is too long", requestID)
		return
	}
	record, userID := s.authenticate(r)
	if userID == "" {
		writeError(w, 401, "INVALID_API_KEY", "invalid API key", requestID)
		return
	}
	paid, status, ok := s.Store.UserStatus(userID)
	if !ok || status != "granted" || !entitlement.Eligible(paid, s.threshold()) {
		writeError(w, 403, "SPONSORSHIP_REQUIRED", "lifetime sponsorship threshold is required", requestID)
		return
	}
	if s.Limiter != nil && !s.Limiter.Allow(record.ID, time.Now().UTC()) {
		writeError(w, 429, "RATE_LIMITED", "rate limit exceeded", requestID)
		return
	}
	if s.UpstreamSlots != nil {
		select {
		case s.UpstreamSlots <- struct{}{}:
			defer func() { <-s.UpstreamSlots }()
		default:
			writeError(w, 503, "UPSTREAM_BUSY", "translation upstream is at capacity", requestID)
			return
		}
	}
	response, err := s.DLX.Translate(r.Context(), payload)
	now := time.Now().UTC()
	if touchErr := s.Store.TouchKey(record.ID, now); touchErr != nil {
		writeError(w, 503, "SERVICE_UNAVAILABLE", "usage accounting unavailable", requestID)
		return
	}
	if usageStore, ok := s.Store.(UsageRecorder); ok {
		if usageErr := usageStore.RecordUsage(userID, len([]rune(payload.Text)), err != nil, now); usageErr != nil {
			writeError(w, 503, "SERVICE_UNAVAILABLE", "usage accounting unavailable", requestID)
			return
		}
	}
	if err != nil {
		writeUpstreamError(w, err, requestID)
		return
	}
	writeJSON(w, 200, response)
}
func (s *Server) account(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, 405, "METHOD_NOT_ALLOWED", "method not allowed", "")
		return
	}
	_, userID := s.authenticate(r)
	if userID == "" {
		writeError(w, 401, "INVALID_API_KEY", "invalid API key", "")
		return
	}
	paid, status, _ := s.Store.UserStatus(userID)
	writeJSON(w, 200, map[string]interface{}{"lifetime_paid_fen": paid, "eligible": entitlement.Eligible(paid, s.threshold()), "status": status})
}
func (s *Server) threshold() int64 {
	if s.ThresholdFen <= 0 {
		return entitlement.DefaultThresholdFen
	}
	return s.ThresholdFen
}
func writeUpstreamError(w http.ResponseWriter, err error, requestID string) {
	var upstream *translate.UpstreamError
	if errors.As(err, &upstream) {
		switch upstream.Kind {
		case translate.ErrorTimeout:
			writeError(w, 504, "UPSTREAM_TIMEOUT", "translation upstream timed out", requestID)
		case translate.ErrorBusy:
			writeError(w, 503, "UPSTREAM_BUSY", "translation upstream unavailable", requestID)
		default:
			writeError(w, 503, "SERVICE_UNAVAILABLE", "translation service unavailable", requestID)
		}
		return
	}
	writeError(w, 503, "SERVICE_UNAVAILABLE", "translation service unavailable", requestID)
}
func newRequestID() string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return "request-unavailable"
	}
	return hex.EncodeToString(bytes)
}
func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, code, message, requestID string) {
	writeJSON(w, status, errorBody{Code: code, Message: message, RequestID: requestID})
}
