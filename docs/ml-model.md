# Razorpay Recovery Intelligence (RRI) — ML Model & Explainability

---

## 1. Machine Learning Strategy

Instead of treating payment recovery as an unstructured generative prompt (which is unpredictable and dangerous for financial workflows), RRI leverages **XGBoost (Extreme Gradient Boosting)** combined with **SHAP (SHapley Additive exPlanations)**.

- **Primary Objective**: Predict the continuous probability of successful recovery $\hat{P} \in [0.0, 1.0]$.
- **Economic Formulation**:
$$\text{Expected Net Value} = (\text{Transaction Amount} \times \hat{P}_{\text{strategy}}) - \text{Strategy Cost}$$

---

## 2. Feature Architecture

| Feature Category | Features | Description |
|---|---|---|
| **Transaction Profile** | `transaction_amount`, `value_per_attempt` | Value at stake and value normalized per attempt. |
| **Payment Instrument** | `pm_card`, `pm_upi`, `pm_netbanking`, `pm_wallet`, `pm_emi` | One-hot encoded payment rails. |
| **Failure Category** | `reason_BANK_TIMEOUT`, `reason_NETWORK_ERROR`, `reason_INSUFFICIENT_FUNDS`, `reason_CARD_EXPIRED`, `reason_PAYMENT_METHOD_FAILURE`, `reason_CUSTOMER_ABANDONMENT`, `reason_TECHNICAL_ERROR` | Core root cause classifications. |
| **Historical Trust** | `customer_success_rate`, `customer_failure_rate`, `customer_risk_ratio`, `customer_value` | Customer lifetime financial reliability. |
| **Attempt State** | `attempt_count` | Number of previous retries. |
| **Temporal Context** | `hour_of_day`, `day_of_week`, `is_peak_hour`, `is_weekend` | Processing window dynamics. |

---

## 3. Model Benchmark Comparison

Trained on 12,500 deterministic synthetic payment events:

| Metric | Baseline (Logistic Regression) | Primary (XGBoost Classifier) |
|---|---|---|
| **Accuracy** | 63.40% | 63.24% |
| **Precision** | 64.50% | 64.22% |
| **Recall** | 81.99% | **82.61%** |
| **F1 Score** | 72.20% | **72.26%** |
| **ROC-AUC** | 64.09% | 63.91% |
| **PR-AUC** | 68.27% | 67.63% |

---

## 4. SHAP Feature Attribution

Top feature contributors:
1. `attempt_count` (Heavy negative drag after attempt 2)
2. `reason_INSUFFICIENT_FUNDS` (Requires strategic delayed retries)
3. `customer_risk_ratio`
4. `reason_NETWORK_ERROR` (High positive recovery potential)
5. `reason_CUSTOMER_ABANDONMENT` (High lift with payment link outreach)
6. `customer_success_rate`
