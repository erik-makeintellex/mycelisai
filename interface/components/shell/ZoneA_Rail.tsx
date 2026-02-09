import React from 'react';
import { Network, Cpu, LayoutGrid, Settings, Shield } from 'lucide-react';

export function ZoneA() {
    return (
        <div className="w-16 md:w-64 bg-zinc-950 text-white flex flex-col border-r border-zinc-800 z-50 flex-shrink-0 transition-all duration-300">
            {/* 1. Identity / Logo */}
            <div className="h-14 flex items-center justify-center md:justify-start md:px-4 border-b border-zinc-900">
                <div className="w-8 h-8 bg-indigo-600 rounded-lg flex items-center justify-center">
                    <Network className="w-5 h-5 text-white" />
                </div>
                <span className="hidden md:block ml-3 font-bold text-sm tracking-widest uppercase text-zinc-400">
                    Mycelis
                </span>
            </div>

            {/* 2. Navigation Items */}
            <div className="flex-1 flex flex-col py-4 gap-2 px-2">
                <NavItem icon={LayoutGrid} label="Mission Control" active />
                <NavItem icon={Cpu} label="System Status" />
                <NavItem icon={Shield} label="Governance" />
            </div>

            {/* 3. Footer / User / Settings */}
            <div className="p-2 border-t border-zinc-900">
                <NavItem icon={Settings} label="Settings" />
            </div>
        </div>
    );
}

function NavItem({ icon: Icon, label, active = false }: { icon: any; label: string; active?: boolean }) {
    return (
        <button
            className={`
        flex items-center justify-center md:justify-start w-full p-2 rounded-md transition-colors
        ${active ? 'bg-zinc-800 text-white' : 'text-zinc-500 hover:text-zinc-300 hover:bg-zinc-900'}
      `}
        >
            <Icon className="w-5 h-5" />
            <span className="hidden md:block ml-3 text-sm font-medium">{label}</span>
        </button>
    );
}
