"use client";

import {
    LayoutDashboard,
    Network,
    Bot,
    BrainCircuit,
    Users,
    Activity,
    Server,
    Terminal,
    Database,
    Settings
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

const NAV_ITEMS = [
    {
        category: "Operations",
        items: [
            { name: "Dashboard", icon: LayoutDashboard, href: "/" },
            { name: "The Loom", icon: Network, href: "/loom" },
            { name: "Agents", icon: Bot, href: "/agents" },
        ]
    },
    {
        category: "Intelligence",
        items: [
            { name: "Cognitive Matrix", icon: BrainCircuit, href: "/matrix" },
            { name: "Knowledge Base", icon: Database, href: "/knowledge" },
        ]
    },
    {
        category: "Organization",
        items: [
            { name: "Team", icon: Users, href: "/team" },
            { name: "System Health", icon: Activity, href: "/health" },
            { name: "Infrastructure", icon: Server, href: "/infra" },
            { name: "Settings", icon: Settings, href: "/settings" },
        ]
    }
];

export function Sidebar() {
    const pathname = usePathname();
    // Mock Active Team
    const activeTeam = "Operations";

    return (
        <div className="w-64 h-full bg-zinc-50 border-r border-zinc-200 flex flex-col">
            <div className="p-6 border-b border-zinc-200">
                <h1 className="text-xl font-bold tracking-tighter text-emerald-600 flex items-center gap-2">
                    <Terminal className="w-6 h-6" />
                    CORTEX <span className="text-xs bg-emerald-100 text-emerald-700 px-1 py-0.5 rounded border border-emerald-200">v6</span>
                </h1>
                <div className="mt-2 text-xs font-semibold text-zinc-500 uppercase tracking-wider flex items-center gap-1">
                    <span className="w-2 h-2 rounded-full bg-emerald-500"></span>
                    {activeTeam}
                </div>
            </div>

            <nav className="flex-1 overflow-y-auto py-6 px-4 space-y-8">
                {NAV_ITEMS.map((section) => (
                    <div key={section.category}>
                        <h3 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-3 px-2">
                            {section.category}
                        </h3>
                        <div className="space-y-1">
                            {section.items.map((item) => {
                                const isActive = pathname === item.href;
                                return (
                                    <Link
                                        key={item.href}
                                        href={item.href}
                                        className={cn(
                                            "flex items-center gap-3 px-2 py-2 text-sm font-medium rounded-md transition-colors",
                                            isActive
                                                ? "bg-emerald-50 text-emerald-700 border border-emerald-100"
                                                : "text-zinc-600 hover:bg-zinc-100 hover:text-zinc-900"
                                        )}
                                    >
                                        <item.icon className="w-4 h-4" />
                                        {item.name}
                                    </Link>
                                );
                            })}
                        </div>
                    </div>
                ))}
            </nav>

            <div className="p-4 border-t border-zinc-200">
                <div className="flex items-center gap-3 p-2 rounded-md bg-white border border-zinc-200 shadow-sm">
                    <div className="w-8 h-8 rounded bg-gradient-to-br from-emerald-500 to-emerald-700 flex items-center justify-center text-white font-bold text-xs">
                        OP
                    </div>
                    <div className="overflow-hidden">
                        <p className="text-sm font-medium text-zinc-800 truncate">Operator</p>
                        <p className="text-xs text-zinc-500 truncate">Tier 0 Access</p>
                    </div>
                </div>
            </div>
        </div>
    );
}
