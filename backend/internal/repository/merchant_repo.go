package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"razorpay-recovery-intelligence/backend/internal/domain"

	"github.com/google/uuid"
)

func DefaultMerchantRecoverySettings() domain.MerchantRecoverySettings {
	return domain.MerchantRecoverySettings{
		MaxRetryAttempts:              3,
		MinConfidenceThreshold:        0.65,
		MaxCommAttempts:               2,
		HumanApprovalThreshold:        50000,
		HighValueTransactionThreshold: 25000,
		AutoExecutionEnabled:          true,
	}
}

type MerchantRepo struct{ db *DB }

func NewMerchantRepo(db *DB) *MerchantRepo { return &MerchantRepo{db: db} }

func (r *MerchantRepo) GetRecoverySettings(ctx context.Context, merchantID uuid.UUID) (domain.MerchantRecoverySettings, error) {
	settings := DefaultMerchantRecoverySettings()
	var raw []byte
	err := r.db.Pool.QueryRow(ctx, `SELECT recovery_settings FROM merchants WHERE id = $1`, merchantID).Scan(&raw)
	if err != nil {
		return settings, fmt.Errorf("load merchant recovery settings: %w", err)
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return DefaultMerchantRecoverySettings(), fmt.Errorf("decode merchant recovery settings: %w", err)
	}
	return settings, nil
}
