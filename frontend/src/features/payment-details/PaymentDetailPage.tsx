import React, { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  ArrowLeft,
  AlertOctagon,
  ShieldCheck,
  CheckCircle2,
  XCircle,
  TrendingUp,
  BrainCircuit,
  Zap,
  Play,
  Clock,
  Layers,
  Sparkles,
  Info,
} from 'lucide-react';
import { api } from '../../api/client';
import { ExecutionModal } from '../../components/ExecutionModal';
import { formatINR, formatPercent, formatDate, getStrategyBadgeClass, getFailureReasonBadgeClass } from '../../utils/formatters';

export const PaymentDetailPage: React.FC = () => {
  const { paymentId } = useParams<{ paymentId: string }>();
  const [isExecuteModalOpen, setIsExecuteModalOpen] = useState(false);

  const { data: analysis, isLoading, error, refetch } = useQuery({
    queryKey: ['payment-recovery', paymentId],
    queryFn: () => (paymentId ? api.getPaymentRecovery(paymentId) : Promise.reject('No ID')),
    enabled: !!paymentId,
    refetchInterval: 6000,
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px] text-slate-400 text-sm">
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 border-2 border-razor-500 border-t-transparent rounded-full animate-spin" />
          <span>Analyzing payment failure patterns with XGBoost & SHAP...</span>
        </div>
      </div>
    );
  }

  if (error || !analysis) {
    return (
      <div className="p-6 rounded-2xl bg-rose-500/10 border border-rose-500/30 text-rose-400 text-sm space-y-3">
        <p>Failed to retrieve payment recovery analysis for ID {paymentId}.</p>
        <Link to="/opportunities" className="text-xs font-semibold text-white underline">
          &larr; Return to opportunities
        </Link>
      </div>
    );
  }

  const { payment, customer, prediction, decision, action, outcome, alternative_strategies } = analysis;

  const handleApprove = async () => {
    if (!paymentId) return;
    try {
      await api.approveRecovery(paymentId);
      refetch();
    } catch (err: any) {
      alert(`Approval error: ${err.message}`);
    }
  };

  const positiveShap = prediction?.shap_factors?.filter((f) => f.direction === 'positive') || [];
  const negativeShap = prediction?.shap_factors?.filter((f) => f.direction === 'negative') || [];

  return (
    <div className="space-y-8 animate-in fade-in duration-300">
      {/* Back link */}
      <div>
        <Link
          to="/opportunities"
          className="text-xs font-semibold text-slate-400 hover:text-white inline-flex items-center gap-1.5 transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Back to Opportunities</span>
        </Link>
      </div>

      {/* 1. Failed Payment Banner */}
      <div className="p-6 rounded-2xl bg-surface-900 border border-slate-800 shadow-xl space-y-5">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
          <div className="flex items-start gap-3.5">
            <div className="p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-400 mt-1">
              <AlertOctagon className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-xs font-bold uppercase tracking-wider px-2 py-0.5 rounded bg-rose-500/20 text-rose-400 border border-rose-500/30">
                  PAYMENT {payment.status}
                </span>
                <span className="text-xs text-slate-400 font-mono">ID: {payment.id}</span>
              </div>
              <h1 className="text-3xl font-black text-white font-mono mt-1">{formatINR(payment.amount)}</h1>
            </div>
          </div>

          {/* Quick Action Button */}
          <div className="flex items-center gap-3">
            {decision?.policy_status === 'PENDING_HUMAN_APPROVAL' && (
              <button
                onClick={handleApprove}
                className="px-4 py-2.5 rounded-xl text-xs font-semibold bg-amber-500 hover:bg-amber-600 text-white flex items-center gap-1.5 shadow-lg shadow-amber-500/20 transition-all"
              >
                <ShieldCheck className="w-4 h-4" />
                <span>Authorize High-Value Policy</span>
              </button>
            )}

            <button
              onClick={() => setIsExecuteModalOpen(true)}
              className="px-5 py-2.5 rounded-xl text-xs font-semibold bg-razor-500 hover:bg-razor-600 text-white flex items-center gap-2 shadow-lg shadow-razor-500/25 transition-all active:scale-95"
            >
              <Play className="w-4 h-4" />
              <span>Execute Recovery Action</span>
            </button>
          </div>
        </div>

        {/* Payment Metadata Grid */}
        <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-6 gap-3 pt-4 border-t border-slate-800/80 text-xs">
          <div>
            <span className="text-slate-500 uppercase tracking-wider text-[10px]">Payment Method</span>
            <p className="font-semibold text-slate-200 uppercase mt-0.5">{payment.payment_method}</p>
          </div>
          <div>
            <span className="text-slate-500 uppercase tracking-wider text-[10px]">Failure Reason</span>
            <p className="mt-0.5">
              <span className={`px-2 py-0.5 rounded text-[11px] font-medium border ${getFailureReasonBadgeClass(payment.failure_reason)}`}>
                {payment.failure_reason.replace(/_/g, ' ')}
              </span>
            </p>
          </div>
          <div>
            <span className="text-slate-500 uppercase tracking-wider text-[10px]">Attempt Count</span>
            <p className="font-mono font-semibold text-slate-200 mt-0.5">Attempt #{payment.attempt_count}</p>
          </div>
          <div>
            <span className="text-slate-500 uppercase tracking-wider text-[10px]">Customer Value</span>
            <p className="font-mono font-semibold text-slate-200 mt-0.5">{formatINR(customer.customer_value)}</p>
          </div>
          <div>
            <span className="text-slate-500 uppercase tracking-wider text-[10px]">Historical Success</span>
            <p className="font-mono font-semibold text-emerald-400 mt-0.5">{formatPercent(customer.historical_success_rate)}</p>
          </div>
          <div>
            <span className="text-slate-500 uppercase tracking-wider text-[10px]">Timestamp</span>
            <p className="font-mono text-slate-400 mt-0.5">{formatDate(payment.created_at)}</p>
          </div>
        </div>
      </div>

      {/* 2. AI Recovery Scorecard & Recommended Strategy */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* AI Scorecard */}
        <div className="p-6 rounded-2xl bg-surface-900 border border-slate-800 shadow-xl space-y-4">
          <div className="flex items-center gap-2 text-razor-400 text-xs font-bold uppercase tracking-wider">
            <BrainCircuit className="w-4 h-4" />
            <span>AI Recovery Analysis</span>
          </div>

          <div className="text-center py-3">
            <div className="inline-block relative">
              <span className="text-5xl font-black font-mono text-emerald-400 tracking-tight">
                {formatPercent(prediction?.recovery_probability || 0)}
              </span>
              <p className="text-xs text-slate-400 mt-1 font-medium">Estimated Recovery Probability</p>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3 pt-3 border-t border-slate-800/80 text-center">
            <div className="p-3 rounded-xl bg-surface-950 border border-slate-800">
              <span className="text-[10px] text-slate-500 uppercase">ML Confidence</span>
              <p className="text-base font-mono font-bold text-white mt-0.5">
                {formatPercent(prediction?.confidence || 0)}
              </p>
            </div>
            <div className="p-3 rounded-xl bg-surface-950 border border-slate-800">
              <span className="text-[10px] text-slate-500 uppercase">Priority Score</span>
              <p className="text-base font-mono font-bold text-razor-400 mt-0.5">
                {decision?.priority_score?.toFixed(1) || '0.0'} / 100
              </p>
            </div>
          </div>
        </div>

        {/* Recommended Strategy with Economic Formula */}
        <div className="lg:col-span-2 p-6 rounded-2xl bg-gradient-to-br from-razor-950/40 via-surface-900 to-surface-900 border border-razor-500/30 shadow-xl space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-razor-400">
              <Sparkles className="w-4 h-4" />
              <span>Recommended Recovery Strategy</span>
            </div>
            <span className={`px-3 py-1 rounded-lg text-xs font-bold border ${getStrategyBadgeClass(decision?.strategy || '')}`}>
              {decision?.strategy.replace(/_/g, ' ')}
            </span>
          </div>

          <p className="text-xs text-slate-300 leading-relaxed font-medium bg-surface-950/60 p-3 rounded-xl border border-slate-800">
            {decision?.explanation}
          </p>

          {/* Economic Formula Breakdown */}
          <div className="space-y-2 pt-2">
            <div className="flex items-center justify-between text-xs text-slate-400">
              <span className="font-mono text-[11px] text-slate-400">
                Formula: Net Value = (Amount &times; P_strat) - Action Cost
              </span>
              <span className="text-slate-400">Economic Value Matrix</span>
            </div>

            <div className="grid grid-cols-3 gap-3">
              <div className="p-3 rounded-xl bg-surface-950 border border-slate-800 text-center">
                <span className="text-[10px] text-slate-500 uppercase">Expected Revenue</span>
                <p className="text-sm font-mono font-bold text-white mt-0.5">
                  {formatINR(decision?.expected_revenue || 0)}
                </p>
              </div>
              <div className="p-3 rounded-xl bg-surface-950 border border-slate-800 text-center">
                <span className="text-[10px] text-slate-500 uppercase">Strategy Cost</span>
                <p className="text-sm font-mono font-bold text-rose-400 mt-0.5">
                  {formatINR(decision?.expected_cost || 0)}
                </p>
              </div>
              <div className="p-3 rounded-xl bg-surface-950 border border-emerald-500/30 text-center">
                <span className="text-[10px] text-emerald-400 uppercase font-semibold">Expected Net Value</span>
                <p className="text-sm font-mono font-bold text-emerald-400 mt-0.5">
                  {formatINR(decision?.expected_net_value || 0)}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* 3. Why? SHAP Model Explainability */}
      <div className="p-6 rounded-2xl bg-surface-900 border border-slate-800 shadow-xl space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 text-white font-bold text-sm">
            <Info className="w-4 h-4 text-razor-400" />
            <span>Why? SHAP Model Feature Contributions</span>
          </div>
          <span className="text-xs text-slate-400 font-mono">TreeExplainer v1.0.0</span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-2">
          {/* Positive Factors */}
          <div className="p-4 rounded-xl bg-emerald-950/15 border border-emerald-500/20 space-y-2.5">
            <span className="text-xs font-bold text-emerald-400 uppercase tracking-wider flex items-center gap-1.5">
              <CheckCircle2 className="w-4 h-4" />
              <span>Positive Drivers (+ Lift)</span>
            </span>
            <div className="space-y-2">
              {positiveShap.length > 0 ? (
                positiveShap.map((f, i) => (
                  <div key={i} className="flex items-center justify-between text-xs bg-surface-950/60 p-2 rounded-lg border border-emerald-500/10">
                    <span className="text-slate-300">{f.description}</span>
                    <span className="font-mono text-emerald-400 font-bold">+{f.impact.toFixed(3)}</span>
                  </div>
                ))
              ) : (
                <p className="text-xs text-slate-400">Standard baseline features.</p>
              )}
            </div>
          </div>

          {/* Negative Factors */}
          <div className="p-4 rounded-xl bg-rose-950/15 border border-rose-500/20 space-y-2.5">
            <span className="text-xs font-bold text-rose-400 uppercase tracking-wider flex items-center gap-1.5">
              <XCircle className="w-4 h-4" />
              <span>Risk & Negative Constraints (- Drag)</span>
            </span>
            <div className="space-y-2">
              {negativeShap.length > 0 ? (
                negativeShap.map((f, i) => (
                  <div key={i} className="flex items-center justify-between text-xs bg-surface-950/60 p-2 rounded-lg border border-rose-500/10">
                    <span className="text-slate-300">{f.description}</span>
                    <span className="font-mono text-rose-400 font-bold">{f.impact.toFixed(3)}</span>
                  </div>
                ))
              ) : (
                <p className="text-xs text-slate-400">Zero major risk constraints identified.</p>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* 4. Policy Engine Validation Checklist */}
      <div className="p-6 rounded-2xl bg-surface-900 border border-slate-800 shadow-xl space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 text-white font-bold text-sm">
            <ShieldCheck className="w-4 h-4 text-emerald-400" />
            <span>Policy Engine Validation & Compliance Checks</span>
          </div>
          <span
            className={`px-2.5 py-0.5 rounded text-xs font-bold ${
              decision?.policy_status === 'APPROVED'
                ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30'
                : decision?.policy_status === 'PENDING_HUMAN_APPROVAL'
                ? 'bg-amber-500/10 text-amber-400 border border-amber-500/30'
                : 'bg-rose-500/10 text-rose-400 border border-rose-500/30'
            }`}
          >
            Policy: {decision?.policy_status}
          </span>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 pt-2">
          {decision?.policy_checks?.map((chk, i) => (
            <div
              key={i}
              className={`p-3.5 rounded-xl border flex items-start gap-2.5 ${
                chk.passed
                  ? 'bg-surface-950 border-emerald-500/30 text-emerald-400'
                  : 'bg-surface-950 border-rose-500/30 text-rose-400'
              }`}
            >
              {chk.passed ? <CheckCircle2 className="w-4 h-4 shrink-0 mt-0.5" /> : <XCircle className="w-4 h-4 shrink-0 mt-0.5" />}
              <div>
                <span className="text-xs font-bold text-white block">{chk.name}</span>
                <p className="text-[11px] text-slate-400 mt-0.5 leading-snug">{chk.description}</p>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* 5. Alternative Strategy Comparison Matrix */}
      {alternative_strategies && alternative_strategies.length > 0 && (
        <div className="p-6 rounded-2xl bg-surface-900 border border-slate-800 shadow-xl space-y-4">
          <div>
            <h3 className="text-sm font-bold text-white">Full Strategy Candidate Matrix</h3>
            <p className="text-xs text-slate-400">Comparing expected net values across all recovery avenues</p>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr>
                  <th className="fin-table-header">Strategy Avenue</th>
                  <th className="fin-table-header">Channel Probability</th>
                  <th className="fin-table-header">Expected Cost</th>
                  <th className="fin-table-header">Expected Gross</th>
                  <th className="fin-table-header">Net Recovery Value</th>
                  <th className="fin-table-header text-right">Status</th>
                </tr>
              </thead>
              <tbody>
                {alternative_strategies.map((strat) => {
                  const isSelected = strat.strategy === decision?.strategy;
                  return (
                    <tr
                      key={strat.strategy}
                      className={`fin-table-row ${isSelected ? 'bg-razor-500/10 border-razor-500/30' : ''}`}
                    >
                      <td className="fin-table-cell">
                        <span className={`px-2 py-0.5 rounded text-xs font-semibold border ${getStrategyBadgeClass(strat.strategy)}`}>
                          {strat.strategy.replace(/_/g, ' ')}
                        </span>
                      </td>
                      <td className="fin-table-cell font-mono text-xs text-slate-300">
                        {formatPercent(strat.probability)}
                      </td>
                      <td className="fin-table-cell font-mono text-xs text-rose-400">
                        {formatINR(strat.expected_cost)}
                      </td>
                      <td className="fin-table-cell font-mono text-xs text-white">
                        {formatINR(strat.expected_revenue)}
                      </td>
                      <td className="fin-table-cell font-mono text-xs font-bold text-emerald-400">
                        {formatINR(strat.expected_net_value)}
                      </td>
                      <td className="fin-table-cell text-right">
                        {isSelected ? (
                          <span className="px-2 py-0.5 rounded bg-razor-500 text-white font-bold text-[10px] uppercase">
                            AI Picked
                          </span>
                        ) : (
                          <span className="text-[11px] text-slate-500">Suboptimal</span>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Execution Modal */}
      {payment && decision && (
        <ExecutionModal
          isOpen={isExecuteModalOpen}
          onClose={() => setIsExecuteModalOpen(false)}
          paymentId={payment.id}
          amount={payment.amount}
          strategy={decision.strategy}
          probability={prediction?.recovery_probability || 0.5}
          expectedNetValue={decision.expected_net_value}
          onSuccess={() => refetch()}
        />
      )}
    </div>
  );
};
