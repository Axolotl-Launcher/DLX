package entitlement

import "testing"

func TestNetPaidFenUsesFinalOrderStates(t *testing.T) {
	orders := []Order{{ActualPaidFen: 490, Status: "refunded"}, {ActualPaidFen: 500, Status: "success"}, {ActualPaidFen: 999, Status: "pending"}}
	if got := NetPaidFen(orders); got != 500 {
		t.Fatalf("got %d, want 500", got)
	}
}
