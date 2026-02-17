"use client";

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Network, Cpu, LayoutGrid, Settings, Shield, Share2, BrainCircuit, Cable, Store, BookOpen, Brain, BarChart3, Users, Home } from 'lucide-react';

export function ZoneA() {
    return (
        <div className="w-16 md:w-64 bg-cortex-surface text-cortex-text-main flex flex-col border-r border-cortex-border z-50 flex-shrink-0 transition-all duration-300">
            {/* 1. Identity / Logo */}
            <div className="h-14 flex items-center justify-center md:justify-start md:px-4 border-b border-cortex-border">
                <div className="w-8 h-8 bg-cortex-primary rounded-lg flex items-center justify-center shadow-[0_4px_14px_0_rgba(115,103,240,0.39)]">
                    <Network className="w-5 h-5 text-white" />
                </div>
                <span className="hidden md:block ml-3 font-bold text-sm tracking-widest uppercase text-cortex-text-muted">
                    Mycelis
                </span>
            </div>

            {/* 2. Navigation Items */}
            <div className="flex-1 flex flex-col py-4 gap-1 px-2">
                <NavItem href="/" icon={Home} label="Product Home" />
                <NavItem href="/dashboard" icon={BarChart3} label="Dashboard" />
                <NavItem href="/architect" icon={Share2} label="Swarm Architect" />
                <NavItem href="/matrix" icon={BrainCircuit} label="Cognitive Matrix" />
                <NavItem href="/wiring" icon={Cable} label="Neural Wiring" />
                <NavItem href="/teams" icon={Users} label="Team Management" />
                <NavItem href="/catalogue" icon={BookOpen} label="Agent Catalogue" />
                <NavItem href="/marketplace" icon={Store} label="Skills Market" />
                <NavItem href="/memory" icon={Brain} label="Memory" />
                <NavItem href="/telemetry" icon={Cpu} label="System Status" />
                <NavItem href="/approvals" icon={Shield} label="Governance" />
            </div>

            {/* 3. Footer / Settings */}
            <div className="p-2 border-t border-cortex-border">
                <NavItem href="/settings" icon={Settings} label="Settings" />
            </div>
        </div>
    );
}

function NavItem({ icon: Icon, label, href }: { icon: any; label: string; href: string }) {
    const pathname = usePathname();
    const isActive = pathname === href;

    return (
        <Link
            href={href}
            className={`
                flex items-center justify-center md:justify-start w-full p-2.5 rounded-lg transition-all duration-200
                ${isActive
                    ? 'bg-cortex-primary text-white shadow-[0_4px_14px_0_rgba(115,103,240,0.39)]'
                    : 'text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg'
                }
            `}
        >
            <Icon className="w-5 h-5 flex-shrink-0" />
            <span className="hidden md:block ml-3 text-sm font-medium">{label}</span>
        </Link>
    );
}
