-- ==========================================================
-- Razorpay Recovery Intelligence (RRI) - PostgreSQL Schema
-- ==========================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. MERCHANTS
CREATE TABLE IF NOT EXISTS merchants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    recovery_settings JSONB DEFAULT '{
        "max_retry_attempts": 3,
        "min_confidence_threshold": 0.65,
        "max_comm_attempts": 2,
        "human_approval_threshold": 50000,
        "high_value_transaction_threshold": 25000,
        "auto_execution_enabled": true
    }'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. CUSTOMERS
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    merchant_id UUID REFERENCES merchants(id) ON DELETE CASCADE,
    email VARCHAR(255),
    phone VARCHAR(50),
    historical_success_rate DECIMAL(5, 4) DEFAULT 0.8500,
    historical_failure_rate DECIMAL(5, 4) DEFAULT 0.1500,
    customer_value DECIMAL(12, 2) DEFAULT 10000.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. PAYMENTS
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    merchant_id UUID REFERENCES merchants(id) ON DELETE CASCADE,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    amount BIGINT NOT NULL, -- In paise or currency subunits (e.g., 25000 = Rs 250.00 or Rs 25,000 depending on denomination)
    currency VARCHAR(10) DEFAULT 'INR',
    payment_method VARCHAR(50) NOT NULL, -- card, upi, netbanking, wallet, emi
    status VARCHAR(50) NOT NULL DEFAULT 'FAILED', -- FAILED, RECOVERED, RECOVERING, ABANDONED
    failure_reason VARCHAR(100) NOT NULL, -- BANK_TIMEOUT, NETWORK_ERROR, INSUFFICIENT_FUNDS, CARD_EXPIRED, PAYMENT_METHOD_FAILURE, CUSTOMER_ABANDONMENT, TECHNICAL_ERROR
    attempt_count INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. RECOVERY PREDICTIONS
CREATE TABLE IF NOT EXISTS recovery_predictions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_id UUID REFERENCES payments(id) ON DELETE CASCADE,
    recovery_probability DECIMAL(5, 4) NOT NULL,
    model_version VARCHAR(100) NOT NULL,
    confidence DECIMAL(5, 4) NOT NULL,
    shap_factors JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. RECOVERY DECISIONS
CREATE TABLE IF NOT EXISTS recovery_decisions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_id UUID REFERENCES payments(id) ON DELETE CASCADE,
    strategy VARCHAR(100) NOT NULL, -- RETRY_NOW, RETRY_LATER, SWITCH_PAYMENT_METHOD, SEND_PAYMENT_LINK, ESCALATE_TO_HUMAN, STOP_RECOVERY
    expected_revenue DECIMAL(12, 2) NOT NULL,
    expected_cost DECIMAL(12, 2) NOT NULL,
    expected_net_value DECIMAL(12, 2) NOT NULL,
    priority_score DECIMAL(5, 2) DEFAULT 0.0,
    explanation TEXT,
    policy_status VARCHAR(50) DEFAULT 'APPROVED', -- APPROVED, REJECTED, PENDING_HUMAN_APPROVAL
    policy_checks JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 6. RECOVERY ACTIONS
CREATE TABLE IF NOT EXISTS recovery_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_id UUID REFERENCES payments(id) ON DELETE CASCADE,
    action_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- PENDING, EXECUTED, SIMULATED, FAILED
    execution_mode VARCHAR(50) DEFAULT 'SIMULATED', -- SIMULATED, MOCK, REAL
    payload JSONB DEFAULT '{}'::jsonb,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 7. RECOVERY OUTCOMES
CREATE TABLE IF NOT EXISTS recovery_outcomes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    action_id UUID REFERENCES recovery_actions(id) ON DELETE CASCADE,
    payment_id UUID REFERENCES payments(id) ON DELETE CASCADE,
    success BOOLEAN NOT NULL,
    recovered_amount DECIMAL(12, 2) NOT NULL DEFAULT 0.0,
    recovery_cost DECIMAL(12, 2) NOT NULL DEFAULT 0.0,
    net_recovery_value DECIMAL(12, 2) NOT NULL DEFAULT 0.0,
    completed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 8. AUDIT LOGS
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_id UUID,
    event_type VARCHAR(100) NOT NULL,
    actor VARCHAR(100) NOT NULL DEFAULT 'rri.system',
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ==========================================================
-- INDEXES FOR PERFORMANCE
-- ==========================================================
CREATE INDEX IF NOT EXISTS idx_customers_merchant_id ON customers(merchant_id);
CREATE INDEX IF NOT EXISTS idx_payments_merchant_id ON payments(merchant_id);
CREATE INDEX IF NOT EXISTS idx_payments_customer_id ON payments(customer_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_failure_reason ON payments(failure_reason);
CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_recovery_predictions_payment_id ON recovery_predictions(payment_id);
CREATE INDEX IF NOT EXISTS idx_recovery_decisions_payment_id ON recovery_decisions(payment_id);
CREATE INDEX IF NOT EXISTS idx_recovery_decisions_strategy ON recovery_decisions(strategy);
CREATE INDEX IF NOT EXISTS idx_recovery_actions_payment_id ON recovery_actions(payment_id);
CREATE INDEX IF NOT EXISTS idx_recovery_outcomes_payment_id ON recovery_outcomes(payment_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_payment_id ON audit_logs(payment_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs(event_type);
