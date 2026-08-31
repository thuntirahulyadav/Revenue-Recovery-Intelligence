import React, { useState, useEffect } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Settings, Shield, Sliders, CheckCircle2, Save, AlertCircle } from 'lucide-react';
import { api } from '../../api/client';
import { MerchantRecoverySettings } from '../../types';
import { formatINR } from '../../utils/formatters';

export const SettingsPage: React.FC = () => {
  const { data: settingsData, isLoading } = useQuery({
    queryKey: ['merchant-settings'],
    queryFn: api.getSettings,
  });

  const [settings, setSettings] = useState<MerchantRecoverySettings>({
    max_retry_attempts: 3,
    min_confidence_threshold: 0.65,
    max_comm_attempts: 2,
    human_approval_threshold: 50000,
    high_value_transaction_threshold: 25000,
    auto_execution_enabled: true,
  });

  const [savedSuccess, setSavedSuccess] = useState(false);

  useEffect(() => {
    if (settingsData) {
      setSettings(settingsData);
    }
  }, [settingsData]);

  const updateMutation = useMutation({
    mutationFn: (newSettings: MerchantRecoverySettings) => api.updateSettings(newSettings),
    onSuccess: () => {
      setSavedSuccess(true);
      setTimeout(() => setSavedSuccess(false), 3000);
    },
  });

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    updateMutation.mutate(settings);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px] text-slate-400 text-sm">
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 border-2 border-razor-500 border-t-transparent rounded-full animate-spin" />
          <span>Loading merchant policy settings...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl space-y-8 animate-in fade-in duration-300">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-black text-white tracking-tight flex items-center gap-2.5">
          <Settings className="w-6 h-6 text-razor-400" />
          <span>Merchant Recovery Policy Settings</span>
        </h1>
        <p className="text-xs text-slate-400 mt-1">
          Configure rule thresholds, policy safety gates, and automated execution constraints for Razorpay Recovery Intelligence.
        </p>
      </div>

      <form onSubmit={handleSave} className="space-y-6">
        {/* Core Constraints */}
        <div className="p-6 rounded-2xl bg-surface-900 border border-slate-800 shadow-xl space-y-5">
          <div className="flex items-center gap-2 text-white font-bold text-sm border-b border-slate-800 pb-3">
            <Sliders className="w-4 h-4 text-razor-400" />
            <span>Execution & Confidence Thresholds</span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
            {/* Max Retry Attempts */}
            <div>
              <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Maximum Retry Attempts
              </label>
              <select
                value={settings.max_retry_attempts}
                onChange={(e) => setSettings({ ...settings, max_retry_attempts: Number(e.target.value) })}
                className="w-full px-3.5 py-2.5 rounded-xl bg-surface-950 border border-slate-800 text-white text-xs focus:outline-none focus:border-razor-500"
              >
                <option value="1">1 attempt (Strict)</option>
                <option value="2">2 attempts (Conservative)</option>
                <option value="3">3 attempts (Recommended standard)</option>
                <option value="4">4 attempts (Aggressive)</option>
                <option value="5">5 attempts (Maximum)</option>
              </select>
              <p className="text-[11px] text-slate-500 mt-1">
                Transactions exceeding this limit are aborted to eliminate useless gateway fees.
              </p>
            </div>

            {/* Min ML Confidence */}
            <div>
              <div className="flex items-center justify-between mb-1.5">
                <label className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
                  Minimum Confidence Threshold
                </label>
                <span className="font-mono text-xs font-bold text-razor-400">
                  {((settings.min_confidence_threshold || 0.65) * 100).toFixed(0)}%
                </span>
              </div>
              <input
                type="range"
                min="0.50"
                max="0.90"
                step="0.05"
                value={settings.min_confidence_threshold}
                onChange={(e) => setSettings({ ...settings, min_confidence_threshold: Number(e.target.value) })}
                className="w-full accent-razor-500 mt-2"
              />
              <p className="text-[11px] text-slate-500 mt-1">
                Decisions below this certainty level are flagged for human operator review.
              </p>
            </div>
          </div>
        </div>

        {/* Financial & Safety Gates */}
        <div className="p-6 rounded-2xl bg-surface-900 border border-slate-800 shadow-xl space-y-5">
          <div className="flex items-center gap-2 text-white font-bold text-sm border-b border-slate-800 pb-3">
            <Shield className="w-4 h-4 text-emerald-400" />
            <span>Financial Safeguards & Authorization Gates</span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
            {/* Human Approval Threshold */}
            <div>
              <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Human Approval Threshold (INR)
              </label>
              <div className="relative">
                <span className="absolute left-3.5 top-2.5 text-slate-400 font-mono text-xs">₹</span>
                <input
                  type="number"
                  step="5000"
                  min="5000"
                  value={settings.human_approval_threshold}
                  onChange={(e) => setSettings({ ...settings, human_approval_threshold: Number(e.target.value) })}
                  className="w-full pl-8 pr-4 py-2.5 rounded-xl bg-surface-950 border border-slate-800 text-white font-mono text-xs focus:outline-none focus:border-razor-500"
                />
              </div>
              <p className="text-[11px] text-slate-500 mt-1">
                Transactions at or above this value trigger a mandatory human approval policy check.
              </p>
            </div>

            {/* High Value Transaction Threshold */}
            <div>
              <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                High Value Outreach Threshold (INR)
              </label>
              <div className="relative">
                <span className="absolute left-3.5 top-2.5 text-slate-400 font-mono text-xs">₹</span>
                <input
                  type="number"
                  step="5000"
                  min="5000"
                  value={settings.high_value_transaction_threshold}
                  onChange={(e) => setSettings({ ...settings, high_value_transaction_threshold: Number(e.target.value) })}
                  className="w-full pl-8 pr-4 py-2.5 rounded-xl bg-surface-950 border border-slate-800 text-white font-mono text-xs focus:outline-none focus:border-razor-500"
                />
              </div>
              <p className="text-[11px] text-slate-500 mt-1">
                Threshold for priority VIP agent escalation routing.
              </p>
            </div>
          </div>

          {/* Auto-execution toggle */}
          <div className="pt-4 border-t border-slate-800/80 flex items-center justify-between">
            <div>
              <span className="text-xs font-bold text-white block">Automated Action Dispatch</span>
              <p className="text-[11px] text-slate-400">
                Automatically execute recovery strategies if policy checks pass with 100% compliance.
              </p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={settings.auto_execution_enabled}
                onChange={(e) => setSettings({ ...settings, auto_execution_enabled: e.target.checked })}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-slate-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-razor-500"></div>
            </label>
          </div>
        </div>

        {/* Save Bar */}
        <div className="flex items-center justify-end gap-3 pt-2">
          {savedSuccess && (
            <span className="text-xs text-emerald-400 font-semibold flex items-center gap-1 animate-in fade-in duration-200">
              <CheckCircle2 className="w-4 h-4" />
              <span>Policy settings saved successfully!</span>
            </span>
          )}

          <button
            type="submit"
            disabled={updateMutation.isPending}
            className="px-6 py-2.5 rounded-xl text-xs font-semibold bg-razor-500 hover:bg-razor-600 text-white flex items-center gap-2 shadow-lg shadow-razor-500/25 transition-all disabled:opacity-50"
          >
            <Save className="w-4 h-4" />
            <span>{updateMutation.isPending ? 'Saving...' : 'Save Policy Changes'}</span>
          </button>
        </div>
      </form>
    </div>
  );
};
