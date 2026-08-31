import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  FlaskConical,
  TrendingUp,
  ArrowRight,
  ShieldCheck,
  Zap,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  Play,
  RotateCcw,
} from 'lucide-react';
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
  Cell,
} from 'recharts';
import { api } from '../../api/client';
import { formatCompactINR, formatINR, formatPercent } from '../../utils/formatters';

export const SimulationPage: React.FC = () => {
  const [sampleSize, setSampleSize] = useState(2500);

  const { data: sim, isLoading, refetch, isFetching } = useQuery({
    queryKey: ['simulation-comparison', sampleSize],
    queryFn: () => api.getSimulationComparison(sampleSize),
  });

  if (isLoading || !sim) {
    return (
      <div className="flex items-center justify-center min-h-[400px] text-slate-400 text-sm">
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 border-2 border-razor-500 border-t-transparent rounded-full animate-spin" />
          <span>Simulating 2,500+ transaction cohorts across recovery policies...</span>
        </div>
      </div>
    );
  }

  const { baseline_strategy: baseline, ai_strategy: ai, incremental_comparison: inc } = sim;

  const comparisonChartData = [
    {
      name: 'Gross Recovered',
      Baseline: baseline.total_gross_recovered,
      'Recovery Intelligence': ai.total_gross_recovered,
    },
    {
      name: 'Action Cost',
      Baseline: baseline.total_action_cost,
      'Recovery Intelligence': ai.total_action_cost,
    },
    {
      name: 'Net Recovery Value',
      Baseline: baseline.net_recovery_value,
      'Recovery Intelligence': ai.net_recovery_value,
    },
  ];

  return (
    <div className="space-y-8 animate-in fade-in duration-300">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-black text-white tracking-tight flex items-center gap-2.5">
            <FlaskConical className="w-6 h-6 text-razor-400" />
            <span>Economic Simulation Lab</span>
          </h1>
          <p className="text-xs text-slate-400 mt-1">
            Proving ROI: Baseline Blind Retry Policy vs AI Recovery Intelligence on identical failed payment batches.
          </p>
        </div>

        {/* Cohort selector & Re-run */}
        <div className="flex items-center gap-3">
          <select
            value={sampleSize}
            onChange={(e) => setSampleSize(Number(e.target.value))}
            className="px-3 py-2 rounded-xl bg-surface-900 border border-slate-700 text-xs text-slate-300 focus:outline-none"
          >
            <option value="1000">Cohort: 1,000 Payments</option>
            <option value="2500">Cohort: 2,500 Payments</option>
            <option value="5000">Cohort: 5,000 Payments</option>
            <option value="10000">Cohort: 10,000 Payments</option>
          </select>

          <button
            onClick={() => refetch()}
            disabled={isFetching}
            className="px-4 py-2 rounded-xl text-xs font-semibold bg-razor-500 hover:bg-razor-600 text-white flex items-center gap-1.5 shadow-lg shadow-razor-500/20 transition-all disabled:opacity-50"
          >
            <RotateCcw className={`w-3.5 h-3.5 ${isFetching ? 'animate-spin' : ''}`} />
            <span>Re-run Simulation</span>
          </button>
        </div>
      </div>

      {/* Top Incremental Uplift Banner */}
      <div className="p-6 rounded-2xl bg-gradient-to-r from-emerald-950/40 via-surface-900 to-razor-950/40 border border-emerald-500/30 shadow-2xl space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-emerald-400">
            <TrendingUp className="w-4 h-4" />
            <span>Net Incremental Gain with Recovery Intelligence</span>
          </div>
          <span className="px-2.5 py-0.5 rounded-full text-xs font-mono font-bold bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
            {inc.roi_improvement_multiple}x Net ROI Uplift
          </span>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="p-4 rounded-xl bg-surface-950/80 border border-slate-800">
            <span className="text-[10px] text-slate-500 uppercase">Net Value Uplift</span>
            <p className="text-xl font-mono font-bold text-emerald-400 mt-0.5">+{formatINR(inc.net_value_uplift)}</p>
            <p className="text-[11px] text-slate-400 mt-1">Pure bottom-line profit lift</p>
          </div>

          <div className="p-4 rounded-xl bg-surface-950/80 border border-slate-800">
            <span className="text-[10px] text-slate-500 uppercase">Gross Revenue Gain</span>
            <p className="text-xl font-mono font-bold text-white mt-0.5">+{formatINR(inc.incremental_gross_revenue)}</p>
            <p className="text-[11px] text-slate-400 mt-1">Recovered from tailored channels</p>
          </div>

          <div className="p-4 rounded-xl bg-surface-950/80 border border-slate-800">
            <span className="text-[10px] text-slate-500 uppercase">Action Cost Savings</span>
            <p className="text-xl font-mono font-bold text-purple-400 mt-0.5">+{formatINR(inc.action_cost_reduction)}</p>
            <p className="text-[11px] text-slate-400 mt-1">Avoided futile retry gateway fees</p>
          </div>

          <div className="p-4 rounded-xl bg-surface-950/80 border border-slate-800">
            <span className="text-[10px] text-slate-500 uppercase">Recovery Rate Jump</span>
            <p className="text-xl font-mono font-bold text-razor-400 mt-0.5">+{inc.recovery_rate_gain_pct}%</p>
            <p className="text-[11px] text-slate-400 mt-1">{formatPercent(baseline.recovery_rate)} &rarr; {formatPercent(ai.recovery_rate)}</p>
          </div>
        </div>
      </div>

      {/* Side-by-Side Comparison Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Baseline Card */}
        <div className="p-6 rounded-2xl bg-surface-900 border border-slate-800 space-y-5">
          <div className="flex items-center justify-between pb-3 border-b border-slate-800">
            <div>
              <span className="text-xs font-bold text-slate-400 uppercase tracking-wider block">Policy Approach A</span>
              <h2 className="text-lg font-bold text-white">Baseline (Blind Retry All)</h2>
            </div>
            <span className="px-2.5 py-1 rounded text-xs font-semibold bg-slate-800 text-slate-400 border border-slate-700">
              Generic Strategy
            </span>
          </div>

          <p className="text-xs text-slate-400 leading-relaxed">
            Retries every failed transaction up to 3 times indiscriminately. Incurs continuous gateway retry fees even for expired cards or invalid balances.
          </p>

          <div className="space-y-3 pt-2">
            <div className="flex items-center justify-between text-xs py-2 border-b border-slate-800/60">
              <span className="text-slate-400">Total Retry Attempts:</span>
              <span className="font-mono text-white font-semibold">{baseline.total_attempts.toLocaleString()}</span>
            </div>
            <div className="flex items-center justify-between text-xs py-2 border-b border-slate-800/60">
              <span className="text-slate-400">Recovery Rate:</span>
              <span className="font-mono text-slate-300 font-bold">{formatPercent(baseline.recovery_rate)}</span>
            </div>
            <div className="flex items-center justify-between text-xs py-2 border-b border-slate-800/60">
              <span className="text-slate-400">Gross Recovered Revenue:</span>
              <span className="font-mono text-white font-bold">{formatINR(baseline.total_gross_recovered)}</span>
            </div>
            <div className="flex items-center justify-between text-xs py-2 border-b border-slate-800/60">
              <span className="text-slate-400">Total Action & Retry Costs:</span>
              <span className="font-mono text-rose-400 font-bold">{formatINR(baseline.total_action_cost)}</span>
            </div>
            <div className="flex items-center justify-between text-xs py-2 border-b border-slate-800/60">
              <span className="text-slate-400">Wasted Futile Retries:</span>
              <span className="font-mono text-amber-400 font-semibold">{baseline.wasted_retries.toLocaleString()}</span>
            </div>
            <div className="flex items-center justify-between text-sm py-2 font-bold">
              <span className="text-slate-300">Net Recovery Value:</span>
              <span className="font-mono text-slate-100 text-base">{formatINR(baseline.net_recovery_value)}</span>
            </div>
          </div>
        </div>

        {/* AI Strategic Recovery Card */}
        <div className="p-6 rounded-2xl bg-surface-900 border border-razor-500/40 shadow-xl space-y-5 relative overflow-hidden">
          <div className="absolute top-0 right-0 w-32 h-32 bg-razor-500/10 blur-2xl pointer-events-none" />

          <div className="flex items-center justify-between pb-3 border-b border-slate-800">
            <div>
              <span className="text-xs font-bold text-razor-400 uppercase tracking-wider block">Policy Approach B</span>
              <h2 className="text-lg font-bold text-white">Recovery Intelligence (RRI)</h2>
            </div>
            <span className="px-2.5 py-1 rounded text-xs font-bold bg-razor-500/20 text-razor-300 border border-razor-500/40">
              AI Decision Engine
            </span>
          </div>

          <p className="text-xs text-slate-400 leading-relaxed">
            Routes failures to optimal channels (smart delay, method switch, payment links), while aborting non-viable transactions to save cost.
          </p>

          <div className="space-y-3 pt-2">
            <div className="flex items-center justify-between text-xs py-2 border-b border-slate-800/60">
              <span className="text-slate-400">Total Strategy Actions:</span>
              <span className="font-mono text-white font-semibold">{ai.total_attempts.toLocaleString()}</span>
            </div>
            <div className="flex items-center justify-between text-xs py-2 border-b border-slate-800/60">
              <span className="text-slate-400">Recovery Rate:</span>
              <span className="font-mono text-emerald-400 font-bold">{formatPercent(ai.recovery_rate)}</span>
            </div>
            <div className="flex items-center justify-between text-xs py-2 border-b border-slate-800/60">
              <span className="text-slate-400">Gross Recovered Revenue:</span>
              <span className="font-mono text-white font-bold">{formatINR(ai.total_gross_recovered)}</span>
            </div>
            <div className="flex items-center justify-between text-xs py-2 border-b border-slate-800/60">
              <span className="text-slate-400">Total Strategy Execution Costs:</span>
              <span className="font-mono text-purple-400 font-bold">{formatINR(ai.total_action_cost)}</span>
            </div>
            <div className="flex items-center justify-between text-xs py-2 border-b border-slate-800/60">
              <span className="text-slate-400">Wasted Retries Avoided:</span>
              <span className="font-mono text-emerald-400 font-semibold">{(baseline.wasted_retries - ai.wasted_retries).toLocaleString()} saved</span>
            </div>
            <div className="flex items-center justify-between text-sm py-2 font-bold">
              <span className="text-emerald-400">Net Recovery Value:</span>
              <span className="font-mono text-emerald-400 text-lg font-black">{formatINR(ai.net_recovery_value)}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Comparison Visual Chart */}
      <div className="p-6 rounded-2xl bg-surface-900 border border-slate-800 shadow-xl space-y-4">
        <div>
          <h2 className="text-sm font-bold text-white tracking-wide">Comparative Economic Breakdown</h2>
          <p className="text-xs text-slate-400">Comparing Gross Revenue, Strategy Overhead Cost, and Net Value</p>
        </div>

        <div className="h-80 w-full pt-4">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={comparisonChartData} margin={{ top: 20, right: 20, left: 10, bottom: 5 }}>
              <XAxis dataKey="name" stroke="#64748b" fontSize={12} tickLine={false} />
              <YAxis
                stroke="#64748b"
                fontSize={12}
                tickLine={false}
                tickFormatter={(val) => formatCompactINR(val)}
              />
              <Tooltip
                contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', fontSize: '12px' }}
                formatter={(val: any) => [formatINR(Number(val)), '']}
              />
              <Legend />
              <Bar dataKey="Baseline" fill="#64748b" radius={[6, 6, 0, 0]} />
              <Bar dataKey="Recovery Intelligence" fill="#0c8cee" radius={[6, 6, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
};
