package afdian

import "testing"

func TestVerifiedOrderCarriesServerAuthoritativeFields(t *testing.T) {
	order := VerifiedOrder{OutTradeNo: "order-1", AfdianUserID: "afdian-1", ActualPaidFen: 990, Status: "paid"}
	if order.ActualPaidFen != 990 || order.AfdianUserID == "" {
		t.Fatal("incomplete verified order")
	}
}
