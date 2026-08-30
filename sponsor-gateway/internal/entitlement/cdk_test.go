package entitlement

import "testing"

func TestValidCDKTransitionHasSingleTerminalUse(t *testing.T) {
	for _, to := range []CDKStatus{CDKRedeemed, CDKRevoked, CDKExpired} {
		if !ValidCDKTransition(CDKActive, to) { t.Errorf("active -> %s should be valid", to) }
	}
	for _, from := range []CDKStatus{CDKRedeemed, CDKRevoked, CDKExpired} {
		if ValidCDKTransition(from, CDKRedeemed) { t.Errorf("%s -> redeemed should be invalid", from) }
	}
	if ValidCDKTransition(CDKActive, CDKActive) { t.Fatal("active -> active should be invalid") }
}

func TestValidateBatchUsesPositiveFenAndQuantity(t *testing.T) {
	if err := ValidateBatch(CDKBatch{AmountFen: 1, Quantity: 1}); err != nil { t.Fatal(err) }
	if err := ValidateBatch(CDKBatch{AmountFen: 0, Quantity: 1}); err == nil { t.Fatal("zero amount must be rejected") }
	if err := ValidateBatch(CDKBatch{AmountFen: 1, Quantity: 0}); err == nil { t.Fatal("zero quantity must be rejected") }
}

func TestLedgerBalanceSumsFenAndClampsNegative(t *testing.T) {
	if got := LedgerBalance([]LedgerEntry{{AmountFen: 500}, {AmountFen: -125}}); got != 375 { t.Fatalf("got %d, want 375", got) }
	if got := LedgerBalance([]LedgerEntry{{AmountFen: -1}}); got != 0 { t.Fatalf("got %d, want 0", got) }
}
