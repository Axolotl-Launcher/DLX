package afdian

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestQueryOrderUsesSignedFormAndActualAmount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/query-order" {
			t.Fatal("wrong path")
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("user_id") != "creator" || r.Form.Get("sign") == "" || !strings.Contains(r.Form.Get("params"), "order-1") {
			t.Fatal("missing signed fields")
		}
		w.Write([]byte(`{"ec":200,"data":{"out_trade_no":"order-1","user_id":"supporter-1","total_amount":"9.90","status":2}}`))
	}))
	defer server.Close()
	client := &Client{UserID: "creator", Token: "token", BaseURL: server.URL, Now: func() time.Time { return time.Unix(100, 0) }}
	order, err := client.VerifyOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatal(err)
	}
	if order.ActualPaidFen != 990 || order.AfdianUserID != "supporter-1" || order.Status != "paid" {
		t.Fatalf("unexpected order %#v", order)
	}
}
