# Razorpay Recovery Intelligence (RRI) — API Contracts

Base URL: `http://localhost:8080/api/v1`

---

## 1. Event Ingestion

### `POST /events/payment-failed`
Ingests a failed payment event from Razorpay webhooks or internal payment systems and immediately triggers the closed-loop recovery orchestration pipeline.

#### Request Body
```json
{
  "payment_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "merchant_id": "00000000-0000-0000-0000-000000000001",
  "customer_id": "3fa85f64-5717-4562-b3fc-2c963f66afa7",
  "amount": 25000,
  "currency": "INR",
  "payment_method": "card",
  "failure_reason": "BANK_TIMEOUT",
  "attempt_count": 1
}
```

#### Response (`201 Created`)
```json
{
  "success": true,
  "data": {
    "payment": {
      "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
      "amount": 25000,
      "status": "FAILED",
      "failure_reason": "BANK_TIMEOUT",
      "attempt_count": 1
    },
    "customer": {
      "historical_success_rate": 0.92,
      "customer_value": 50000
    },
    "prediction": {
      "recovery_probability": 0.82,
      "confidence": 0.88,
      "model_version": "v1.0.0",
      "shap_factors": [
        {
          "feature": "reason_BANK_TIMEOUT",
          "impact": 0.24,
          "direction": "positive",
          "description": "Transient core bank timeout"
        }
      ]
    },
    "decision": {
      "strategy": "RETRY_LATER",
      "expected_revenue": 20500.0,
      "expected_cost": 8.0,
      "expected_net_value": 20492.0,
      "priority_score": 38.5,
      "policy_status": "APPROVED",
      "explanation": "Delayed retry recommended for BANK_TIMEOUT to allow issuer/balance resolution."
    }
  },
  "meta": {
    "correlation_id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "timestamp": "2026-08-29T14:30:00Z"
  }
}
```

---

## 2. Recovery Analysis

### `GET /payments/:payment_id/recovery`
Returns the complete recovery breakdown for a payment, including payment details, ML recovery probability, SHAP factors, economic candidate matrix, policy checks, action status, and outcomes.

---

## 3. Opportunities Queue

### `GET /recovery/opportunities`
Query parameters:
- `page`: int (default `1`)
- `limit`: int (default `15`)
- `failure_reason`: string
- `strategy`: string
- `min_probability`: float (0.0 to 1.0)
- `min_amount`: int
- `sort_by`: string (`priority_score`, `amount`, `recovery_probability`, `expected_net_value`)
- `sort_order`: string (`DESC`, `ASC`)
- `search`: string

---

## 4. Policy Approval & Action Execution

### `POST /recovery/:payment_id/approve`
Manually clears human approval gates for high-value transactions.

### `POST /recovery/:payment_id/execute`
Executes or simulates the chosen recovery strategy action.
#### Request Body
```json
{
  "execution_mode": "SIMULATED" // "SIMULATED" | "MOCK" | "REAL"
}
```

---

## 5. Simulation Comparison

### `GET /simulation/compare?sample_size=2500`
Returns side-by-side performance analytics of Baseline (blind retry) vs AI Recovery Intelligence.
