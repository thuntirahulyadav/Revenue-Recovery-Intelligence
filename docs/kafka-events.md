# Razorpay Recovery Intelligence (RRI) — Kafka Event Lifecycle

RRI implements a real-time event-driven event architecture. All messages are wrapped in a standard JSON event envelope.

---

## Event Lifecycle

```text
payment.failed
       ↓
payment.enriched
       ↓
recovery.predicted
       ↓
recovery.decision.created
       ↓
recovery.policy.approved
       ↓
recovery.action.executed
       ↓
recovery.outcome.recorded
```

---

## Event Envelope Schema

```json
{
  "event_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
  "event_type": "payment.failed",
  "event_version": "1.0",
  "timestamp": "2026-08-29T14:30:00Z",
  "source": "rri.backend",
  "correlation_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "payload": {}
}
```

---

## Topics Specification

1. **`payment.failed`**: Emitted upon receiving a failed transaction from checkout / webhooks.
2. **`payment.enriched`**: Emitted after attaching historical customer lifetime value and reliability scores.
3. **`recovery.predicted`**: Emitted when the XGBoost ML model computes recovery probability, confidence, and SHAP factors.
4. **`recovery.decision.created`**: Emitted after economic net recovery optimization selects the highest net value strategy.
5. **`recovery.policy.approved`**: Emitted when policy engine rules or human operators grant execution clearance.
6. **`recovery.action.executed`**: Emitted when the recovery action (retry, payment link, channel switch) is dispatched.
7. **`recovery.outcome.recorded`**: Emitted when the stochastic or live outcome is measured, logging net recovered revenue and incurred cost.
