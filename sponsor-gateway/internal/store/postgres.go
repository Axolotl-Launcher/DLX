// Package store contains durable repository adapters. The adapter intentionally
// depends only on database/sql; production wires it to a PostgreSQL driver.
package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/OwO-Network/DLX/sponsor-gateway/internal/afdian"
	"github.com/OwO-Network/DLX/sponsor-gateway/internal/api"
)

type Postgres struct {
	db      *sql.DB
	timeout time.Duration
}

func NewPostgres(db *sql.DB) *Postgres { return &Postgres{db: db, timeout: 3 * time.Second} }
func (s *Postgres) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.timeout)
}
func (s *Postgres) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

func (s *Postgres) FindKey(hash string) (api.KeyRecord, bool) {
	ctx, cancel := s.context()
	defer cancel()
	row := s.db.QueryRowContext(ctx, `SELECT id::text, user_id::text, secret_hash, status, COALESCE(last_used_at, 'epoch'::timestamptz) FROM api_keys WHERE secret_hash = $1 LIMIT 1`, hash)
	var record api.KeyRecord
	if err := row.Scan(&record.ID, &record.UserID, &record.Hash, &record.Status, &record.LastUsedAt); err != nil {
		return api.KeyRecord{}, false
	}
	return record, true
}
func (s *Postgres) UserStatus(userID string) (int64, string, bool) {
	ctx, cancel := s.context()
	defer cancel()
	row := s.db.QueryRowContext(ctx, `SELECT lifetime_paid_fen, status FROM entitlements WHERE user_id = $1::uuid`, userID)
	var paid int64
	var status string
	if err := row.Scan(&paid, &status); err != nil {
		return 0, "", false
	}
	return paid, status, true
}
func (s *Postgres) TouchKey(id string, at time.Time) error {
	ctx, cancel := s.context()
	defer cancel()
	_, err := s.db.ExecContext(ctx, `UPDATE api_keys SET last_used_at = $2 WHERE id = $1::uuid AND status = 'active'`, id, at)
	return err
}

// RecordUsage atomically aggregates metadata only. It intentionally takes no
// translation body or credential so such material cannot reach usage_daily.
func (s *Postgres) UsageSummary(userID string, since time.Time) (api.UsageSummary, error) {
	ctx, cancel := s.context(); defer cancel()
	rows, err := s.db.QueryContext(ctx, `SELECT d::date, COALESCE(u.request_count,0), COALESCE(u.input_chars,0), COALESCE(u.error_count,0) FROM generate_series($2::date, CURRENT_DATE, interval '1 day') d LEFT JOIN usage_daily u ON u.user_id=$1::uuid AND u.date=d::date ORDER BY d`, userID, since.UTC().Format("2006-01-02")); if err != nil { return api.UsageSummary{}, err }; defer rows.Close()
	result:=api.UsageSummary{Days:make([]api.UsageDay,0,365)}
	for rows.Next(){var day api.UsageDay;var date time.Time;if err:=rows.Scan(&date,&day.RequestCount,&day.InputChars,&day.ErrorCount);err!=nil{return api.UsageSummary{},err};day.Date=date.Format("2006-01-02");result.Days=append(result.Days,day);result.TotalRequestCount+=day.RequestCount;result.TotalInputChars+=day.InputChars;result.TotalErrorCount+=day.ErrorCount}
	return result, rows.Err()
}

