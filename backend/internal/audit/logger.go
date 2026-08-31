package audit

import (
	"context"
	"log"
	"time"

	"razorpay-recovery-intelligence/backend/internal/domain"
	"razorpay-recovery-intelligence/backend/internal/repository"

	"github.com/google/uuid"
)

type Logger struct {
	repo *repository.AuditRepo
}

func NewLogger(repo *repository.AuditRepo) *Logger {
	return &Logger{repo: repo}
}

func (l *Logger) Log(ctx context.Context, paymentID *uuid.UUID, eventType, actor string, metadata map[string]interface{}) {
	entry := domain.AuditLog{
		ID:        uuid.New(),
		PaymentID: paymentID,
		EventType: eventType,
		Actor:     actor,
		Metadata:  metadata,
		CreatedAt: time.Now().UTC(),
	}

	log.Printf("[AuditLog] [%s] actor=%s payment_id=%v", eventType, actor, paymentID)

	go func() {
		// Log asynchronously to ensure zero latency overhead on main pipeline
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := l.repo.LogEvent(bgCtx, entry); err != nil {
			log.Printf("[AuditLog] Failed to persist audit log: %v", err)
		}
	}()
}
