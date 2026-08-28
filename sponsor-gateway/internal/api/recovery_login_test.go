package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/OwO-Network/DLX/sponsor-gateway/internal/auth"
)

type recoveryStore struct {
	fakeStore
	recoveryHash string
	userID       string
}

func (s *recoveryStore) FindUserByRecoveryHash(_ context.Context, hash string) (string, bool) {
	return s.userID, s.recoveryHash == hash
}
func TestRecoveryLoginSetsSecureHttpOnlySession(t *testing.T) {
	hasher := testHasher(t)
	code, digest, err := auth.NewRecoveryCode(hasher)
	if err != nil {
		t.Fatal(err)
	}
	sessions, err := auth.NewSessions("01234567890123456789012345678901", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	store := &recoveryStore{recoveryHash: digest, userID: "user-1"}
	server := &Server{Store: store, Hasher: hasher, Sessions: sessions, CookieDomain: ".axlmc.org"}
	request := httptest.NewRequest(http.MethodPost, "/auth/recovery-login", strings.NewReader("{\"login_code\":\""+code+"\"}"))
	recorder := httptest.NewRecorder()
	server.Routes().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || !cookies[0].Secure || !strings.Contains(recorder.Header().Get("Set-Cookie"), "Domain=axlmc.org") {
		t.Fatalf("unsafe or missing session cookie: %#v", cookies)
	}
}

func TestMeReturnsOnlyAccountEntitlement(t *testing.T) {
	sessions, err := auth.NewSessions("01234567890123456789012345678901", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := sessions.Issue("user-1", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	server := &Server{Store: &fakeStore{paid: 990, status: "granted"}, Sessions: sessions}
	request := httptest.NewRequest(http.MethodGet, "/me", nil)
	request.AddCookie(&http.Cookie{Name: "axl_sponsor_session", Value: token})
	recorder := httptest.NewRecorder()
	server.Routes().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "lifetime_paid_fen") || strings.Contains(recorder.Body.String(), "api_key") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}
