"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import {
    Terminal, // Command
    Network, // Orchestration
    ShieldCheck, // Approvals
    Activity, // Telemetry
    Database, // Registry
    Settings,
    ChevronDown,
    User
} from "lucide-react"
import React from "react"

export function CommandRail() {
    const pathname = usePathname()

    const isActive = (path: string) => pathname === path

    return (
        <aside className="w-16 lg:w-64 h-screen bg-white border-r border-zinc-200 flex flex-col z-40 transition-all duration-300">
            {/* 1. Brand Mark */}
            <div className="h-16 flex items-center px-4 border-b border-zinc-200 overflow-hidden whitespace-nowrap">
                <div className="flex items-center gap-2 font-bold text-slate-900 text-lg tracking-tight">
                    <div className="min-w-6 w-6 h-6 bg-slate-900 rounded-sm" />
                    <span className="hidden lg:block">Mycelis</span>
                </div>
            </div>

            {/* 2. Workspace Context */}
            <div className="p-3 border-b border-zinc-200 hidden lg:block">
                <button className="w-full h-10 flex items-center gap-3 px-2 bg-zinc-50 hover:bg-zinc-100 border border-zinc-200 rounded-md transition-colors">
                    <div className="w-6 h-6 bg-indigo-600 rounded flex items-center justify-center text-[10px] text-white font-bold">
                        SW
                    </div>
                    <div className="flex flex-col items-start flex-1 overflow-hidden">
                        <span className="text-xs font-semibold text-slate-900 truncate">System Workspace</span>
                    </div>
                    <ChevronDown size={14} className="text-zinc-400" />
                </button>
            </div>

            {/* 3. Primary Navigation */}
            <nav className="flex-1 py-4 px-2 lg:px-3 space-y-1">
                <NavItem
                    href="/"
                    icon={<Terminal size={20} />}
                    label="Command"
                    active={isActive("/")}
                />
                <NavItem
                    href="/orchestration"
                    icon={<Network size={20} />}
                    label="Orchestration"
                    active={isActive("/orchestration")}
                />
                <NavItem
                    href="/approvals"
                    icon={<ShieldCheck size={20} />}
                    label="Approvals"
                    active={isActive("/approvals")}
                    badge="2"
                />
                <NavItem
                    href="/telemetry"
                    icon={<Activity size={20} />}
                    label="Telemetry"
                    active={isActive("/telemetry")}
                />
                <NavItem
                    href="/registry"
                    icon={<Database size={20} />}
                    label="Registry"
                    active={isActive("/registry")}
                />
            </nav>

            {/* 4. Secondary / Settings */}
            <div className="px-2 lg:px-3 py-4 border-t border-zinc-200 space-y-1">
                <NavItem
                    href="/settings"
                    icon={<Settings size={20} />}
                    label="Settings"
                    active={isActive("/settings")}
                />
            </div>

            {/* 5. User Profile */}
            <div className="h-14 mt-auto border-t border-zinc-200 px-4 flex items-center justify-center lg:justify-between">
                <div className="flex items-center gap-0 lg:gap-2">
                    <div className="relative">
                        <div className="w-8 h-8 bg-zinc-100 rounded-full flex items-center justify-center border border-zinc-200">
                            <User size={14} className="text-zinc-500" />
                        </div>
                        <div className="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 bg-emerald-500 border-2 border-white rounded-full" />
                    </div>
                    <div className="hidden lg:flex flex-col">
                        <span className="text-xs font-medium text-slate-900">Operator</span>
                        <span className="text-[10px] text-zinc-500">Online</span>
                    </div>
                </div>
            </div>
        </aside>
    )
}

function NavItem({
    href,
    icon,
    label,
    active,
    badge
}: {
    href: string,
    icon: React.ReactNode,
    label: string,
    active: boolean,
    badge?: string
}) {
    return (
        <Link
            href={href}
            title={label}
            className={`
        flex items-center justify-center lg:justify-start gap-3 px-3 py-2 rounded-md text-sm font-medium transition-all
        ${active
                    ? "bg-zinc-100 text-sky-700 border-l-0 lg:border-l-4 border-sky-600 shadow-sm"
                    : "text-slate-500 hover:text-slate-900 hover:bg-zinc-50"
                }
      `}
        >
            {icon}
            <span className="hidden lg:inline flex-1">{label}</span>
            {badge && (
                <span className="hidden lg:inline px-1.5 py-0.5 text-[10px] bg-rose-100 text-rose-700 rounded-full font-bold shadow-sm">
                    {badge}
                </span>
            )}
            {/* Mobile Badge Dot */}
            {badge && (
                <span className="lg:hidden absolute top-0 right-0 w-2 h-2 bg-rose-500 rounded-full"></span>
            )}
        </Link>
    )
}
