// Package afdian defines the verified-order boundary. The real client must be
// configured from the team's Afdian developer credentials and must never trust
// values supplied by a browser other than the order number.
package afdian

import (
	"context"
	"errors"
)

type VerifiedOrder struct {
	OutTradeNo    string
	AfdianUserID  string
	ActualPaidFen int64
	Status        string
}
type Verifier interface {
	VerifyOrder(ctx context.Context, outTradeNo string) (VerifiedOrder, error)
}

var ErrUnavailable = errors.New("Afdian verification is unavailable")
