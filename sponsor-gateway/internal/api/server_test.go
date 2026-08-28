package api

import (
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/auth"
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/translate"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeStore struct {
	key    KeyRecord
	paid   int64
	status string
}

func (s *fakeStore) FindKey(hash string) (KeyRecord, bool)   { return s.key, s.key.Hash == hash }
func (s *fakeStore) UserStatus(string) (int64, string, bool) { return s.paid, s.status, true }
func (s *fakeStore) TouchKey(string, time.Time) error        { return nil }
func testHasher(t *testing.T) *auth.Hasher {
	t.Helper()
	h, err := auth.NewHasher("01234567890123456789012345678901")
	if err != nil {
		t.Fatal(err)
	}
	return h
}
func TestTranslateRejectsIneligibleSponsorBeforeUpstream(t *testing.T) {
	h := testHasher(t)
	key, digest, err := auth.NewAPIKey("key-1", h)
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeStore{key: KeyRecord{ID: "key-1", UserID: "user-1", Hash: digest, Status: "active"}, paid: 989, status: "granted"}
	server := &Server{Store: store, Hasher: h, DLX: &translate.Client{BaseURL: "http://127.0.0.1:1"}}
	payload := "{\"text\":\"secret text\",\"source_lang\":\"EN\",\"target_lang\":\"ZH\"}"
	request := httptest.NewRequest(http.MethodPost, "/v1/translate", strings.NewReader(payload))
	request.Header.Set("Authorization", "Bearer "+key)
	recorder := httptest.NewRecorder()
	server.Routes().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "SPONSORSHIP_REQUIRED") {
		t.Fatal("missing stable entitlement code")
	}
}
