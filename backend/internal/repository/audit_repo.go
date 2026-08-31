package repository

import (
	"context"
	"encoding/json"

	"razorpay-recovery-intelligence/backend/internal/domain"

	"github.com/google/uuid"
)

type AuditRepo struct {
	db *DB
}

func NewAuditRepo(db *DB) *AuditRepo {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) LogEvent(ctx context.Context, log domain.AuditLog) error {
	metaJSON, err := json.Marshal(log.Metadata)
	if err != nil {
		metaJSON = []byte("{}")
	}

	query := `
		INSERT INTO audit_logs (id, payment_id, event_type, actor, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = r.db.Pool.Exec(ctx, query,
		log.ID, log.PaymentID, log.EventType, log.Actor, metaJSON, log.CreatedAt,
	)
	return err
}

func (r *AuditRepo) GetLogsByPaymentID(ctx context.Context, paymentID uuid.UUID) ([]domain.AuditLog, error) {
	query := `
		SELECT id, payment_id, event_type, actor, metadata, created_at
		FROM audit_logs
		WHERE payment_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]domain.AuditLog, 0)
	for rows.Next() {
		var l domain.AuditLog
		var metaJSON []byte
		if err := rows.Scan(&l.ID, &l.PaymentID, &l.EventType, &l.Actor, &metaJSON, &l.CreatedAt); err == nil {
			if len(metaJSON) > 0 {
				_ = json.Unmarshal(metaJSON, &l.Metadata)
			}
			logs = append(logs, l)
		}
	}
	return logs, nil
}
