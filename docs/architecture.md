# Razorpay Recovery Intelligence (RRI) — Architecture Specification

> **Track**: Razorpay Buildathon — AI Revenue Recovery Track  
> **Tagline**: *Don't retry every failure. Recover the revenue worth recovering.*

---

## 1. System Philosophy & Executive Summary

Traditional payment systems handle failed transactions with blind, uniform retries (e.g. 3 attempts over 24 hours). This causes three severe problems:
1. **Wasted Transaction Overhead**: Retrying unrecoverable errors (e.g. `CARD_EXPIRED`, persistent zero balance) burns gateway fees and exhausts issuer rate limits.
2. **Customer Churn**: Repetitive failed auto-debit SMS alerts erode merchant trust.
3. **Suboptimal Recovery Channels**: Dropped 3DS OTP sessions (`CUSTOMER_ABANDONMENT`) need instant WhatsApp/SMS Payment Links, not automated backend retries.

**Razorpay Recovery Intelligence (RRI)** provides a **closed-loop revenue recovery orchestration engine** combining machine learning (XGBoost + SHAP), real-time enrichment, economic net value optimization, and merchant safety policy validation.

---

## 2. Real Closed-Loop Workflow

```text
[DETECT] Failed Payment Event (Webhook / Kafka `payment.failed`)
   │
   ▼
[ENRICH] Customer History + Merchant Context + Transaction Profile (`payment.enriched`)
   │
   ▼
[PREDICT] ML Probability & Confidence Estimation via XGBoost (`recovery.predicted`)
   │
   ▼
[PRIORITIZE] Value-at-Risk × Recovery Probability Matrix
   │
   ▼
[SELECT STRATEGY] Maximize Expected Net Value = (Amount × P_strategy) - Cost (`recovery.decision.created`)
   │
   ▼
[VALIDATE POLICY] Retry limits, confidence gates, merchant constraints (`recovery.policy.approved`)
   │
   ▼
[EXECUTE / SIMULATE] Safe execution or simulated recovery action (`recovery.action.executed`)
   │
   ▼
[MEASURE OUTCOME] Cost, revenue recovered, net ROI recorded (`recovery.outcome.recorded`)
```

---

## 3. High-Level Architectural Topology

```text
                         ┌─────────────────────────┐
                         │   React + TypeScript    │
                         │    Frontend UI (Vite)   │
                         └────────────┬────────────┘
                                      │ HTTP / REST
                                      ▼
                         ┌─────────────────────────┐
                         │     Go API Gateway      │
                         │    Backend (Gin / REST) │
                         └────────────┬────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
             ┌─────────────┐   ┌────────────┐   ┌─────────────┐
             │ PostgreSQL  │   │   Redis    │   │    Kafka    │
             │ Persistent  │   │ Caching &  │   │ Event-Driven│
             │ RDBMS Data  │   │ Idempotency│   │ Stream Bus  │
             └─────────────┘   └────────────┘   └──────┬──────┘
                                                        │
                                                        ▼
                                             ┌──────────────────┐
                                             │ Python ML Engine │
                                             │ XGBoost + SHAP   │
                                             └─────────┬────────┘
                                                       │
                                                       ▼
                                             ┌──────────────────┐
                                             │ Decision Engine  │
                                             │ + Policy Engine  │
                                             └──────────────────┘
```

---

## 4. Component Responsibilities

### 4.1 Frontend UI (React + TypeScript + Vite + Tailwind CSS + Recharts)
- **Recovery Command Center (`/dashboard`)**: KPI metrics (Revenue at Risk, Recovered, Incremental, Saved Costs), time-series area charts, root failure cause distributions.
- **Recovery Opportunities Queue (`/opportunities`)**: Real-time prioritized queue with multi-dimensional filtering, searching, sorting, and action triggers.
- **Payment Detail Deep Dive (`/payments/:id`)**: Comprehensive breakdown of failure, ML recovery probability gauge, SHAP positive/negative drivers, candidate strategy matrix, policy checklist, and action executor.
- **Simulation Lab (`/simulation`)**: High-fidelity benchmark demonstrating Baseline ("Blind Retry All") vs RRI ("AI Prioritized Strategy"), displaying incremental revenue, net profit lift, and ROI multiple.
- **Merchant Policy Controls (`/settings`)**: Configurable thresholds for retry limits, confidence gates, high-value transaction human approvals, and auto-execution toggles.

### 4.2 Go Backend API Gateway
- **Layered Clean Architecture**: `domain/` &rarr; `repository/` &rarr; `service/` &rarr; `handler/` &rarr; `middleware/`.
- **Economic Decision Engine**: Evaluates expected net recovery values across `RETRY_NOW`, `RETRY_LATER`, `SWITCH_PAYMENT_METHOD`, `SEND_PAYMENT_LINK`, `ESCALATE_TO_HUMAN`, `STOP_RECOVERY`.
- **Policy Engine**: Validates compliance with max retry limits, minimum ML confidence thresholds, and high-value authorization rules.
- **Event Bus**: Kafka integration with resilient in-memory dispatcher fallback for local zero-dependency testing.
- **Idempotency & Rate Limiting**: Redis-backed mutex keys preventing duplicate event ingestion or execution.
- **Audit Logger**: Structured compliance audit trail for every state change and decision.

### 4.3 Python ML Service (FastAPI + XGBoost + SHAP)
- **Feature Engineering Pipeline**: Transforms raw failure reason, payment method, customer history, attempt count, and timing into high-signal feature vectors.
- **XGBoost Classifier**: Primary non-linear tree model optimized for PR-AUC and Recall on recovery probability prediction.
- **Logistic Regression**: Baseline linear benchmark for continuous model evaluation.
- **SHAP TreeExplainer**: Real-time instance-level positive and negative feature attribution.
- **FastAPI Endpoints**: `/predict`, `/explain`, `/metrics`, `/health`.
