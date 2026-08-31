import React, { useState } from 'react';
import { PlusCircle, Activity, Sparkles, CheckCircle2, ArrowRight } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import { FailureReason, PaymentMethod } from '../types';

interface IngestEventModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess?: () => void;
}

export const IngestEventModal: React.FC<IngestEventModalProps> = ({ isOpen, onClose, onSuccess }) => {
  const navigate = useNavigate();
  const [amount, setAmount] = useState<number>(4500);
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('card');
  const [failureReason, setFailureReason] = useState<FailureReason>('BANK_TIMEOUT');
  const [attemptCount, setAttemptCount] = useState<number>(1);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [result, setResult] = useState<any>(null);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const res = await api.ingestPaymentFailed({
        amount,
        payment_method: paymentMethod,
        failure_reason: failureReason,
        attempt_count: attemptCount,
      });
      setResult(res);
      if (onSuccess) onSuccess();
    } catch (err: any) {
      alert(`Ingestion failed: ${err.message}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleViewAnalysis = () => {
    if (result?.payment?.id) {
      onClose();
      navigate(`/payments/${result.payment.id}`);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="w-full max-w-lg bg-surface-900 border border-slate-700 rounded-2xl shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="p-6 border-b border-slate-800 bg-surface-950/60 flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-lg bg-razor-500/10 border border-razor-500/30 text-razor-400">
              <PlusCircle className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-lg font-bold text-white">Simulate Failed Payment Event</h3>
              <p className="text-xs text-slate-400">Triggers real-time ML enrichment & strategy decision pipeline</p>
            </div>
          </div>
          <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors">✕</button>
        </div>

        {/* Content */}
        {!result ? (
          <form onSubmit={handleSubmit} className="p-6 space-y-4">
            <div>
              <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Transaction Amount (INR)
              </label>
              <div className="relative">
                <span className="absolute left-3.5 top-2.5 text-slate-400 font-mono text-sm">₹</span>
                <input
                  type="number"
                  min="100"
                  step="50"
                  value={amount}
                  onChange={(e) => setAmount(Number(e.target.value))}
                  className="w-full pl-8 pr-4 py-2.5 rounded-lg bg-surface-950 border border-slate-700 text-white font-mono text-sm focus:outline-none focus:border-razor-500 transition-colors"
                  required
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  Payment Method
                </label>
                <select
                  value={paymentMethod}
                  onChange={(e) => setPaymentMethod(e.target.value as PaymentMethod)}
                  className="w-full px-3 py-2.5 rounded-lg bg-surface-950 border border-slate-700 text-white text-sm focus:outline-none focus:border-razor-500"
                >
                  <option value="card">Credit / Debit Card</option>
                  <option value="upi">UPI Instant</option>
                  <option value="netbanking">NetBanking</option>
                  <option value="wallet">Digital Wallet</option>
                  <option value="emi">Card EMI</option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  Attempt Number
                </label>
                <select
                  value={attemptCount}
                  onChange={(e) => setAttemptCount(Number(e.target.value))}
                  className="w-full px-3 py-2.5 rounded-lg bg-surface-950 border border-slate-700 text-white text-sm focus:outline-none focus:border-razor-500"
                >
                  <option value="1">Attempt 1 (Initial Failure)</option>
                  <option value="2">Attempt 2 (First Retry)</option>
                  <option value="3">Attempt 3 (Second Retry)</option>
                  <option value="4">Attempt 4 (High Retries)</option>
                </select>
              </div>
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Failure Category
              </label>
              <select
                value={failureReason}
                onChange={(e) => setFailureReason(e.target.value as FailureReason)}
                className="w-full px-3 py-2.5 rounded-lg bg-surface-950 border border-slate-700 text-white text-sm focus:outline-none focus:border-razor-500"
              >
                <option value="BANK_TIMEOUT">BANK_TIMEOUT (Transient bank core gateway delay)</option>
                <option value="NETWORK_ERROR">NETWORK_ERROR (Network packet drop / glitch)</option>
                <option value="INSUFFICIENT_FUNDS">INSUFFICIENT_FUNDS (Customer account low balance)</option>
                <option value="CARD_EXPIRED">CARD_EXPIRED (Expired card instrument)</option>
                <option value="PAYMENT_METHOD_FAILURE">PAYMENT_METHOD_FAILURE (Issuer outage)</option>
                <option value="CUSTOMER_ABANDONMENT">CUSTOMER_ABANDONMENT (3DS OTP drop-off)</option>
                <option value="TECHNICAL_ERROR">TECHNICAL_ERROR (Upstream 5xx exception)</option>
              </select>
            </div>

            <div className="pt-3 border-t border-slate-800 flex items-center justify-end gap-3">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 rounded-lg text-xs font-semibold text-slate-300 hover:bg-slate-800 transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isSubmitting}
                className="px-5 py-2.5 rounded-lg text-xs font-semibold bg-razor-500 hover:bg-razor-600 text-white flex items-center gap-2 shadow-lg shadow-razor-500/25 transition-all disabled:opacity-50"
              >
                <Activity className="w-4 h-4" />
                {isSubmitting ? 'Simulating Pipeline...' : 'Emit payment.failed Event'}
              </button>
            </div>
          </form>
        ) : (
          <div className="p-6 space-y-4 text-center animate-in zoom-in-95 duration-200">
            <div className="inline-flex p-3 rounded-full bg-emerald-500/10 border border-emerald-500/30 text-emerald-400">
              <CheckCircle2 className="w-10 h-10" />
            </div>

            <div>
              <h4 className="text-lg font-bold text-white">Event Processed Through AI Pipeline</h4>
              <p className="text-xs text-slate-400 mt-1">
                Enriched customer history, evaluated XGBoost model, selected optimal strategy, and validated policy.
              </p>
            </div>

            <div className="p-4 rounded-xl bg-surface-950 border border-slate-800 text-left space-y-2">
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-400">Recommended Strategy:</span>
                <span className="font-bold text-razor-400">{result.decision?.strategy}</span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-400">Recovery Probability:</span>
                <span className="font-mono font-bold text-emerald-400">
                  {((result.prediction?.recovery_probability || 0) * 100).toFixed(1)}%
                </span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-400">Expected Net Value:</span>
                <span className="font-mono font-bold text-white">
                  ₹{result.decision?.expected_net_value?.toFixed(2)}
                </span>
              </div>
            </div>

            <div className="pt-2 flex items-center justify-end gap-3">
              <button
                onClick={onClose}
                className="px-4 py-2 rounded-lg text-xs font-semibold text-slate-300 hover:bg-slate-800"
              >
                Close
              </button>
              <button
                onClick={handleViewAnalysis}
                className="px-5 py-2.5 rounded-lg text-xs font-semibold bg-razor-500 hover:bg-razor-600 text-white flex items-center gap-1.5 shadow-lg shadow-razor-500/25"
              >
                <span>View Full AI Breakdown</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
