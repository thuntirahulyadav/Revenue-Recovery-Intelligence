export type PaymentMethod = 'card' | 'upi' | 'netbanking' | 'wallet' | 'emi';

export type FailureReason =
  | 'BANK_TIMEOUT'
  | 'NETWORK_ERROR'
  | 'INSUFFICIENT_FUNDS'
  | 'CARD_EXPIRED'
  | 'PAYMENT_METHOD_FAILURE'
  | 'CUSTOMER_ABANDONMENT'
  | 'TECHNICAL_ERROR';

export type RecoveryStrategy =
  | 'RETRY_NOW'
  | 'RETRY_LATER'
  | 'SWITCH_PAYMENT_METHOD'
  | 'SEND_PAYMENT_LINK'
  | 'ESCALATE_TO_HUMAN'
  | 'STOP_RECOVERY';

export type PolicyStatus = 'APPROVED' | 'REJECTED' | 'PENDING_HUMAN_APPROVAL';
export type PaymentStatus = 'FAILED' | 'RECOVERED' | 'RECOVERING' | 'ABANDONED';
export type ExecutionMode = 'SIMULATED' | 'MOCK' | 'REAL';

export interface SHAPFactor {
  feature: string;
  impact: number;
  direction: 'positive' | 'negative';
  description: string;
}

export interface PolicyCheck {
  name: string;
  passed: boolean;
  description: string;
}

export interface Payment {
  id: string;
  merchant_id: string;
  customer_id: string;
  amount: number;
  currency: string;
  payment_method: PaymentMethod;
  status: PaymentStatus;
  failure_reason: FailureReason;
  attempt_count: number;
  created_at: string;
}

export interface Customer {
  id: string;
  merchant_id: string;
  email: string;
  phone: string;
  historical_success_rate: number;
  historical_failure_rate: number;
  customer_value: number;
  created_at: string;
}

export interface RecoveryPrediction {
  id: string;
  payment_id: string;
  recovery_probability: number;
  model_version: string;
  confidence: number;
  shap_factors: SHAPFactor[];
  created_at: string;
}

export interface RecoveryDecision {
  id: string;
  payment_id: string;
  strategy: RecoveryStrategy;
  expected_revenue: number;
  expected_cost: number;
  expected_net_value: number;
  priority_score: number;
  explanation: string;
  policy_status: PolicyStatus;
  policy_checks: PolicyCheck[];
  created_at: string;
}

export interface RecoveryAction {
  id: string;
  payment_id: string;
  action_type: string;
  status: string;
  execution_mode: ExecutionMode;
  payload: Record<string, any>;
  executed_at: string;
}

export interface RecoveryOutcome {
  id: string;
  action_id: string;
  payment_id: string;
  success: boolean;
  recovered_amount: number;
  recovery_cost: number;
  net_recovery_value: number;
  completed_at: string;
}

export interface StrategyComparison {
  strategy: RecoveryStrategy;
  probability: number;
  expected_cost: number;
  expected_revenue: number;
  expected_net_value: number;
  is_selected: boolean;
}

export interface FullPaymentRecoveryAnalysis {
  payment: Payment;
  customer: Customer;
  prediction?: RecoveryPrediction;
  decision?: RecoveryDecision;
  action?: RecoveryAction;
  outcome?: RecoveryOutcome;
  alternative_strategies?: StrategyComparison[];
}

export interface OpportunityItem {
  payment_id: string;
  merchant_id: string;
  customer_id: string;
  amount: number;
  currency: string;
  payment_method: PaymentMethod;
  status: PaymentStatus;
  failure_reason: FailureReason;
  attempt_count: number;
  created_at: string;
  recovery_probability: number;
  confidence: number;
  strategy: RecoveryStrategy;
  expected_revenue: number;
  expected_cost: number;
  expected_net_value: number;
  priority_score: number;
  policy_status: PolicyStatus;
  customer_value: number;
  customer_success_rate: number;
}

export interface DashboardOverview {
  kpis: {
    revenue_at_risk: number;
    revenue_recovered: number;
    recovery_rate: number;
    incremental_recovery: number;
    total_failed_payments: number;
    recovered_count: number;
    active_opportunities: number;
    saved_retry_costs: number;
  };
  recovery_over_time: Array<{
    date: string;
    failed_revenue: number;
    recovered_revenue: number;
    baseline_revenue: number;
  }>;
  failure_distribution: Array<{
    reason: string;
    count: number;
    percentage: number;
    avg_amount: number;
  }>;
  strategy_performance: Array<{
    strategy: string;
    count: number;
    success_rate: number;
    net_recovered: number;
    avg_cost: number;
  }>;
}

export interface SimulationStrategyMetrics {
  total_attempts: number;
  successful_recoveries: number;
  recovery_rate: number;
  total_gross_recovered: number;
  total_action_cost: number;
  net_recovery_value: number;
  wasted_retries: number;
  avg_cost_per_recovery: number;
}

export interface SimulationComparisonResponse {
  total_transactions_analyzed: number;
  total_revenue_at_risk: number;
  baseline_strategy: SimulationStrategyMetrics;
  ai_strategy: SimulationStrategyMetrics;
  incremental_comparison: {
    incremental_gross_revenue: number;
    action_cost_reduction: number;
    net_value_uplift: number;
    recovery_rate_gain_pct: number;
    roi_improvement_multiple: number;
  };
  scenario_description: string;
}

export interface MerchantRecoverySettings {
  max_retry_attempts: number;
  min_confidence_threshold: number;
  max_comm_attempts: number;
  human_approval_threshold: number;
  high_value_transaction_threshold: number;
  auto_execution_enabled: boolean;
}
