package translate

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientUsesPrivateDLXEndpointAndAdaptsResponse(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/translate" {
			t.Errorf("path=%s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer internal-token" {
			t.Error("missing internal token")
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "\"text\":\"Hello\"") {
			t.Error("missing compatible request text")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"code\":200,\"data\":\"你好\",\"source_lang\":\"EN\"}"))
	}))
	defer upstream.Close()
	client := &Client{BaseURL: upstream.URL, InternalToken: "internal-token"}
	response, err := client.Translate(context.Background(), Request{Text: "Hello", SourceLang: "EN", TargetLang: "ZH"})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Translations) != 1 || response.Translations[0].Text != "你好" {
		t.Fatalf("unexpected response: %#v", response)
	}
}
