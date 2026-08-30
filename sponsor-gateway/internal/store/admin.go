package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/OwO-Network/DLX/sponsor-gateway/internal/api"
)

func maskAdminEmail(email sql.NullString) *string {
	if !email.Valid || email.String == "" {
		return nil
	}
	at := strings.LastIndexByte(email.String, '@')
	if at <= 0 || at == len(email.String)-1 {
		return nil
	}
	local := email.String[:at]
	if len([]rune(local)) == 1 {
		local = "*"
	} else {
		local = string([]rune(local)[:1]) + "***"
	}
	masked := local + email.String[at:]
	return &masked
}

func (s *Postgres) AdminOverview(ctx context.Context) (api.AdminOverview, error) {
	var result api.AdminOverview
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE status='active'), COUNT(*) FILTER (WHERE status='suspended'), COUNT(*) FILTER (WHERE status='blocked') FROM users`).Scan(&result.Users.Total, &result.Users.Active, &result.Users.Suspended, &result.Users.Blocked); err != nil {
		return result, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FILTER (WHERE status='granted'), COUNT(*) FILTER (WHERE status='pending'), COUNT(*) FILTER (WHERE status='suspended'), COUNT(*) FILTER (WHERE status='manual_review') FROM entitlements`).Scan(&result.Entitlements.Granted, &result.Entitlements.Pending, &result.Entitlements.Suspended, &result.Entitlements.ManualReview); err != nil {
		return result, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FILTER (WHERE status IN ('paid','success')), COALESCE(SUM(actual_paid_fen) FILTER (WHERE status IN ('paid','success')),0), COUNT(*) FILTER (WHERE status='refunded') FROM afdian_orders`).Scan(&result.Orders.PaidCount, &result.Orders.PaidAmountFen, &result.Orders.RefundedCount); err != nil {
		return result, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(request_count),0), COALESCE(SUM(input_chars),0), COALESCE(SUM(error_count),0) FROM usage_daily WHERE date=CURRENT_DATE`).Scan(&result.Usage.TodayRequestCount, &result.Usage.TodayInputChars, &result.Usage.TodayErrorCount); err != nil {
		return result, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FILTER (WHERE status='active'), COUNT(*) FILTER (WHERE status='redeemed'), COALESCE(SUM(amount_fen) FILTER (WHERE status='redeemed'),0) FROM cdks`).Scan(&result.CDKs.ActiveCount, &result.CDKs.RedeemedCount, &result.CDKs.RedeemedAmountFen); err != nil {
		return result, err
	}
	result.GeneratedAt = time.Now().UTC()
	return result, nil
}

func (s *Postgres) ListAdminUsers(ctx context.Context, q api.AdminUserQuery) (api.AdminUserPage, error) {
	result := api.AdminUserPage{Page: q.Page, PageSize: q.PageSize, Items: make([]api.AdminUser, 0)}
	where := []string{"1=1"}
	args := make([]any, 0, 3)
	if q.Search != "" {
		args = append(args, "%"+strings.ToLower(q.Search)+"%")
		where = append(where, "LOWER(COALESCE(u.email,'')) LIKE $"+fmt.Sprint(len(args)))
	}
	if q.Status != "" {
		args = append(args, q.Status)
		where = append(where, "u.status=$"+fmt.Sprint(len(args)))
	}
	if q.EntitlementStatus != "" {
		args = append(args, q.EntitlementStatus)
		where = append(where, "e.status=$"+fmt.Sprint(len(args)))
	}
	filter := strings.Join(where, " AND ")
	countQuery := `SELECT COUNT(*) FROM users u LEFT JOIN entitlements e ON e.user_id=u.id WHERE ` + filter
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&result.Total); err != nil {
		return result, err
	}
	dataArgs := append([]any{}, args...)
	limitPos := len(dataArgs) + 1
	offsetPos := limitPos + 1
	dataArgs = append(dataArgs, q.PageSize, (q.Page-1)*q.PageSize)
	query := `SELECT u.id::text,u.email,u.status,u.created_at,COALESCE(e.status,'pending'),COALESCE(e.lifetime_paid_fen,0),e.granted_at,k.status,k.created_at,k.last_used_at FROM users u LEFT JOIN entitlements e ON e.user_id=u.id LEFT JOIN LATERAL (SELECT status,created_at,last_used_at FROM api_keys WHERE user_id=u.id AND status='active' ORDER BY created_at DESC LIMIT 1) k ON true WHERE ` + filter + ` ORDER BY u.created_at DESC,u.id DESC LIMIT $` + fmt.Sprint(limitPos) + ` OFFSET $` + fmt.Sprint(offsetPos)
	rows, err := s.db.QueryContext(ctx, query, dataArgs...)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		user, err := scanAdminUser(rows)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, user)
	}
	return result, rows.Err()
}

type rowScanner interface{ Scan(...any) error }

func scanAdminUser(row rowScanner) (api.AdminUser, error) {
	var user api.AdminUser
	var email sql.NullString
	var entitlementStatus string
	var grantedAt, keyCreatedAt, keyLastUsedAt sql.NullTime
	var keyStatus sql.NullString
	if err := row.Scan(&user.ID, &email, &user.Status, &user.CreatedAt, &entitlementStatus, &user.LifetimePaidFen, &grantedAt, &keyStatus, &keyCreatedAt, &keyLastUsedAt); err != nil {
		return user, err
	}
	user.Email = maskAdminEmail(email)
	user.EntitlementStatus = entitlementStatus
	if grantedAt.Valid {
		user.GrantedAt = &grantedAt.Time
	}
	if keyStatus.Valid {
		user.ActiveAPIKey = &api.AdminAPIKey{Status: keyStatus.String, CreatedAt: keyCreatedAt.Time}
		if keyLastUsedAt.Valid {
			user.ActiveAPIKey.LastUsedAt = &keyLastUsedAt.Time
		}
	}
	return user, nil
}

func (s *Postgres) GetAdminUser(ctx context.Context, userID string) (api.AdminUserDetail, error) {
	row := s.db.QueryRowContext(ctx, `SELECT u.id::text,u.email,u.status,u.created_at,COALESCE(e.status,'pending'),COALESCE(e.lifetime_paid_fen,0),e.granted_at,k.status,k.created_at,k.last_used_at,e.recalculated_at FROM users u LEFT JOIN entitlements e ON e.user_id=u.id LEFT JOIN LATERAL (SELECT status,created_at,last_used_at FROM api_keys WHERE user_id=u.id AND status='active' ORDER BY created_at DESC LIMIT 1) k ON true WHERE u.id=$1::uuid`, userID)
	var detail api.AdminUserDetail
	var email sql.NullString
	var entitlementStatus string
	var grantedAt, keyCreatedAt, keyLastUsedAt, recalculatedAt sql.NullTime
	var keyStatus sql.NullString
	if err := row.Scan(&detail.ID, &email, &detail.Status, &detail.CreatedAt, &entitlementStatus, &detail.LifetimePaidFen, &grantedAt, &keyStatus, &keyCreatedAt, &keyLastUsedAt, &recalculatedAt); err == sql.ErrNoRows {
		return detail, api.ErrAdminUserNotFound
	} else if err != nil {
		return detail, err
	}
	detail.Email = maskAdminEmail(email)
	detail.EntitlementStatus = entitlementStatus
	if grantedAt.Valid {
		detail.GrantedAt = &grantedAt.Time
	}
	if recalculatedAt.Valid {
		detail.RecalculatedAt = &recalculatedAt.Time
	}
	if keyStatus.Valid {
		detail.ActiveAPIKey = &api.AdminAPIKey{Status: keyStatus.String, CreatedAt: keyCreatedAt.Time}
		if keyLastUsedAt.Valid {
			detail.ActiveAPIKey.LastUsedAt = &keyLastUsedAt.Time
		}
	}
	usageSummary, err := s.AdminUserUsage(ctx, userID, api.UsageQuery{From: time.Now().UTC().AddDate(0, 0, -29).Format("2006-01-02"), To: time.Now().UTC().Format("2006-01-02")})
	if err != nil {
		return detail, err
	}
	detail.UsageSummary = usageSummary
	return detail, nil
}

func (s *Postgres) AdminUserUsage(ctx context.Context, userID string, q api.UsageQuery) (api.UsageSummary, error) {
	var exists bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1::uuid)`, userID).Scan(&exists); err != nil {
		return api.UsageSummary{}, err
	}
	if !exists {
		return api.UsageSummary{}, api.ErrAdminUserNotFound
	}
	from, _ := time.Parse("2006-01-02", q.From)
	to, _ := time.Parse("2006-01-02", q.To)
	rows, err := s.db.QueryContext(ctx, `SELECT d::date,COALESCE(u.request_count,0),COALESCE(u.input_chars,0),COALESCE(u.error_count,0) FROM generate_series($2::date,$3::date,interval '1 day') d LEFT JOIN usage_daily u ON u.user_id=$1::uuid AND u.date=d::date ORDER BY d`, userID, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return api.UsageSummary{}, err
	}
	defer rows.Close()
	result := api.UsageSummary{Days: make([]api.UsageDay, 0)}
	for rows.Next() {
		var day api.UsageDay
		var date time.Time
		if err := rows.Scan(&date, &day.RequestCount, &day.InputChars, &day.ErrorCount); err != nil {
			return api.UsageSummary{}, err
		}
		day.Date = date.Format("2006-01-02")
		result.Days = append(result.Days, day)
		result.TotalRequestCount += day.RequestCount
		result.TotalInputChars += day.InputChars
		result.TotalErrorCount += day.ErrorCount
	}
	return result, rows.Err()
}

