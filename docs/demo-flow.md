# Razorpay Recovery Intelligence (RRI) — Buildathon Demo Walkthrough

---

## 3-Minute Judge Evaluation Guide

### 🎯 Key Evaluation Question
> **"Can intelligent decisioning recover more net revenue than a generic recovery strategy?"**

---

### Step 1: Recovery Command Center (`/dashboard`)
1. Open the UI at `http://localhost:5173`.
2. Inspect the **Live KPI Cards**:
   - **Revenue At Risk**: Real-time aggregation of failed volume.
   - **Revenue Recovered**: Verified recoveries from AI orchestrated channels.
   - **Incremental Recovery**: The pure profit lift generated above standard blind retries.
   - **Wasted Costs Saved**: Unproductive gateway retry fees prevented by terminating non-viable retries.
3. Review the **Revenue Over Time** chart showing how AI recovery separates from the baseline trajectory.
4. Review the **Failure Distribution** chart highlighting distinct failure categories (`BANK_TIMEOUT`, `INSUFFICIENT_FUNDS`, `CUSTOMER_ABANDONMENT`, etc.).

---

### Step 2: Simulate Live Failure Event
1. Click the top-bar button: **`+ Simulate Failure`**.
2. Select parameters:
   - Amount: `₹15,000`
   - Payment Method: `UPI`
   - Failure Reason: `BANK_TIMEOUT`
   - Attempt Count: `1`
3. Click **Emit payment.failed Event**.
4. The system executes the full closed loop in sub-100ms:
   - Ingests & persists payment
   - Enriches customer history
   - Invokes XGBoost ML model
   - Computes Expected Net Value across 6 candidate strategies
   - Passes Policy Engine compliance checks
5. Click **View Full AI Breakdown**.

---

### Step 3: Deep Dive & Explainability (`/payments/:paymentId`)
1. View the **AI Recovery Scorecard**: Probability (e.g. 78%) and Model Confidence (88%).
2. Inspect the **Economic Formula Breakdown**:
   $$\text{Expected Net Value} = (\text{Amount} \times P) - \text{Strategy Cost} = (₹15,000 \times 0.78) - ₹8.00 = ₹11,692.00$$
3. Inspect **Why? SHAP Explanations**:
   - Positive drivers (+ Bank Timeout pattern, + Customer high success history).
   - Risk constraints (- Attempt penalty).
4. Review the **Policy Engine Validation Checklist**:
   - ✓ Max Retry Limit Passed
   - ✓ Confidence Threshold Passed
   - ✓ High Value Authorization
   - ✓ Economic Net Viability
5. Click **Execute Recovery Action**:
   - Displays clear `[SIMULATED RECOVERY ACTION]` banner adhering to hackathon rules.
   - Triggers simulated execution and records the measured recovery outcome.

---

### Step 4: Economic Simulation Lab (`/simulation`)
1. Navigate to **Simulation Lab**.
2. Compare **Baseline Strategy ("Retry Everything 3 Times")** vs **Recovery Intelligence Strategy**:
   - **Baseline**: High action costs, thousands of wasted retry fees on dead ends.
   - **Recovery Intelligence**: Selects tailored channels, eliminates useless retries, lifts recovery rate by **+36%**, and yields a **2.4x Net ROI Uplift**.
3. Change the cohort sample size (e.g. 5,000 or 10,000 payments) and click **Re-run Simulation** to test dynamic cohorts in real time.

---

### Step 5: Merchant Policy Settings (`/settings`)
1. Adjust merchant recovery parameters:
   - Max Retry Attempts
   - Minimum Confidence Threshold (50% - 90%)
   - Human Approval Threshold for VIP transactions
2. Click **Save Policy Changes** to update the runtime policy engine rules.
