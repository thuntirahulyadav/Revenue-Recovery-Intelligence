package repository

import (
	"context"
	"time"

	"razorpay-recovery-intelligence/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CustomerRepo struct {
	db *DB
}

func NewCustomerRepo(db *DB) *CustomerRepo {
	return &CustomerRepo{db: db}
}

func (r *CustomerRepo) GetOrCreateCustomer(ctx context.Context, merchantID, customerID uuid.UUID) (*domain.Customer, error) {
	query := `
		SELECT id, merchant_id, COALESCE(email, ''), COALESCE(phone, ''), historical_success_rate, historical_failure_rate, customer_value, created_at
		FROM customers
		WHERE id = $1
	`
	var c domain.Customer
	err := r.db.Pool.QueryRow(ctx, query, customerID).Scan(
		&c.ID, &c.MerchantID, &c.Email, &c.Phone, &c.HistoricalSuccessRate, &c.HistoricalFailureRate, &c.CustomerValue, &c.CreatedAt,
	)
	if err == nil {
		return &c, nil
	}
	if err != pgx.ErrNoRows {
		return nil, err
	}

	// Insert default customer
	insertQuery := `
		INSERT INTO customers (id, merchant_id, email, phone, historical_success_rate, historical_failure_rate, customer_value, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO NOTHING
		RETURNING id, merchant_id, COALESCE(email, ''), COALESCE(phone, ''), historical_success_rate, historical_failure_rate, customer_value, created_at
	`
	defaultEmail := "customer_" + customerID.String()[:8] + "@example.com"
	defaultPhone := "+9198" + customerID.String()[:8]
	now := time.Now().UTC()

	err = r.db.Pool.QueryRow(ctx, insertQuery,
		customerID, merchantID, defaultEmail, defaultPhone, 0.8500, 0.1500, 15000.00, now,
	).Scan(
		&c.ID, &c.MerchantID, &c.Email, &c.Phone, &c.HistoricalSuccessRate, &c.HistoricalFailureRate, &c.CustomerValue, &c.CreatedAt,
	)
	if err != nil {
		// Fallback query if conflict occurred concurrently
		_ = r.db.Pool.QueryRow(ctx, query, customerID).Scan(
			&c.ID, &c.MerchantID, &c.Email, &c.Phone, &c.HistoricalSuccessRate, &c.HistoricalFailureRate, &c.CustomerValue, &c.CreatedAt,
		)
		return &c, nil
	}
	return &c, nil
}

func (r *CustomerRepo) GetCustomerByID(ctx context.Context, customerID uuid.UUID) (*domain.Customer, error) {
	query := `
		SELECT id, merchant_id, COALESCE(email, ''), COALESCE(phone, ''), historical_success_rate, historical_failure_rate, customer_value, created_at
		FROM customers
		WHERE id = $1
	`
	var c domain.Customer
	err := r.db.Pool.QueryRow(ctx, query, customerID).Scan(
		&c.ID, &c.MerchantID, &c.Email, &c.Phone, &c.HistoricalSuccessRate, &c.HistoricalFailureRate, &c.CustomerValue, &c.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}
