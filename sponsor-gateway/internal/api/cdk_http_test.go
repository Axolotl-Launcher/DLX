package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OwO-Network/DLX/sponsor-gateway/internal/entitlement"
)

type cdkHTTPStore struct {
	fakeStore
	created int
}

func (s *cdkHTTPStore) CreateCDKs(_ context.Context, amount int64, quantity int, note string, digests []string) (string, error) {
	s.created = quantity
	if amount != 42 || note != "test" || len(digests) != quantity {
		return "", context.Canceled
	}
	return "batch-1", nil
}

func (s *cdkHTTPStore) ListCDKs(context.Context) ([]entitlement.CDK, error) {
	return []entitlement.CDK{{ID: "cdk-id", BatchID: "batch-id", Digest: "secret-digest", AmountFen: 42, Status: entitlement.CDKRedeemed, RedeemedBy: "user-secret"}}, nil
}

func TestAdminCDKCreationUsesConfiguredTokenAndReturnsPlaintextOnce(t *testing.T) {
	h := testHasher(t)
	st := &cdkHTTPStore{}
	srv := &Server{Store: st, Hasher: h, AdminToken: "configured"}
	req := httptest.NewRequest(http.MethodPost, "/admin/cdks", strings.NewReader(`{"amount_fen":42,"quantity":2,"note":"test"}`))
	req.Header.Set("X-Admin-Token", "configured")
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if st.created != 2 {
		t.Fatalf("created=%d", st.created)
	}
}

func TestAdminCDKListRedactsSensitiveFields(t *testing.T) {
	srv := &Server{Store: &cdkHTTPStore{}, AdminToken: "configured"}
	req := httptest.NewRequest(http.MethodGet, "/admin/cdks", nil)
	req.Header.Set("Authorization", "Bearer configured")
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if strings.Contains(body, "secret-digest") || strings.Contains(body, "user-secret") || strings.Contains(body, "digest") || strings.Contains(body, "redeemed_by") {
		t.Fatalf("sensitive CDK fields leaked: %s", body)
	}
	var result []adminCDKDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil || len(result) != 1 || result[0].AmountFen != 42 {
		t.Fatalf("unexpected DTO: %#v error=%v", result, err)
	}
}

func TestAdminCDKRejectsTrailingJSON(t *testing.T) {
	srv := &Server{Store: &cdkHTTPStore{}, Hasher: testHasher(t), AdminToken: "configured"}
	req := httptest.NewRequest(http.MethodPost, "/admin/cdks", strings.NewReader(`{"amount_fen":42,"quantity":1,"note":"test"}{}`))
	req.Header.Set("X-Admin-Token", "configured")
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminCDKPreflightAllowsAdminToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/admin/cdks", nil)
	req.Header.Set("Origin", "https://sponsor.axlmc.org")
	rec := httptest.NewRecorder()
	(&Server{}).Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent || !strings.Contains(rec.Header().Get("Access-Control-Allow-Headers"), "X-Admin-Token") {
		t.Fatalf("status=%d headers=%q", rec.Code, rec.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestAdminCDKRejectsWrongToken(t *testing.T) {
	srv := &Server{Store: &cdkHTTPStore{}, AdminToken: "configured"}
	req := httptest.NewRequest(http.MethodGet, "/admin/cdks", nil)
	req.Header.Set("X-Admin-Token", "wrong")
	rec := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}
