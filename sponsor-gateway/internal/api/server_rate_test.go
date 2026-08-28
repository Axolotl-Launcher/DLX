package api

import (
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/auth"
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/translate"
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/usage"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTranslateRateLimitPreventsSecondUpstreamCall(t *testing.T) {
	h := testHasher(t)
	key, digest, err := auth.NewAPIKey("key-1", h)
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeStore{key: KeyRecord{ID: "key-1", UserID: "user-1", Hash: digest, Status: "active"}, paid: 990, status: "granted"}
	server := &Server{Store: store, Hasher: h, DLX: &translate.Client{BaseURL: "http://127.0.0.1:1"}, Limiter: usage.NewFixedWindow(1, time.Minute)}
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/translate", strings.NewReader("{\"text\":\"safe\",\"target_lang\":\"ZH\"}"))
		req.Header.Set("Authorization", "Bearer "+key)
		rec := httptest.NewRecorder()
		server.Routes().ServeHTTP(rec, req)
		if i == 0 && rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("first status=%d", rec.Code)
		}
		if i == 1 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("second status=%d body=%s", rec.Code, rec.Body.String())
		}
	}
}
