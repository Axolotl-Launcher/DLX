package entitlement

import (
	"context"
	"errors"
	"time"
)

// Fen is the smallest currency unit used by the gateway. It deliberately
// prevents callers from passing decimal currency into entitlement operations.
type Fen int64

type CDKStatus string

const (
	CDKActive   CDKStatus = "active"
	CDKRedeemed CDKStatus = "redeemed"
	CDKRevoked  CDKStatus = "revoked"
	CDKExpired  CDKStatus = "expired"
)

type CDKBatch struct {
	ID        string
	Name      string
	AmountFen Fen
	Quantity  int
	ExpiresAt *time.Time
}

type CDK struct {
	ID         string
	BatchID    string
	Digest     string // HMAC digest; never the plaintext CDK.
	AmountFen  Fen
	Status     CDKStatus
	RedeemedBy string
	RedeemedAt *time.Time
}

type LedgerEntry struct {
	ID             int64
	UserID         string
	AmountFen      Fen
	SourceType     string
	SourceID       string
	IdempotencyKey string
	CreatedAt      time.Time
}

// CDKRepository is intentionally separate from the existing API Store
// interface. It describes the atomic redemption boundary without adding
// methods to or otherwise changing that compatibility interface.
type CDKRepository interface {
	RedeemCDK(ctx context.Context, digest, userID string, at time.Time, thresholdFen int64) (LedgerEntry, error)
}

var ErrInvalidCDKTransition = errors.New("invalid CDK status transition")
var ErrCDKNotFound = errors.New("CDK not found")
var ErrCDKUsed = errors.New("CDK already used")

// ValidCDKTransition permits only one terminal transition and makes repeated
// redemption impossible at the domain boundary.
func ValidCDKTransition(from, to CDKStatus) bool {
	return from == CDKActive && (to == CDKRedeemed || to == CDKRevoked || to == CDKExpired)
}

func ValidateBatch(batch CDKBatch) error {
	if batch.AmountFen <= 0 {
		return errors.New("CDK amount must be positive")
	}
	if batch.Quantity <= 0 {
		return errors.New("CDK quantity must be positive")
	}
	return nil
}

// LedgerBalance sums immutable movements, clamping a corrupt negative result
// to zero so eligibility cannot accidentally be granted from signed overflow
// or malformed historical data.
func LedgerBalance(entries []LedgerEntry) Fen {
	var total Fen
	for _, entry := range entries {
		total += entry.AmountFen
	}
	if total < 0 {
		return 0
	}
	return total
}
