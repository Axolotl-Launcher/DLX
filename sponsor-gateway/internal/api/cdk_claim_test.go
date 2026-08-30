package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/OwO-Network/DLX/sponsor-gateway/internal/entitlement"
)

type cdkClaimStore struct {
	fakeStore
	amount entitlement.Fen
	err    error
	claims int
	got    struct {
		digest, userID, recoveryHash string
		at                           time.Time
		threshold                    int64
	}
}

func (s *cdkClaimStore) ClaimCDK(_ context.Context, cdkDigest, userID, recoveryHash string, at time.Time, thresholdFen int64) (entitlement.Fen, error) {
	s.claims++
	s.got.digest, s.got.userID, s.got.recoveryHash = cdkDigest, userID, recoveryHash
	s.got.at, s.got.threshold = at, thresholdFen
	if s.err != nil {
		return 0, s.err
	}
	return s.amount, nil
}

func TestCdkClaimCreatesAccountAndReturnsLoginCode(t *testing.T) {
	h := testHasher(t)
	st := &cdkClaimStore{amount: 990}
	req := httptest.NewRequest(http.MethodPost, "/cdk/claim", strings.NewReader(`{"cdk":"cdk_abcdef"}`))
	rec := httptest.NewRecorder()
	(&Server{Store: st, Hasher: h}).Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		LoginCode string `json:"login_code"`
		AmountFen int64  `json:"amount_fen"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(body.LoginCode, "axl_login_") || body.AmountFen != 990 {
		t.Fatalf("unexpected body: %+v", body)
	}
	if st.claims != 1 {
		t.Fatalf("claims=%d", st.claims)
	}
	if st.got.digest != h.Hash("cdk_abcdef") {
		t.Fatalf("digest mismatch: %s", st.got.digest)
	}
	if st.got.userID == "" {
		t.Fatal("no user id generated")
	}
	if st.got.recoveryHash != h.Hash("recovery:"+body.LoginCode) {
		t.Fatal("recovery hash does not match the returned login code")
	}
}

func TestCdkClaimRejectsInvalidBody(t *testing.T) {
	st := &cdkClaimStore{}
	for _, payload := range []string{`{}`, `{"cdk":""}`, `{"cdk":"   "}`, `not-json`, `{"cdk":"x"}{}`} {
		req := httptest.NewRequest(http.MethodPost, "/cdk/claim", strings.NewReader(payload))
		rec := httptest.NewRecorder()
		(&Server{Store: st, Hasher: testHasher(t)}).Routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("payload=%q status=%d", payload, rec.Code)
		}
	}
	if st.claims != 0 {
		t.Fatalf("store called %d times", st.claims)
	}
}

func TestCdkClaimUnavailableCDK(t *testing.T) {
	for _, err := range []error{entitlement.ErrCDKNotFound, entitlement.ErrCDKUsed} {
		st := &cdkClaimStore{err: err}
		req := httptest.NewRequest(http.MethodPost, "/cdk/claim", strings.NewReader(`{"cdk":"cdk_abcdef"}`))
		rec := httptest.NewRecorder()
		(&Server{Store: st, Hasher: testHasher(t)}).Routes().ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Errorf("err=%v status=%d", err, rec.Code)
		}
	}
}

func TestCdkClaimRequiresHasher(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/cdk/claim", strings.NewReader(`{"cdk":"cdk_abcdef"}`))
	rec := httptest.NewRecorder()
	(&Server{Store: &cdkClaimStore{}}).Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestCdkClaimMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/cdk/claim", nil)
	rec := httptest.NewRecorder()
	(&Server{Store: &cdkClaimStore{}, Hasher: testHasher(t)}).Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestCdkClaimWorksWithoutSession(t *testing.T) {
	// No session cookie is set; a CDK claim must still succeed like an order claim.
	h := testHasher(t)
	st := &cdkClaimStore{amount: 500}
	req := httptest.NewRequest(http.MethodPost, "/cdk/claim", strings.NewReader(`{"cdk":"cdk_gift"}`))
	rec := httptest.NewRecorder()
	(&Server{Store: st, Hasher: h}).Routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if st.got.userID != "" && st.got.recoveryHash == "" {
		t.Fatal("claim did not create a recovery account")
	}
}