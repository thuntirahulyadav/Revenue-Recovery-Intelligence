import React from 'react';
import { LucideIcon } from 'lucide-react';

interface KPICardProps {
  title: string;
  value: string;
  subtitle?: string;
  icon: LucideIcon;
  trend?: {
    value: string;
    isPositive: boolean;
  };
  variant?: 'default' | 'emerald' | 'blue' | 'purple' | 'amber';
}

export const KPICard: React.FC<KPICardProps> = ({
  title,
  value,
  subtitle,
  icon: Icon,
  trend,
  variant = 'default',
}) => {
  const getGlow = () => {
    switch (variant) {
      case 'emerald':
        return 'border-emerald-500/30 bg-gradient-to-br from-emerald-950/20 via-surface-900 to-surface-900 text-emerald-400';
      case 'blue':
        return 'border-razor-500/30 bg-gradient-to-br from-razor-950/40 via-surface-900 to-surface-900 text-razor-400';
      case 'purple':
        return 'border-purple-500/30 bg-gradient-to-br from-purple-950/20 via-surface-900 to-surface-900 text-purple-400';
      case 'amber':
        return 'border-amber-500/30 bg-gradient-to-br from-amber-950/20 via-surface-900 to-surface-900 text-amber-400';
      default:
        return 'border-slate-800 bg-surface-900 text-slate-300';
    }
  };

  return (
    <div className={`p-5 rounded-xl border shadow-lg backdrop-blur-sm transition-all duration-200 hover:border-slate-700 ${getGlow()}`}>
      <div className="flex items-center justify-between">
        <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">{title}</span>
        <div className="p-2 rounded-lg bg-surface-950/60 border border-slate-800/80">
          <Icon className="w-5 h-5" />
        </div>
      </div>
      <div className="mt-3 flex items-baseline gap-2">
        <span className="text-2xl font-bold tracking-tight text-white font-mono">{value}</span>
        {trend && (
          <span
            className={`text-xs font-semibold px-1.5 py-0.5 rounded ${
              trend.isPositive
                ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
            }`}
          >
            {trend.value}
          </span>
        )}
      </div>
      {subtitle && <p className="mt-1 text-xs text-slate-400">{subtitle}</p>}
    </div>
  );
};
