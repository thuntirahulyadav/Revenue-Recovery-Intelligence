package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventPaymentFailed           EventType = "payment.failed"
	EventPaymentEnriched         EventType = "payment.enriched"
	EventRecoveryPredicted       EventType = "recovery.predicted"
	EventRecoveryDecisionCreated EventType = "recovery.decision.created"
	EventRecoveryPolicyApproved  EventType = "recovery.policy.approved"
	EventRecoveryActionExecuted  EventType = "recovery.action.executed"
	EventRecoveryOutcomeRecorded EventType = "recovery.outcome.recorded"
)

type EventEnvelope struct {
	EventID       uuid.UUID   `json:"event_id"`
	EventType     EventType   `json:"event_type"`
	EventVersion  string      `json:"event_version"`
	Timestamp     string      `json:"timestamp"` // ISO-8601
	Source        string      `json:"source"`
	CorrelationID uuid.UUID   `json:"correlation_id"`
	Payload       interface{} `json:"payload"`
}

func NewEventEnvelope(eventType EventType, correlationID uuid.UUID, payload interface{}) EventEnvelope {
	return EventEnvelope{
		EventID:       uuid.New(),
		EventType:     eventType,
		EventVersion:  "1.0",
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Source:        "rri.backend",
		CorrelationID: correlationID,
		Payload:       payload,
	}
}
