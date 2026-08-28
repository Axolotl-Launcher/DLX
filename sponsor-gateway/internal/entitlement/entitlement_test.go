package entitlement

import "testing"

func TestLifetimeThresholdUsesFen(t *testing.T) {
	const paid = 490 + 500
	if !Eligible(paid, 990) {
		t.Fatal("4.90 + 5.00 must grant permanent eligibility")
	}
	if RemainingFen(paid, 990) != 0 {
		t.Fatal("threshold amount has no remainder")
	}
	if Eligible(989, 990) {
		t.Fatal("one fen below threshold must not be eligible")
	}
	if RemainingFen(989, 990) != 1 {
		t.Fatal("unexpected remaining amount")
	}
}