func (s *Postgres) ListAdminOrders(ctx context.Context, userID string, q api.AdminOrderQuery) (api.AdminOrderPage, error) {
	result := api.AdminOrderPage{Page: q.Page, PageSize: q.PageSize, Items: make([]api.AdminOrder, 0)}
	where := []string{"1=1"}
	args := make([]any, 0, 4)
	if userID != "" {
		args = append(args, userID)
		where = append(where, "u.id=$"+fmt.Sprint(len(args))+"::uuid")
	}
	if q.Status != "" {
		args = append(args, q.Status)
		where = append(where, "o.status=$"+fmt.Sprint(len(args)))
	}
	args = append(args, q.From)
	where = append(where, "o.synced_at >= $"+fmt.Sprint(len(args))+"::date")
	args = append(args, q.To)
	where = append(where, "o.synced_at < ("+"$"+fmt.Sprint(len(args))+"::date + interval '1 day')")
	filter := strings.Join(where, " AND ")
	countArgs := append([]any{}, args...)
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM afdian_orders o JOIN afdian_identities i ON i.afdian_user_id=o.afdian_user_id JOIN users u ON u.id=i.user_id WHERE `+filter, countArgs...).Scan(&result.Total); err != nil {
		return result, err
	}
	limitPos := len(args) + 1
	offsetPos := limitPos + 1
	args = append(args, q.PageSize, (q.Page-1)*q.PageSize)
	rows, err := s.db.QueryContext(ctx, `SELECT u.id::text,u.email,o.actual_paid_fen,o.status,o.synced_at FROM afdian_orders o JOIN afdian_identities i ON i.afdian_user_id=o.afdian_user_id JOIN users u ON u.id=i.user_id WHERE `+filter+` ORDER BY o.synced_at DESC,o.out_trade_no DESC LIMIT $`+fmt.Sprint(limitPos)+` OFFSET $`+fmt.Sprint(offsetPos), args...)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var order api.AdminOrder
		var email sql.NullString
		if err := rows.Scan(&order.UserID, &email, &order.AmountFen, &order.Status, &order.SyncedAt); err != nil {
			return result, err
		}
		order.UserEmail = maskAdminEmail(email)
		result.Items = append(result.Items, order)
	}
	return result, rows.Err()
}

func (s *Postgres) ListAdminAPIKeys(ctx context.Context, q api.AdminAPIKeyQuery) (api.AdminAPIKeyPage, error) {
	result := api.AdminAPIKeyPage{Page: q.Page, PageSize: q.PageSize, Items: make([]api.AdminAPIKeyRecord, 0)}
	where := []string{"1=1"}
	args := make([]any, 0, 3)
	if q.Status != "" {
		args = append(args, q.Status)
		where = append(where, "k.status=$"+fmt.Sprint(len(args)))
	}
	if q.Q != "" {
		args = append(args, "%"+strings.ToLower(q.Q)+"%")
		where = append(where, "LOWER(u.email) LIKE $"+fmt.Sprint(len(args)))
	}
	filter := strings.Join(where, " AND ")
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM api_keys k JOIN users u ON u.id=k.user_id WHERE `+filter, args...).Scan(&result.Total); err != nil {
		return result, err
	}
	dataArgs := append([]any{}, args...)
	limitPos := len(dataArgs) + 1
	offsetPos := limitPos + 1
	dataArgs = append(dataArgs, q.PageSize, (q.Page-1)*q.PageSize)
	rows, err := s.db.QueryContext(ctx, `SELECT k.id::text,u.id::text,u.email,k.status,k.created_at,k.last_used_at FROM api_keys k JOIN users u ON u.id=k.user_id WHERE `+filter+` ORDER BY k.created_at DESC,k.id DESC LIMIT $`+fmt.Sprint(limitPos)+` OFFSET $`+fmt.Sprint(offsetPos), dataArgs...)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var record api.AdminAPIKeyRecord
		var email sql.NullString
		var lastUsedAt sql.NullTime
		if err := rows.Scan(&record.ID, &record.UserID, &email, &record.Status, &record.CreatedAt, &lastUsedAt); err != nil {
			return result, err
		}
		record.UserEmail = maskAdminEmail(email)
		if lastUsedAt.Valid {
			record.LastUsedAt = &lastUsedAt.Time
		}
		result.Items = append(result.Items, record)
	}
	return result, rows.Err()
}

var _ api.AdminStore = (*Postgres)(nil)
