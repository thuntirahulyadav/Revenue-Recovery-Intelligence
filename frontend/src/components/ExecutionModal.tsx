import React, { useState } from 'react';
import { Play, CheckCircle2, XCircle, AlertTriangle, ShieldCheck, Zap } from 'lucide-react';
import { api } from '../api/client';
import { formatINR, formatPercent, getStrategyBadgeClass } from '../utils/formatters';
import { RecoveryOutcome } from '../types';

interface ExecutionModalProps {
  isOpen: boolean;
  onClose: () => void;
  paymentId: string;
  amount: number;
  strategy: string;
  probability: number;
  expectedNetValue: number;
  onSuccess?: () => void;
}

export const ExecutionModal: React.FC<ExecutionModalProps> = ({
  isOpen,
  onClose,
  paymentId,
  amount,
  strategy,
  probability,
  expectedNetValue,
  onSuccess,
}) => {
  const [executionMode, setExecutionMode] = useState<'SIMULATED' | 'MOCK'>('SIMULATED');
  const [isExecuting, setIsExecuting] = useState(false);
  const [outcome, setOutcome] = useState<RecoveryOutcome | null>(null);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleExecute = async () => {
    setIsExecuting(true);
    setError(null);
    try {
      const res = await api.executeRecovery(paymentId, executionMode);
      setOutcome(res);
      if (onSuccess) {
        onSuccess();
      }
    } catch (err: any) {
      setError(err.message || 'Execution failed');
    } finally {
      setIsExecuting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="w-full max-w-lg bg-surface-900 border border-slate-700 rounded-2xl shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="p-6 border-b border-slate-800 bg-surface-950/60">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <div className="p-2 rounded-lg bg-razor-500/10 border border-razor-500/30 text-razor-400">
                <Zap className="w-5 h-5" />
              </div>
              <div>
                <h3 className="text-lg font-bold text-white">Execute Recovery Action</h3>
                <p className="text-xs text-slate-400 font-mono">Payment ID: {paymentId.slice(0, 8)}...</p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
            >
              ✕
            </button>
          </div>

          {/* Hackathon Requirement: Transparent Simulation Badge */}
          <div className="mt-4 p-2.5 rounded-lg bg-amber-500/10 border border-amber-500/30 flex items-center gap-2 text-amber-300 text-xs">
            <AlertTriangle className="w-4 h-4 shrink-0 text-amber-400" />
            <span>
              <strong>Action Status:</strong> SIMULATED RECOVERY ACTION (Test Sandbox Environment)
            </span>
          </div>
        </div>

        {/* Content */}
        <div className="p-6 space-y-5">
          {!outcome ? (
            <>
              {/* Strategy Card */}
              <div className="p-4 rounded-xl bg-surface-950/80 border border-slate-800 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-xs text-slate-400 font-medium">Selected Strategy</span>
                  <span className={`px-2.5 py-1 rounded-md text-xs font-semibold border ${getStrategyBadgeClass(strategy)}`}>
                    {strategy.replace(/_/g, ' ')}
                  </span>
                </div>
                <div className="grid grid-cols-3 gap-2 pt-2 border-t border-slate-800/80 text-center">
                  <div>
                    <span className="text-[10px] text-slate-500 uppercase tracking-wider">Amount</span>
                    <p className="text-sm font-mono font-bold text-white">{formatINR(amount)}</p>
                  </div>
                  <div>
                    <span className="text-[10px] text-slate-500 uppercase tracking-wider">AI Probability</span>
                    <p className="text-sm font-mono font-bold text-emerald-400">{formatPercent(probability)}</p>
                  </div>
                  <div>
                    <span className="text-[10px] text-slate-500 uppercase tracking-wider">Expected Net</span>
                    <p className="text-sm font-mono font-bold text-razor-400">{formatINR(expectedNetValue)}</p>
                  </div>
                </div>
              </div>

              {/* Execution Mode Selector */}
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-2">Execution Environment</label>
                <div className="grid grid-cols-2 gap-2">
                  {(['SIMULATED', 'MOCK'] as const).map((mode) => (
                    <button
                      key={mode}
                      type="button"
                      onClick={() => setExecutionMode(mode)}
                      className={`px-3 py-2 rounded-lg text-xs font-semibold border transition-all text-center ${
                        executionMode === mode
                          ? 'bg-razor-500/20 border-razor-500 text-white'
                          : 'bg-surface-950 border-slate-800 text-slate-400 hover:border-slate-700'
                      }`}
                    >
                      {mode === 'SIMULATED' && '⚡ Simulated'}
                      {mode === 'MOCK' && '🛡️ Mock Gateway'}
                    </button>
                  ))}
                </div>
              </div>

              {error && (
                <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/30 text-rose-400 text-xs flex items-center gap-2">
                  <XCircle className="w-4 h-4 shrink-0" />
                  <span>{error}</span>
                </div>
              )}
            </>
          ) : (
            /* Outcome Result */
            <div className="text-center py-4 space-y-4 animate-in zoom-in-95 duration-200">
              <div className="inline-flex p-3 rounded-full bg-surface-950 border border-slate-800 shadow-inner">
                {outcome.success ? (
                  <CheckCircle2 className="w-12 h-12 text-emerald-400" />
                ) : (
                  <XCircle className="w-12 h-12 text-rose-400" />
                )}
              </div>

              <div>
                <h4 className="text-lg font-bold text-white">
                  {outcome.success ? 'Recovery Action Succeeded!' : 'Recovery Action Unsuccessful'}
                </h4>
                <p className="text-xs text-slate-400 mt-1">
                  {outcome.success
                    ? `Successfully recovered ${formatINR(outcome.recovered_amount)} with action cost of ${formatINR(outcome.recovery_cost)}.`
                    : `Recovery attempt concluded. Action cost incurred: ${formatINR(outcome.recovery_cost)}.`}
                </p>
              </div>

              <div className="p-4 rounded-xl bg-surface-950 border border-slate-800 grid grid-cols-3 gap-2 text-center">
                <div>
                  <span className="text-[10px] text-slate-500 uppercase">Gross Recovered</span>
                  <p className="text-sm font-mono font-bold text-white">{formatINR(outcome.recovered_amount)}</p>
                </div>
                <div>
                  <span className="text-[10px] text-slate-500 uppercase">Action Cost</span>
                  <p className="text-sm font-mono font-bold text-rose-400">{formatINR(outcome.recovery_cost)}</p>
                </div>
                <div>
                  <span className="text-[10px] text-slate-500 uppercase">Net Value</span>
                  <p className="text-sm font-mono font-bold text-emerald-400">{formatINR(outcome.net_recovery_value)}</p>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-slate-800 bg-surface-950/60 flex items-center justify-end gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-lg text-xs font-semibold text-slate-300 hover:bg-slate-800 transition-colors"
          >
            {outcome ? 'Close' : 'Cancel'}
          </button>
          {!outcome && (
            <button
              onClick={handleExecute}
              disabled={isExecuting}
              className="px-5 py-2 rounded-lg text-xs font-semibold bg-razor-500 hover:bg-razor-600 text-white flex items-center gap-1.5 shadow-lg shadow-razor-500/20 transition-all disabled:opacity-50"
            >
              <Play className="w-3.5 h-3.5" />
              {isExecuting ? 'Executing...' : 'Confirm & Execute'}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};