func (s *Postgres) RecordUsage(userID string, inputChars int, errorOccurred bool, at time.Time) error {
	ctx, cancel := s.context()
	defer cancel()
	errors := 0
	if errorOccurred {
		errors = 1
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO usage_daily (user_id,date,request_count,input_chars,error_count) VALUES ($1::uuid,$2,1,$3,$4) ON CONFLICT (user_id,date) DO UPDATE SET request_count=usage_daily.request_count+1,input_chars=usage_daily.input_chars+EXCLUDED.input_chars,error_count=usage_daily.error_count+EXCLUDED.error_count`, userID, at.UTC().Format("2006-01-02"), inputChars, errors)
	return err
}

type Entitlement struct {
	LifetimePaidFen int64
	Status          string
}

// RecalculateEntitlement derives eligibility from the immutable order ledger in
// one transaction. A refund/revocation below the threshold suspends active keys.
func (s *Postgres) RecalculateEntitlement(ctx context.Context, userID string, thresholdFen int64) (Entitlement, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Entitlement{}, err
	}
	defer tx.Rollback()
	var paid int64
	err = tx.QueryRowContext(ctx, "SELECT COALESCE(SUM(CASE WHEN o.status IN ('paid','success') THEN o.actual_paid_fen ELSE 0 END),0) FROM afdian_orders o JOIN afdian_identities i ON i.afdian_user_id=o.afdian_user_id WHERE i.user_id=$1::uuid", userID).Scan(&paid)
	if err != nil {
		return Entitlement{}, err
	}
	if paid < 0 {
		paid = 0
	}
	desired := "pending"
	if paid >= thresholdFen {
		desired = "granted"
	}
	_, err = tx.ExecContext(ctx, "INSERT INTO entitlements(user_id,lifetime_paid_fen,status,granted_at,recalculated_at) VALUES($1::uuid,$2,$3,CASE WHEN $3='granted' THEN now() ELSE NULL END,now()) ON CONFLICT(user_id) DO UPDATE SET lifetime_paid_fen=EXCLUDED.lifetime_paid_fen,status=CASE WHEN EXCLUDED.status='granted' THEN 'granted' WHEN entitlements.status='granted' THEN 'suspended' ELSE 'pending' END,granted_at=CASE WHEN EXCLUDED.status='granted' THEN COALESCE(entitlements.granted_at,now()) ELSE entitlements.granted_at END,recalculated_at=now()", userID, paid, desired)
	if err != nil {
		return Entitlement{}, err
	}
	if paid < thresholdFen {
		if _, err = tx.ExecContext(ctx, "UPDATE api_keys SET status='suspended' WHERE user_id=$1::uuid AND status='active'", userID); err != nil {
			return Entitlement{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return Entitlement{}, err
	}
	if paid < thresholdFen && desired == "pending" {
		desired = "suspended"
	}
	return Entitlement{LifetimePaidFen: paid, Status: desired}, nil
}

var ErrIdentityClaimed = errors.New("Afdian identity is already claimed")

// BindAfdianIdentity is conflict-safe: a provider identity can only ever be
// bound to its original local user. It does not reveal that user's identity.
func (s *Postgres) BindAfdianIdentity(ctx context.Context, userID, afdianUserID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var boundUserID string
	err = tx.QueryRowContext(ctx, "INSERT INTO afdian_identities(user_id,afdian_user_id,verified_at) VALUES($1::uuid,$2,now()) ON CONFLICT(afdian_user_id) DO UPDATE SET afdian_user_id=EXCLUDED.afdian_user_id WHERE afdian_identities.user_id=EXCLUDED.user_id RETURNING user_id::text", userID, afdianUserID).Scan(&boundUserID)
	if err == sql.ErrNoRows {
		return ErrIdentityClaimed
	}
	if err != nil {
		return err
	}
	if boundUserID != userID {
		return ErrIdentityClaimed
	}
	return tx.Commit()
}

type NewAPIKey struct {
	ID         string
	Prefix     string
	SecretHash string
}

// RotateAPIKey enforces both permanent entitlement and the first-version
// one-active-key rule. The caller shows plaintext only after this commits.
func (s *Postgres) RotateAPIKey(ctx context.Context, userID string, key NewAPIKey) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status string
	if err = tx.QueryRowContext(ctx, "SELECT status FROM entitlements WHERE user_id=$1::uuid FOR UPDATE", userID).Scan(&status); err != nil {
		return err
	}
	if status != "granted" {
		return errors.New("sponsorship entitlement is not granted")
	}
	if _, err = tx.ExecContext(ctx, "UPDATE api_keys SET status='revoked' WHERE user_id=$1::uuid AND status='active'", userID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO api_keys(id,user_id,prefix,secret_hash,status) VALUES($1::uuid,$2::uuid,$3,$4,'active')", key.ID, userID, key.Prefix, key.SecretHash); err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Postgres) RevokeAPIKey(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE api_keys SET status='revoked' WHERE user_id=$1::uuid AND status='active'", userID)
	return err
}

// CreateClaimUser creates an account only after Afdian ownership has been
// independently verified. Its recovery code digest is the sole login secret.
func (s *Postgres) CreateClaimUser(ctx context.Context, userID, recoveryHash string) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO users(id,status,login_code_hash) VALUES($1::uuid,'active',$2)", userID, recoveryHash)
	return err
}
func (s *Postgres) FindUserByRecoveryHash(ctx context.Context, recoveryHash string) (string, bool) {
	var userID string
	err := s.db.QueryRowContext(ctx, "SELECT id::text FROM users WHERE login_code_hash=$1 AND status='active'", recoveryHash).Scan(&userID)
	return userID, err == nil
}

// CreateVerifiedClaim atomically creates the recovery-code account, binds the
// provider identity, writes the verified order, and derives initial entitlement.
func (s *Postgres) CreateVerifiedClaim(ctx context.Context, userID, recoveryHash string, order afdian.VerifiedOrder, thresholdFen int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "INSERT INTO users(id,status,login_code_hash) VALUES($1::uuid,'active',$2)", userID, recoveryHash); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO afdian_identities(user_id,afdian_user_id,verified_at) VALUES($1::uuid,$2,now())", userID, order.AfdianUserID); err != nil {
		return ErrIdentityClaimed
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO afdian_orders(out_trade_no,afdian_user_id,actual_paid_fen,status,raw_payload,synced_at) VALUES($1,$2,$3,$4,'{}'::jsonb,now())", order.OutTradeNo, order.AfdianUserID, order.ActualPaidFen, order.Status); err != nil {
		return err
	}
	status := "pending"
	if order.ActualPaidFen >= thresholdFen {
		status = "granted"
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO entitlements(user_id,lifetime_paid_fen,status,granted_at,recalculated_at) VALUES($1::uuid,$2,$3,CASE WHEN $3='granted' THEN now() ELSE NULL END,now())", userID, order.ActualPaidFen, status); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Postgres) RotateKey(ctx context.Context, userID, id, prefix, hash string) error {
	return s.RotateAPIKey(ctx, userID, NewAPIKey{ID: id, Prefix: prefix, SecretHash: hash})
}
func (s *Postgres) RevokeKey(ctx context.Context, userID string) error {
	return s.RevokeAPIKey(ctx, userID)
}

// BeginWebhookEvent creates an idempotency record before processing. Payload
// must already be a minimized/redacted JSON envelope, never raw provider PII.
func (s *Postgres) BeginWebhookEvent(ctx context.Context, eventKey, payload string) (bool, error) {
	result, err := s.db.ExecContext(ctx, "INSERT INTO webhook_events(provider_event_key,payload) VALUES($1,$2::jsonb) ON CONFLICT(provider_event_key) DO NOTHING", eventKey, payload)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}
func (s *Postgres) CompleteWebhookEvent(ctx context.Context, eventKey, result string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE webhook_events SET processed_at=now(),result=$2 WHERE provider_event_key=$1", eventKey, result)
	return err
}
