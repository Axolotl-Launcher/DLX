package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type adminHTTPStore struct{ fakeStore }

func (adminHTTPStore) AdminOverview(context.Context) (AdminOverview, error) {
	return AdminOverview{Users: AdminUserCounts{Total: 3}}, nil
}
func (adminHTTPStore) ListAdminUsers(_ context.Context, q AdminUserQuery) (AdminUserPage, error) {
	return AdminUserPage{Items: []AdminUser{}, Page: q.Page, PageSize: q.PageSize, Total: 3}, nil
}
func (adminHTTPStore) GetAdminUser(context.Context, string) (AdminUserDetail, error) {
	return AdminUserDetail{AdminUser: AdminUser{ID: "00000000-0000-0000-0000-000000000001"}}, nil
}
func (adminHTTPStore) AdminUserUsage(context.Context, string, UsageQuery) (UsageSummary, error) {
	return UsageSummary{Days: []UsageDay{}}, nil
}
func (adminHTTPStore) ListAdminOrders(_ context.Context, _ string, q AdminOrderQuery) (AdminOrderPage, error) {
	return AdminOrderPage{Items: []AdminOrder{}, Page: q.Page, PageSize: q.PageSize}, nil
}

func TestAdminOverviewRequiresToken(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/admin/overview", nil)
	w := httptest.NewRecorder()
	store := &adminHTTPStore{}
	(&Server{Store: store, AdminToken: "secret"}).Routes().ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", w.Code)
	}
}
func TestAdminUsersValidatesPagination(t *testing.T) {
	for _, query := range []string{"?page=0", "?page_size=101", "?page=1&page=2", "?unexpected=x"} {
		r := httptest.NewRequest(http.MethodGet, "/admin/users"+query, nil)
		r.Header.Set("X-Admin-Token", "secret")
		w := httptest.NewRecorder()
		store := &adminHTTPStore{}
		(&Server{Store: store, AdminToken: "secret"}).Routes().ServeHTTP(w, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("query=%s status=%d", query, w.Code)
		}
	}
}
func TestAdminUserPathValidatesUUID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/admin/users/not-a-uuid", nil)
	r.Header.Set("X-Admin-Token", "secret")
	w := httptest.NewRecorder()
	store := &adminHTTPStore{}
	(&Server{Store: store, AdminToken: "secret"}).Routes().ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
func TestAdminDTODoesNotContainSecrets(t *testing.T) {
	result := AdminUserDetail{AdminUser: AdminUser{ID: "id", Email: nil, ActiveAPIKey: &AdminAPIKey{Status: "active", CreatedAt: time.Now()}}}
	bodyBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	body := strings.TrimSpace(string(bodyBytes))
	for _, secret := range []string{"secret_hash", "ciphertext", "digest", "raw_payload", "afdian_user_id", "out_trade_no"} {
		if strings.Contains(body, secret) {
			t.Fatalf("secret field %q in DTO: %s", secret, body)
		}
	}
}
