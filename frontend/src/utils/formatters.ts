export function formatINR(amount: number): string {
  if (isNaN(amount) || amount === null || amount === undefined) return '₹0.00';
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    maximumFractionDigits: 2,
  }).format(amount);
}

export function formatCompactINR(amount: number): string {
  if (isNaN(amount) || amount === null || amount === undefined) return '₹0';
  if (Math.abs(amount) >= 10000000) {
    return `₹${(amount / 10000000).toFixed(2)} Cr`;
  }
  if (Math.abs(amount) >= 100000) {
    return `₹${(amount / 100000).toFixed(2)} L`;
  }
  if (Math.abs(amount) >= 1000) {
    return `₹${(amount / 1000).toFixed(1)}k`;
  }
  return `₹${amount.toFixed(0)}`;
}

export function formatPercent(value: number): string {
  if (isNaN(value) || value === null || value === undefined) return '0.0%';
  const num = value > 1 ? value : value * 100;
  return `${num.toFixed(1)}%`;
}

export function formatDate(dateString: string): string {
  if (!dateString) return '—';
  try {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-IN', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return dateString;
  }
}

export function getStrategyBadgeClass(strategy: string): string {
  switch (strategy) {
    case 'RETRY_NOW':
      return 'bg-blue-500/10 text-blue-400 border-blue-500/30';
    case 'RETRY_LATER':
      return 'bg-indigo-500/10 text-indigo-400 border-indigo-500/30';
    case 'SWITCH_PAYMENT_METHOD':
      return 'bg-purple-500/10 text-purple-400 border-purple-500/30';
    case 'SEND_PAYMENT_LINK':
      return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30';
    case 'ESCALATE_TO_HUMAN':
      return 'bg-amber-500/10 text-amber-400 border-amber-500/30';
    case 'STOP_RECOVERY':
      return 'bg-rose-500/10 text-rose-400 border-rose-500/30';
    default:
      return 'bg-slate-500/10 text-slate-400 border-slate-500/30';
  }
}

export function getFailureReasonBadgeClass(reason: string): string {
  switch (reason) {
    case 'BANK_TIMEOUT':
      return 'bg-cyan-500/10 text-cyan-400 border-cyan-500/30';
    case 'NETWORK_ERROR':
      return 'bg-sky-500/10 text-sky-400 border-sky-500/30';
    case 'INSUFFICIENT_FUNDS':
      return 'bg-amber-500/10 text-amber-400 border-amber-500/30';
    case 'CARD_EXPIRED':
      return 'bg-rose-500/10 text-rose-400 border-rose-500/30';
    case 'PAYMENT_METHOD_FAILURE':
      return 'bg-violet-500/10 text-violet-400 border-violet-500/30';
    case 'CUSTOMER_ABANDONMENT':
      return 'bg-orange-500/10 text-orange-400 border-orange-500/30';
    case 'TECHNICAL_ERROR':
      return 'bg-red-500/10 text-red-400 border-red-500/30';
    default:
      return 'bg-slate-500/10 text-slate-400 border-slate-500/30';
  }
}
