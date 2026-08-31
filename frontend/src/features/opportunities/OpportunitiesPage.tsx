import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Search,
  Filter,
  ArrowUpDown,
  ArrowRight,
  Sparkles,
  Zap,
  CheckCircle2,
  AlertTriangle,
  Play,
} from 'lucide-react';
import { Link } from 'react-router-dom';
import { api } from '../../api/client';
import { ExecutionModal } from '../../components/ExecutionModal';
import { formatINR, formatPercent, getStrategyBadgeClass, getFailureReasonBadgeClass } from '../../utils/formatters';
import { OpportunityItem } from '../../types';

export const OpportunitiesPage: React.FC = () => {
  const [search, setSearch] = useState('');
  const [failureReason, setFailureReason] = useState('');
  const [strategy, setStrategy] = useState('');
  const [minProb, setMinProb] = useState<number>(0);
  const [sortBy, setSortBy] = useState('priority_score');
  const [sortOrder, setSortOrder] = useState<'DESC' | 'ASC'>('DESC');
  const [page, setPage] = useState(1);
  const limit = 12;

  // Execution Modal state
  const [selectedOpp, setSelectedOpp] = useState<OpportunityItem | null>(null);

  const { data: opps, isLoading, refetch } = useQuery({
    queryKey: ['opportunities', page, failureReason, strategy, minProb, sortBy, sortOrder, search],
    queryFn: () =>
      api.getOpportunities({
        page,
        limit,
        failure_reason: failureReason || undefined,
        strategy: strategy || undefined,
        min_probability: minProb > 0 ? minProb / 100 : undefined,
        sort_by: sortBy,
        sort_order: sortOrder,
        search: search || undefined,
      }),
  });

  return (
    <div className="space-y-6 animate-in fade-in duration-300">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-black text-white tracking-tight">Recovery Opportunities Queue</h1>
          <p className="text-xs text-slate-400 mt-1">
            Filtered and prioritized queue of failed payment events ready for strategic recovery action.
          </p>
        </div>
      </div>

      {/* Filter & Search Bar */}
      <div className="p-4 rounded-2xl bg-surface-900/90 border border-slate-800 space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
          {/* Search */}
          <div className="relative md:col-span-2">
            <Search className="w-4 h-4 absolute left-3 top-3 text-slate-400" />
            <input
              type="text"
              placeholder="Search by Payment ID, customer email..."
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
              className="w-full pl-9 pr-4 py-2 rounded-xl bg-surface-950 border border-slate-800 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-razor-500 transition-colors"
            />
          </div>

          {/* Failure Reason */}
          <div>
            <select
              value={failureReason}
              onChange={(e) => {
                setFailureReason(e.target.value);
                setPage(1);
              }}
              className="w-full px-3 py-2 rounded-xl bg-surface-950 border border-slate-800 text-xs text-slate-300 focus:outline-none focus:border-razor-500"
            >
              <option value="">All Failure Reasons</option>
              <option value="BANK_TIMEOUT">BANK_TIMEOUT</option>
              <option value="NETWORK_ERROR">NETWORK_ERROR</option>
              <option value="INSUFFICIENT_FUNDS">INSUFFICIENT_FUNDS</option>
              <option value="CARD_EXPIRED">CARD_EXPIRED</option>
              <option value="PAYMENT_METHOD_FAILURE">PAYMENT_METHOD_FAILURE</option>
              <option value="CUSTOMER_ABANDONMENT">CUSTOMER_ABANDONMENT</option>
              <option value="TECHNICAL_ERROR">TECHNICAL_ERROR</option>
            </select>
          </div>

          {/* Strategy */}
          <div>
            <select
              value={strategy}
              onChange={(e) => {
                setStrategy(e.target.value);
                setPage(1);
              }}
              className="w-full px-3 py-2 rounded-xl bg-surface-950 border border-slate-800 text-xs text-slate-300 focus:outline-none focus:border-razor-500"
            >
              <option value="">All AI Strategies</option>
              <option value="RETRY_NOW">RETRY_NOW</option>
              <option value="RETRY_LATER">RETRY_LATER</option>
              <option value="SWITCH_PAYMENT_METHOD">SWITCH_PAYMENT_METHOD</option>
              <option value="SEND_PAYMENT_LINK">SEND_PAYMENT_LINK</option>
              <option value="ESCALATE_TO_HUMAN">ESCALATE_TO_HUMAN</option>
              <option value="STOP_RECOVERY">STOP_RECOVERY</option>
            </select>
          </div>
        </div>

        {/* Secondary Filters */}
        <div className="flex flex-wrap items-center justify-between gap-4 pt-3 border-t border-slate-800/80 text-xs text-slate-400">
          <div className="flex items-center gap-4">
            <span className="font-medium text-slate-300">Min Probability: {minProb}%</span>
            <input
              type="range"
              min="0"
              max="90"
              step="10"
              value={minProb}
              onChange={(e) => {
                setMinProb(Number(e.target.value));
                setPage(1);
              }}
              className="w-32 accent-razor-500"
            />
          </div>

          <div className="flex items-center gap-2">
            <span className="text-slate-400">Sort By:</span>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value)}
              className="px-2.5 py-1 rounded-lg bg-surface-950 border border-slate-800 text-slate-200 text-xs focus:outline-none"
            >
              <option value="priority_score">Priority Score</option>
              <option value="amount">Transaction Amount</option>
              <option value="recovery_probability">Recovery Probability</option>
              <option value="expected_net_value">Expected Net Value</option>
              <option value="created_at">Time Created</option>
            </select>

            <button
              onClick={() => setSortOrder((prev) => (prev === 'DESC' ? 'ASC' : 'DESC'))}
              className="px-2.5 py-1 rounded-lg bg-surface-950 border border-slate-800 text-slate-300 hover:text-white transition-colors"
            >
              {sortOrder === 'DESC' ? '↓ Desc' : '↑ Asc'}
            </button>
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="rounded-2xl bg-surface-900/90 border border-slate-800 overflow-hidden shadow-xl">
        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr>
                <th className="fin-table-header">Payment ID</th>
                <th className="fin-table-header">Amount</th>
                <th className="fin-table-header">Failure Reason</th>
                <th className="fin-table-header">Attempt</th>
                <th className="fin-table-header">AI Probability</th>
                <th className="fin-table-header">Strategy</th>
                <th className="fin-table-header">Expected Net</th>
                <th className="fin-table-header">Priority</th>
                <th className="fin-table-header text-right">Action</th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr>
                  <td colSpan={9} className="text-center py-12 text-slate-400 text-xs">
                    Loading opportunities...
                  </td>
                </tr>
              ) : opps && opps.length > 0 ? (
                opps.map((opp) => (
                  <tr key={opp.payment_id} className="fin-table-row">
                    <td className="fin-table-cell">
                      <Link
                        to={`/payments/${opp.payment_id}`}
                        className="font-mono text-xs text-razor-400 hover:underline font-semibold"
                      >
                        {opp.payment_id.slice(0, 8)}...
                      </Link>
                    </td>
                    <td className="fin-table-cell font-mono font-bold text-white">
                      {formatINR(opp.amount)}
                    </td>
                    <td className="fin-table-cell">
                      <span className={`px-2 py-0.5 rounded text-[11px] font-medium border ${getFailureReasonBadgeClass(opp.failure_reason)}`}>
                        {opp.failure_reason.replace(/_/g, ' ')}
                      </span>
                    </td>
                    <td className="fin-table-cell font-mono text-xs text-slate-400">
                      #{opp.attempt_count}
                    </td>
                    <td className="fin-table-cell">
                      <span className="font-mono text-xs font-bold text-emerald-400">
                        {formatPercent(opp.recovery_probability)}
                      </span>
                    </td>
                    <td className="fin-table-cell">
                      <span className={`px-2 py-0.5 rounded text-[11px] font-semibold border ${getStrategyBadgeClass(opp.strategy)}`}>
                        {opp.strategy.replace(/_/g, ' ')}
                      </span>
                    </td>
                    <td className="fin-table-cell font-mono text-xs text-slate-200">
                      {formatINR(opp.expected_net_value)}
                    </td>
                    <td className="fin-table-cell">
                      <div className="flex items-center gap-2">
                        <span className="font-mono text-xs text-white">{opp.priority_score.toFixed(1)}</span>
                        <div className="w-10 h-1.5 rounded-full bg-slate-800 overflow-hidden">
                          <div
                            className="h-full bg-gradient-to-r from-razor-500 to-emerald-400 rounded-full"
                            style={{ width: `${Math.min(100, opp.priority_score)}%` }}
                          />
                        </div>
                      </div>
                    </td>
                    <td className="fin-table-cell text-right space-x-2">
                      <button
                        onClick={() => setSelectedOpp(opp)}
                        className="px-2.5 py-1 rounded-lg text-xs font-semibold bg-razor-500 hover:bg-razor-600 text-white inline-flex items-center gap-1 transition-colors shadow-sm"
                      >
                        <Play className="w-3 h-3" />
                        <span>Execute</span>
                      </button>
                      <Link
                        to={`/payments/${opp.payment_id}`}
                        className="px-2.5 py-1 rounded-lg text-xs font-semibold bg-surface-950 border border-slate-700 hover:border-slate-600 text-slate-300 inline-flex items-center gap-1 transition-colors"
                      >
                        <span>Details</span>
                        <ArrowRight className="w-3 h-3" />
                      </Link>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={9} className="text-center py-12 text-slate-400 text-xs">
                    No matching failed payment opportunities found. Try adjusting your filters.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination Bar */}
        <div className="p-4 border-t border-slate-800 flex items-center justify-between text-xs text-slate-400">
          <span>Page {page}</span>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-3 py-1.5 rounded-lg bg-surface-950 border border-slate-800 text-slate-300 hover:bg-slate-800 transition-colors disabled:opacity-40"
            >
              Previous
            </button>
            <button
              onClick={() => setPage((p) => p + 1)}
              disabled={!opps || opps.length < limit}
              className="px-3 py-1.5 rounded-lg bg-surface-950 border border-slate-800 text-slate-300 hover:bg-slate-800 transition-colors disabled:opacity-40"
            >
              Next
            </button>
          </div>
        </div>
      </div>

      {/* Execution Modal */}
      {selectedOpp && (
        <ExecutionModal
          isOpen={!!selectedOpp}
          onClose={() => setSelectedOpp(null)}
          paymentId={selectedOpp.payment_id}
          amount={selectedOpp.amount}
          strategy={selectedOpp.strategy}
          probability={selectedOpp.recovery_probability}
          expectedNetValue={selectedOpp.expected_net_value}
          onSuccess={() => {
            refetch();
          }}
        />
      )}
    </div>
  );
};
