"use client";

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Network, Settings, Home, Workflow, FolderCog, Brain, Activity, Eye, EyeOff } from 'lucide-react';
import { useCortexStore } from '@/store/useCortexStore';

export function ZoneA() {
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const toggleAdvancedMode = useCortexStore((s) => s.toggleAdvancedMode);

    return (
        <div className="w-16 md:w-64 bg-cortex-surface text-cortex-text-main flex flex-col border-r border-cortex-border z-50 flex-shrink-0 transition-all duration-300">
            {/* 1. Identity / Logo → Home */}
            <Link href="/" className="h-14 flex items-center justify-center md:justify-start md:px-4 border-b border-cortex-border hover:bg-cortex-bg/50 transition-colors">
                <div className="w-8 h-8 bg-cortex-primary rounded-lg flex items-center justify-center shadow-[0_4px_14px_0_rgba(6,182,212,0.39)]">
                    <Network className="w-5 h-5 text-white" />
                </div>
                <span className="hidden md:block ml-3 font-bold text-sm tracking-widest uppercase text-cortex-text-muted">
                    Mycelis
                </span>
            </Link>

            {/* 2. Workflow-First Navigation (V7) */}
            <div className="flex-1 flex flex-col py-4 gap-1 px-2">
                <NavItem href="/dashboard" icon={Home} label="Workspace" />
                <NavItem href="/automations" icon={Workflow} label="Automations" />
                <NavItem href="/resources" icon={FolderCog} label="Resources" />
                <NavItem href="/memory" icon={Brain} label="Memory" />
                {advancedMode && (
                    <NavItem href="/system" icon={Activity} label="System" />
                )}
            </div>

            {/* 3. Footer: Advanced Toggle + Settings */}
            <div className="p-2 border-t border-cortex-border space-y-1">
                <button
                    onClick={toggleAdvancedMode}
                    className="flex items-center justify-center md:justify-start w-full p-2.5 rounded-lg transition-all duration-200 text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg"
                    title={advancedMode ? 'Hide advanced panels' : 'Show advanced panels'}
                >
                    {advancedMode ? (
                        <EyeOff className="w-5 h-5 flex-shrink-0" />
                    ) : (
                        <Eye className="w-5 h-5 flex-shrink-0" />
                    )}
                    <span className="hidden md:block ml-3 text-sm font-medium">
                        {advancedMode ? 'Advanced: On' : 'Advanced: Off'}
                    </span>
                </button>
                <NavItem href="/settings" icon={Settings} label="Settings" />
            </div>
        </div>
    );
}

function NavItem({ icon: Icon, label, href }: { icon: any; label: string; href: string }) {
    const pathname = usePathname();
    const isActive = pathname === href || pathname.startsWith(href + '/');

    return (
        <Link
            href={href}
            className={`
                flex items-center justify-center md:justify-start w-full p-2.5 rounded-lg transition-all duration-200
                ${isActive
                    ? 'bg-cortex-primary text-cortex-bg shadow-[0_4px_14px_0_rgba(6,182,212,0.39)]'
                    : 'text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg'
                }
            `}
        >
            <Icon className="w-5 h-5 flex-shrink-0" />
            <span className="hidden md:block ml-3 text-sm font-medium">{label}</span>
        </Link>
    );
}
