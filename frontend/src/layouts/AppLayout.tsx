import React, { useState } from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import {
  LayoutDashboard,
  Coins,
  FlaskConical,
  Settings,
  PlusCircle,
  ShieldCheck,
  Zap,
  Activity,
  Layers,
} from 'lucide-react';
import { IngestEventModal } from '../components/IngestEventModal';

export const AppLayout: React.FC = () => {
  const [isIngestOpen, setIsIngestOpen] = useState(false);

  const navItems = [
    { name: 'Command Center', path: '/dashboard', icon: LayoutDashboard },
    { name: 'Opportunities', path: '/opportunities', icon: Coins },
    { name: 'Simulation Lab', path: '/simulation', icon: FlaskConical },
    { name: 'Policy Settings', path: '/settings', icon: Settings },
  ];

  return (
    <div className="flex h-screen overflow-hidden bg-[#060e1a] text-slate-100 font-sans">
      {/* Sidebar */}
      <aside className="w-64 flex-shrink-0 bg-[#091424] border-r border-slate-800/80 flex flex-col justify-between z-20">
        <div>
          {/* Brand Header */}
          <div className="p-5 border-b border-slate-800/80 flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-razor-600 to-razor-400 flex items-center justify-center shadow-lg shadow-razor-500/20 text-white font-black text-lg">
              R
            </div>
            <div>
              <div className="flex items-center gap-1.5">
                <span className="font-bold text-sm text-white tracking-tight">Razorpay</span>
                <span className="text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded bg-razor-500/20 text-razor-400 border border-razor-500/30">
                  RRI
                </span>
              </div>
              <p className="text-[11px] text-slate-400 font-medium truncate">Recovery Intelligence</p>
            </div>
          </div>

          {/* Navigation Links */}
          <nav className="p-3 space-y-1">
            <div className="px-3 py-2 text-[10px] font-bold uppercase tracking-wider text-slate-500">
              Platform Modules
            </div>
            {navItems.map((item) => (
              <NavLink
                key={item.path}
                to={item.path}
                className={({ isActive }) =>
                  `flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-xs font-semibold transition-all duration-150 ${
                    isActive
                      ? 'bg-razor-500/15 text-razor-400 border border-razor-500/30 shadow-sm'
                      : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50'
                  }`
                }
              >
                <item.icon className="w-4 h-4" />
                <span>{item.name}</span>
              </NavLink>
            ))}
          </nav>
        </div>

        {/* Sidebar Footer */}
        <div className="p-4 border-t border-slate-800/80 space-y-3">
          <button
            onClick={() => setIsIngestOpen(true)}
            className="w-full px-3.5 py-2.5 rounded-xl text-xs font-semibold bg-gradient-to-r from-razor-500 to-blue-600 hover:from-razor-600 hover:to-blue-700 text-white flex items-center justify-center gap-2 shadow-lg shadow-razor-500/20 transition-all active:scale-95"
          >
            <PlusCircle className="w-4 h-4" />
            <span>Simulate Failure</span>
          </button>

          <div className="p-3 rounded-xl bg-slate-900/80 border border-slate-800 text-[11px] text-slate-400 space-y-1">
            <div className="flex items-center gap-1.5 text-slate-300 font-semibold">
              <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
              <span>Event Engine Live</span>
            </div>
            <p className="text-[10px] text-slate-500">
              Closed-Loop: Detect &rarr; Enrich &rarr; Predict &rarr; Select &rarr; Policy &rarr; Execute
            </p>
          </div>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Top App Bar */}
        <header className="h-16 flex-shrink-0 bg-[#091424]/80 backdrop-blur-md border-b border-slate-800/80 px-8 flex items-center justify-between z-10">
          <div className="flex items-center gap-3">
            <span className="px-2.5 py-1 rounded-md text-[11px] font-bold bg-slate-800 text-slate-300 border border-slate-700 flex items-center gap-1.5">
              <Zap className="w-3.5 h-3.5 text-razor-400" />
              Razorpay — AI Revenue Recovery
            </span>
            <span className="text-xs text-slate-400 hidden md:inline">
              &bull; &ldquo;Don&apos;t retry every failure. Recover the revenue worth recovering.&rdquo;
            </span>
          </div>

          <div className="flex items-center gap-3">
            
            <div className="px-3 py-1.5 rounded-lg bg-surface-900 border border-slate-800 flex items-center gap-2 text-xs text-slate-300">
              <ShieldCheck className="w-3.5 h-3.5 text-razor-400" />
              <span>Policy Engine: Active</span>
            </div>
          </div>
        </header>

        {/* Page Body */}
        <main className="flex-1 overflow-y-auto p-8 bg-[#060e1a]">
          <div className="max-w-7xl mx-auto">
            <Outlet />
          </div>
        </main>
      </div>

      {/* Quick Ingest Modal */}
      <IngestEventModal isOpen={isIngestOpen} onClose={() => setIsIngestOpen(false)} />
    </div>
  );
};
