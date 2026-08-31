import React from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  AlertCircle,
  TrendingUp,
  ShieldAlert,
  Sparkles,
  ArrowUpRight,
  PiggyBank,
  CheckCircle2,
  Clock,
  Layers,
} from 'lucide-react';
import { Link } from 'react-router-dom';
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  BarChart,
  Bar,
  Cell,
  Legend,
} from 'recharts';
import { api } from '../../api/client';
import { KPICard } from '../../components/KPICard';
import { formatCompactINR, formatINR, formatPercent, getStrategyBadgeClass, getFailureReasonBadgeClass } from '../../utils/formatters';

const REASON_COLORS: Record<string, string> = {
  BANK_TIMEOUT: '#06b6d4',
  NETWORK_ERROR: '#38bdf8',
  INSUFFICIENT_FUNDS: '#f59e0b',
  CARD_EXPIRED: '#f43f5e',
  PAYMENT_METHOD_FAILURE: '#a855f7',
  CUSTOMER_ABANDONMENT: '#fb923c',
  TECHNICAL_ERROR: '#ef4444',
};

export const DashboardPage: React.FC = () => {
  const { data: overview, isLoading, error } = useQuery({
    queryKey: ['dashboard-overview'],
    queryFn: api.getDashboardOverview,
    refetchInterval: 10000,
  });

  const { data: topOpps } = useQuery({
    queryKey: ['top-opportunities'],
    queryFn: () => api.getOpportunities({ limit: 6, sort_by: 'priority_score', sort_order: 'DESC' }),
    refetchInterval: 10000,
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px] text-slate-400 text-sm">
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 border-2 border-razor-500 border-t-transparent rounded-full animate-spin" />
          <span>Loading Recovery Command Center metrics...</span>
        </div>
      </div>
    );
  }

  if (error || !overview) {
    return (
      <div className="p-6 rounded-2xl bg-rose-500/10 border border-rose-500/30 text-rose-400 text-sm">
        Failed to load dashboard overview. Ensure backend gateway is running.
      </div>
    );
  }

  const { kpis, recovery_over_time, failure_distribution, strategy_performance } = overview;

  return (
    <div className="space-y-8 animate-in fade-in duration-300">
      {/* Page Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-black text-white tracking-tight flex items-center gap-2.5">
            <span>Recovery Command Center</span>
            <span className="text-xs font-semibold px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
              Live Stream
            </span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Real-time failed payment intelligence, ML recovery probability & net economic decisioning.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Link
            to="/opportunities"
            className="px-4 py-2 rounded-xl text-xs font-semibold bg-surface-900 border border-slate-700 hover:border-slate-600 text-white flex items-center gap-1.5 transition-colors"
          >
            <span>View All Opportunities</span>
            <ArrowUpRight className="w-4 h-4" />
          </Link>
        </div>
      </div>

      {/* KPI Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        <KPICard
          title="Revenue At Risk"
          value={formatCompactINR(kpis.revenue_at_risk)}
          subtitle={`${kpis.total_failed_payments.toLocaleString()} failed transactions`}
          icon={AlertCircle}
          variant="amber"
        />
        <KPICard
          title="Revenue Recovered"
          value={formatCompactINR(kpis.revenue_recovered)}
          subtitle={`${kpis.recovered_count.toLocaleString()} successfully recovered`}
          icon={CheckCircle2}
          variant="emerald"
          trend={{ value: `${formatPercent(kpis.recovery_rate)} rate`, isPositive: true }}
        />
        <KPICard
          title="Incremental Recovery"
          value={formatCompactINR(kpis.incremental_recovery)}
          subtitle="Net revenue gained over baseline blind retry"
          icon={TrendingUp}
          variant="blue"
          trend={{ value: '+36.0% uplift', isPositive: true }}
        />
        <KPICard
          title="Wasted Costs Saved"
          value={formatCompactINR(kpis.saved_retry_costs)}
          subtitle="Unproductive retry fees eliminated by AI"
          icon={PiggyBank}
          variant="purple"
        />
      </div>

      {/* Main Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left: Revenue Recovery Over Time */}
        <div className="lg:col-span-2 p-6 rounded-2xl bg-surface-900/90 border border-slate-800/80 shadow-xl space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-sm font-bold text-white tracking-wide">Revenue Recovery Over Time</h2>
              <p className="text-xs text-slate-400">Comparing Total Failed, AI Recovered, and Estimated Baseline</p>
            </div>
            <div className="flex items-center gap-4 text-xs">
              <div className="flex items-center gap-1.5 text-slate-400">
                <span className="w-2.5 h-2.5 rounded-full bg-slate-500" />
                <span>Failed</span>
              </div>
              <div className="flex items-center gap-1.5 text-slate-400">
                <span className="w-2.5 h-2.5 rounded-full bg-emerald-400" />
                <span>AI Recovered</span>
              </div>
              <div className="flex items-center gap-1.5 text-slate-400">
                <span className="w-2.5 h-2.5 rounded-full bg-razor-400" />
                <span>Baseline</span>
              </div>
            </div>
          </div>

          <div className="h-72 w-full pt-4">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={recovery_over_time} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorRecovered" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#10b981" stopOpacity={0.4} />
                    <stop offset="95%" stopColor="#10b981" stopOpacity={0.0} />
                  </linearGradient>
                  <linearGradient id="colorFailed" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#64748b" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#64748b" stopOpacity={0.0} />
                  </linearGradient>
                </defs>
                <XAxis dataKey="date" stroke="#64748b" fontSize={11} tickLine={false} />
                <YAxis
                  stroke="#64748b"
                  fontSize={11}
                  tickLine={false}
                  tickFormatter={(val) => formatCompactINR(val)}
                />
                <Tooltip
                  contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', fontSize: '12px' }}
                  formatter={(val: any) => [formatINR(Number(val)), '']}
                />
                <Area type="monotone" dataKey="failed_revenue" stroke="#64748b" fillOpacity={1} fill="url(#colorFailed)" name="Failed" />
                <Area type="monotone" dataKey="recovered_revenue" stroke="#10b981" strokeWidth={2} fillOpacity={1} fill="url(#colorRecovered)" name="AI Recovered" />
                <Area type="monotone" dataKey="baseline_revenue" stroke="#0c8cee" strokeDasharray="3 3" fill="none" name="Baseline" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Right: Failure Reason Distribution */}
        <div className="p-6 rounded-2xl bg-surface-900/90 border border-slate-800/80 shadow-xl space-y-4">
          <div>
            <h2 className="text-sm font-bold text-white tracking-wide">Failure Breakdown</h2>
            <p className="text-xs text-slate-400">Distribution by root failure cause</p>
          </div>

          <div className="space-y-3 pt-2">
            {failure_distribution.slice(0, 5).map((item) => (
              <div key={item.reason} className="space-y-1.5">
                <div className="flex items-center justify-between text-xs">
                  <span className="font-medium text-slate-300 truncate">{item.reason.replace(/_/g, ' ')}</span>
                  <span className="font-mono text-slate-400">{formatPercent(item.percentage)}</span>
                </div>
                <div className="w-full h-2 rounded-full bg-slate-800 overflow-hidden">
                  <div
                    className="h-full rounded-full transition-all duration-500"
                    style={{
                      width: `${item.percentage * 100}%`,
                      backgroundColor: REASON_COLORS[item.reason] || '#0c8cee',
                    }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Strategy Performance Matrix */}
      <div className="p-6 rounded-2xl bg-surface-900/90 border border-slate-800/80 shadow-xl space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-sm font-bold text-white tracking-wide">Recovery Strategy Performance</h2>
            <p className="text-xs text-slate-400">Success rate and net value recovered per strategy channel</p>
          </div>
          <Link to="/simulation" className="text-xs text-razor-400 hover:text-razor-300 flex items-center gap-1 font-semibold">
            <span>Open Simulation Lab</span>
            <ArrowUpRight className="w-3.5 h-3.5" />
          </Link>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-6 gap-3 pt-2">
          {strategy_performance.map((strat) => (
            <div key={strat.strategy} className="p-4 rounded-xl bg-surface-950/80 border border-slate-800 space-y-2">
              <span className={`px-2 py-0.5 rounded text-[10px] font-bold border block truncate ${getStrategyBadgeClass(strat.strategy)}`}>
                {strat.strategy.replace(/_/g, ' ')}
              </span>
              <div>
                <span className="text-[10px] text-slate-500 uppercase">Success Rate</span>
                <p className="text-base font-mono font-bold text-white">{formatPercent(strat.success_rate)}</p>
              </div>
              <div className="pt-1 border-t border-slate-800/80 flex items-center justify-between text-[11px]">
                <span className="text-slate-400">Net:</span>
                <span className="font-mono font-semibold text-emerald-400">{formatCompactINR(strat.net_recovered)}</span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Top Urgent Priority Opportunities Table */}
      <div className="p-6 rounded-2xl bg-surface-900/90 border border-slate-800/80 shadow-xl space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-sm font-bold text-white tracking-wide">Priority Recovery Queue</h2>
            <p className="text-xs text-slate-400">High-value failures prioritized by ML recovery probability</p>
          </div>
          <Link to="/opportunities" className="text-xs text-razor-400 hover:text-razor-300 font-semibold flex items-center gap-1">
            <span>View all queue</span>
            <ArrowUpRight className="w-3.5 h-3.5" />
          </Link>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr>
                <th className="fin-table-header">Payment ID</th>
                <th className="fin-table-header">Amount</th>
                <th className="fin-table-header">Failure Reason</th>
                <th className="fin-table-header">AI Probability</th>
                <th className="fin-table-header">Priority</th>
                <th className="fin-table-header">Recommended Strategy</th>
                <th className="fin-table-header text-right">Action</th>
              </tr>
            </thead>
            <tbody>
              {topOpps?.slice(0, 6).map((opp) => (
                <tr key={opp.payment_id} className="fin-table-row">
                  <td className="fin-table-cell font-mono text-xs text-slate-300">
                    {opp.payment_id.slice(0, 8)}...
                  </td>
                  <td className="fin-table-cell font-mono font-bold text-white">
                    {formatINR(opp.amount)}
                  </td>
                  <td className="fin-table-cell">
                    <span className={`px-2 py-0.5 rounded text-[11px] font-medium border ${getFailureReasonBadgeClass(opp.failure_reason)}`}>
                      {opp.failure_reason.replace(/_/g, ' ')}
                    </span>
                  </td>
                  <td className="fin-table-cell font-mono text-emerald-400 font-semibold">
                    {formatPercent(opp.recovery_probability)}
                  </td>
                  <td className="fin-table-cell">
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-xs text-white">{opp.priority_score.toFixed(1)}</span>
                      <div className="w-12 h-1.5 rounded-full bg-slate-800 overflow-hidden">
                        <div
                          className="h-full bg-gradient-to-r from-razor-500 to-emerald-400 rounded-full"
                          style={{ width: `${Math.min(100, opp.priority_score)}%` }}
                        />
                      </div>
                    </div>
                  </td>
                  <td className="fin-table-cell">
                    <span className={`px-2 py-0.5 rounded text-[11px] font-semibold border ${getStrategyBadgeClass(opp.strategy)}`}>
                      {opp.strategy.replace(/_/g, ' ')}
                    </span>
                  </td>
                  <td className="fin-table-cell text-right">
                    <Link
                      to={`/payments/${opp.payment_id}`}
                      className="px-3 py-1 rounded-lg text-xs font-semibold bg-razor-500/10 hover:bg-razor-500/20 text-razor-400 border border-razor-500/30 transition-colors inline-flex items-center gap-1"
                    >
                      <span>Analyze</span>
                      <ArrowUpRight className="w-3.5 h-3.5" />
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
