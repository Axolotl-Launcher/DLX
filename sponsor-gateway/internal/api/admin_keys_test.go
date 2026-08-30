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

type adminAPIKeysStore struct {
	adminHTTPStore
	Page     AdminAPIKeyPage
	Captured AdminAPIKeyQuery
}

func (s *adminAPIKeysStore) ListAdminAPIKeys(_ context.Context, q AdminAPIKeyQuery) (AdminAPIKeyPage, error) {
	s.Captured = q
	return s.Page, nil
}

func TestAdminAPIKeysRequiresTokenAndMethod(t *testing.T) {
	store := &adminAPIKeysStore{}
	srv := &Server{Store: store, AdminToken: "configured"}
	for _, tc := range []struct {
		method string
		header bool
		want   int
	}{
		{http.MethodGet, false, http.StatusUnauthorized},
		{http.MethodPost, true, http.StatusMethodNotAllowed},
		{http.MethodDelete, true, http.StatusMethodNotAllowed},
	} {
		req := httptest.NewRequest(tc.method, "/admin/api-keys", nil)
		if tc.header {
			req.Header.Set("X-Admin-Token", "configured")
		}
		w := httptest.NewRecorder()
		srv.Routes().ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s header=%v status=%d want=%d", tc.method, tc.header, w.Code, tc.want)
		}
	}
}

func TestAdminAPIKeysAcceptsBearerToken(t *testing.T) {
	store := &adminAPIKeysStore{}
	req := httptest.NewRequest(http.MethodGet, "/admin/api-keys", nil)
	req.Header.Set("Authorization", "Bearer configured")
	w := httptest.NewRecorder()
	(&Server{Store: store, AdminToken: "configured"}).Routes().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAdminAPIKeysRejectsWrongToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/api-keys", nil)
	req.Header.Set("X-Admin-Token", "wrong")
	w := httptest.NewRecorder()
	(&Server{Store: &adminAPIKeysStore{}, AdminToken: "configured"}).Routes().ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestAdminAPIKeysValidatesQuery(t *testing.T) {
	for _, query := range []string{"?page=0", "?page_size=101", "?page=1&page=2", "?unexpected=x", "?status=bogus", "?q=" + strings.Repeat("a", 101)} {
		req := httptest.NewRequest(http.MethodGet, "/admin/api-keys"+query, nil)
		req.Header.Set("X-Admin-Token", "secret")
		w := httptest.NewRecorder()
		(&Server{Store: &adminAPIKeysStore{}, AdminToken: "secret"}).Routes().ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("query=%s status=%d", query, w.Code)
		}
	}
}

func TestAdminAPIKeysPassesFiltersAndPagination(t *testing.T) {
	store := &adminAPIKeysStore{Page: AdminAPIKeyPage{Items: []AdminAPIKeyRecord{}, Page: 2, PageSize: 50, Total: 99}}
	req := httptest.NewRequest(http.MethodGet, "/admin/api-keys?page=2&page_size=50&status=active&q=user%40example.com", nil)
	req.Header.Set("X-Admin-Token", "secret")
	w := httptest.NewRecorder()
	(&Server{Store: store, AdminToken: "secret"}).Routes().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if store.Captured.Page != 2 || store.Captured.PageSize != 50 || store.Captured.Status != "active" || store.Captured.Q != "user@example.com" {
		t.Fatalf("query not forwarded: %+v", store.Captured)
	}
	var body AdminAPIKeyPage
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Page != 2 || body.PageSize != 50 || body.Total != 99 {
		t.Fatalf("page not echoed: %+v", body)
	}
}

func TestAdminAPIKeyDTOIsRedacted(t *testing.T) {
	email := "a***@example.com" // store layer masks emails before they reach the DTO
	used := time.Now()
	result := AdminAPIKeyRecord{ID: "00000000-0000-0000-0000-000000000001", UserID: "00000000-0000-0000-0000-000000000002", UserEmail: &email, Status: "active", CreatedAt: time.Now(), LastUsedAt: &used}
	bodyBytes, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	body := strings.TrimSpace(string(bodyBytes))
	for _, secret := range []string{"prefix", "secret_hash", "ciphertext", "digest", "api_key"} {
		if strings.Contains(body, secret) {
			t.Fatalf("secret field %q in DTO: %s", secret, body)
		}
	}
}