import {
  DashboardOverview,
  FullPaymentRecoveryAnalysis,
  OpportunityItem,
  SimulationComparisonResponse,
  MerchantRecoverySettings,
  RecoveryDecision,
  RecoveryOutcome,
} from '../types';

// Use the same origin by default. Vite proxies /api during local development
// and Nginx proxies it in Docker, avoiding a baked-in localhost port mismatch.
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;
  const headers = {
    'Content-Type': 'application/json',
    ...(options.headers || {}),
  };

  const response = await fetch(url, {
    ...options,
    headers,
  });

  const json = await response.json();

  if (!response.ok || !json.success) {
    const errorMsg = json?.error?.message || `Request failed with status ${response.status}`;
    throw new Error(errorMsg);
  }

  return json.data as T;
}

export const api = {
  // Dashboard
  getDashboardOverview: (): Promise<DashboardOverview> => {
    return request<DashboardOverview>('/dashboard/overview');
  },

  // Opportunities
  getOpportunities: (params: {
    page?: number;
    limit?: number;
    failure_reason?: string;
    strategy?: string;
    min_probability?: number;
    min_amount?: number;
    status?: string;
    search?: string;
    sort_by?: string;
    sort_order?: string;
  }): Promise<OpportunityItem[]> => {
    const query = new URLSearchParams();
    if (params.page) query.append('page', params.page.toString());
    if (params.limit) query.append('limit', params.limit.toString());
    if (params.failure_reason) query.append('failure_reason', params.failure_reason);
    if (params.strategy) query.append('strategy', params.strategy);
    if (params.min_probability) query.append('min_probability', params.min_probability.toString());
    if (params.min_amount) query.append('min_amount', params.min_amount.toString());
    if (params.status) query.append('status', params.status);
    if (params.search) query.append('search', params.search);
    if (params.sort_by) query.append('sort_by', params.sort_by);
    if (params.sort_order) query.append('sort_order', params.sort_order);

    return request<OpportunityItem[]>(`/recovery/opportunities?${query.toString()}`);
  },

  // Payment Recovery Details
  getPaymentRecovery: (paymentId: string): Promise<FullPaymentRecoveryAnalysis> => {
    return request<FullPaymentRecoveryAnalysis>(`/payments/${paymentId}/recovery`);
  },

  // Human Policy Approval
  approveRecovery: (paymentId: string): Promise<RecoveryDecision> => {
    return request<RecoveryDecision>(`/recovery/${paymentId}/approve`, {
      method: 'POST',
    });
  },

  // Action Execution (Simulated / Mock / Real)
  executeRecovery: (paymentId: string, executionMode: 'SIMULATED' | 'MOCK'): Promise<RecoveryOutcome> => {
    return request<RecoveryOutcome>(`/recovery/${paymentId}/execute`, {
      method: 'POST',
      body: JSON.stringify({ execution_mode: executionMode }),
    });
  },

  // Simulation Lab
  getSimulationComparison: (sampleSize: number = 2500): Promise<SimulationComparisonResponse> => {
    return request<SimulationComparisonResponse>(`/simulation/compare?sample_size=${sampleSize}`);
  },

  // Settings
  getSettings: (): Promise<MerchantRecoverySettings> => {
    return request<MerchantRecoverySettings>('/settings');
  },

  updateSettings: (settings: MerchantRecoverySettings): Promise<MerchantRecoverySettings> => {
    return request<MerchantRecoverySettings>('/settings', {
      method: 'PUT',
      body: JSON.stringify(settings),
    });
  },

  // Event Ingestion Test
  ingestPaymentFailed: (payload: {
    amount: number;
    payment_method: string;
    failure_reason: string;
    attempt_count?: number;
  }): Promise<FullPaymentRecoveryAnalysis> => {
    return request<FullPaymentRecoveryAnalysis>('/events/payment-failed', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },
};
