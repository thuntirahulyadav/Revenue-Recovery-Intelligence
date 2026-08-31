package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"razorpay-recovery-intelligence/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type OpportunitiesFilter struct {
	FailureReason  string
	Strategy       string
	MinProbability float64
	MinAmount      int64
	Status         string
	Search         string
	SortBy         string // e.g., priority_score, amount, recovery_probability, created_at
	SortOrder      string // ASC or DESC
	Page           int
	Limit          int
}

type OpportunityItem struct {
	PaymentID           uuid.UUID               `json:"payment_id"`
	MerchantID          uuid.UUID               `json:"merchant_id"`
	CustomerID          uuid.UUID               `json:"customer_id"`
	Amount              int64                   `json:"amount"`
	Currency            string                  `json:"currency"`
	PaymentMethod       domain.PaymentMethod    `json:"payment_method"`
	Status              domain.PaymentStatus    `json:"status"`
	FailureReason       domain.FailureReason    `json:"failure_reason"`
	AttemptCount        int                     `json:"attempt_count"`
	CreatedAt           time.Time               `json:"created_at"`
	RecoveryProbability float64                 `json:"recovery_probability"`
	Confidence          float64                 `json:"confidence"`
	Strategy            domain.RecoveryStrategy `json:"strategy"`
	ExpectedRevenue     float64                 `json:"expected_revenue"`
	ExpectedCost        float64                 `json:"expected_cost"`
	ExpectedNetValue    float64                 `json:"expected_net_value"`
	PriorityScore       float64                 `json:"priority_score"`
	PolicyStatus        domain.PolicyStatus     `json:"policy_status"`
	CustomerValue       float64                 `json:"customer_value"`
	CustomerSuccessRate float64                 `json:"customer_success_rate"`
}

type PaginatedOpportunities struct {
	Items      []OpportunityItem `json:"items"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

type PaymentRepo struct {
	db *DB
}

func NewPaymentRepo(db *DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) CreatePayment(ctx context.Context, p *domain.Payment) error {
	query := `
		INSERT INTO payments (id, merchant_id, customer_id, amount, currency, payment_method, status, failure_reason, attempt_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			attempt_count = EXCLUDED.attempt_count
	`
	_, err := r.db.Pool.Exec(ctx, query,
		p.ID, p.MerchantID, p.CustomerID, p.Amount, p.Currency, p.PaymentMethod, p.Status, p.FailureReason, p.AttemptCount, p.CreatedAt,
	)
	return err
}

func (r *PaymentRepo) GetPaymentByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	query := `
		SELECT id, merchant_id, customer_id, amount, currency, payment_method, status, failure_reason, attempt_count, created_at
		FROM payments
		WHERE id = $1
	`
	var p domain.Payment
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.MerchantID, &p.CustomerID, &p.Amount, &p.Currency, &p.PaymentMethod, &p.Status, &p.FailureReason, &p.AttemptCount, &p.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepo) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus, attemptCount int) error {
	query := `UPDATE payments SET status = $1, attempt_count = $2 WHERE id = $3`
	_, err := r.db.Pool.Exec(ctx, query, status, attemptCount, id)
	return err
}

func (r *PaymentRepo) GetRecoveryOpportunities(ctx context.Context, filter OpportunitiesFilter) (*PaginatedOpportunities, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 15
	}
	offset := (filter.Page - 1) * filter.Limit

	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if filter.FailureReason != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("p.failure_reason = $%d", argIdx))
		args = append(args, filter.FailureReason)
		argIdx++
	}

	if filter.Strategy != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("rd.strategy = $%d", argIdx))
		args = append(args, filter.Strategy)
		argIdx++
	}

	if filter.MinProbability > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("rp.recovery_probability >= $%d", argIdx))
		args = append(args, filter.MinProbability)
		argIdx++
	}

	if filter.MinAmount > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("p.amount >= $%d", argIdx))
		args = append(args, filter.MinAmount)
		argIdx++
	}

	if filter.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("p.status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(p.id::text ILIKE $%d OR c.email ILIKE $%d OR p.failure_reason ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Count Query
	countQuery := fmt.Sprintf(`
		SELECT COUNT(p.id)
		FROM payments p
		LEFT JOIN LATERAL (SELECT * FROM recovery_predictions WHERE payment_id = p.id ORDER BY created_at DESC LIMIT 1) rp ON true
		LEFT JOIN LATERAL (SELECT * FROM recovery_decisions WHERE payment_id = p.id ORDER BY created_at DESC LIMIT 1) rd ON true
		LEFT JOIN customers c ON p.customer_id = c.id
		WHERE %s
	`, whereSQL)

	var totalCount int
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("error counting opportunities: %w", err)
	}

	// Sorting
	sortField := "rd.priority_score"
	switch filter.SortBy {
	case "amount":
		sortField = "p.amount"
	case "recovery_probability":
		sortField = "rp.recovery_probability"
	case "expected_net_value":
		sortField = "rd.expected_net_value"
	case "created_at":
		sortField = "p.created_at"
	case "priority_score":
		sortField = "rd.priority_score"
	}

	sortDirection := "DESC"
	if strings.ToUpper(filter.SortOrder) == "ASC" {
		sortDirection = "ASC"
	}

	dataQuery := fmt.Sprintf(`
		SELECT 
			p.id, p.merchant_id, p.customer_id, p.amount, p.currency, p.payment_method, p.status, p.failure_reason, p.attempt_count, p.created_at,
			COALESCE(rp.recovery_probability, 0.0), COALESCE(rp.confidence, 0.0),
			COALESCE(rd.strategy, 'RETRY_LATER'), COALESCE(rd.expected_revenue, 0.0), COALESCE(rd.expected_cost, 0.0),
			COALESCE(rd.expected_net_value, 0.0), COALESCE(rd.priority_score, 0.0), COALESCE(rd.policy_status, 'APPROVED'),
			COALESCE(c.customer_value, 0.0), COALESCE(c.historical_success_rate, 0.85)
		FROM payments p
		LEFT JOIN LATERAL (SELECT * FROM recovery_predictions WHERE payment_id = p.id ORDER BY created_at DESC LIMIT 1) rp ON true
		LEFT JOIN LATERAL (SELECT * FROM recovery_decisions WHERE payment_id = p.id ORDER BY created_at DESC LIMIT 1) rd ON true
		LEFT JOIN customers c ON p.customer_id = c.id
		WHERE %s
		ORDER BY %s %s NULLS LAST
		LIMIT $%d OFFSET $%d
	`, whereSQL, sortField, sortDirection, argIdx, argIdx+1)

	args = append(args, filter.Limit, offset)

	rows, err := r.db.Pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying opportunities: %w", err)
	}
	defer rows.Close()

	items := make([]OpportunityItem, 0, filter.Limit)
	for rows.Next() {
		var item OpportunityItem
		err := rows.Scan(
			&item.PaymentID, &item.MerchantID, &item.CustomerID, &item.Amount, &item.Currency, &item.PaymentMethod, &item.Status, &item.FailureReason, &item.AttemptCount, &item.CreatedAt,
			&item.RecoveryProbability, &item.Confidence,
			&item.Strategy, &item.ExpectedRevenue, &item.ExpectedCost,
			&item.ExpectedNetValue, &item.PriorityScore, &item.PolicyStatus,
			&item.CustomerValue, &item.CustomerSuccessRate,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning opportunity item: %w", err)
		}
		items = append(items, item)
	}

	totalPages := (totalCount + filter.Limit - 1) / filter.Limit
	return &PaginatedOpportunities{
		Items:      items,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}
