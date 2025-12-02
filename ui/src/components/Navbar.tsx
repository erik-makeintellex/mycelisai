'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

export default function Navbar() {
    const pathname = usePathname();

    const isActive = (path: string) => pathname === path;

    return (
        <nav className="border-b border-[--border-light] bg-[--bg-primary]/95 backdrop-blur-md sticky top-0 z-50">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="flex items-center justify-between h-16">
                    <div className="flex items-center">
                        <Link href="/" className="text-xl font-bold text-[--text-primary] tracking-tight">
                            Mycelis<span className="text-[--text-secondary]">AI</span>
                        </Link>
                        <div className="hidden md:block ml-10">
                            <div className="flex items-baseline space-x-4">
                                <Link
                                    href="/"
                                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive('/')
                                        ? 'bg-[--bg-panel] text-[--text-primary]'
                                        : 'text-[--text-secondary] hover:text-[--text-primary] hover:bg-[--bg-panel]/50'
                                        }`}
                                >
                                    Dashboard
                                </Link>
                                <Link
                                    href="/agents"
                                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive('/agents')
                                        ? 'bg-[--bg-panel] text-[--text-primary]'
                                        : 'text-[--text-secondary] hover:text-[--text-primary] hover:bg-[--bg-panel]/50'
                                        }`}
                                >
                                    Agents
                                </Link>
                                <Link
                                    href="/models"
                                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive('/models')
                                        ? 'bg-[--bg-panel] text-[--text-primary]'
                                        : 'text-[--text-secondary] hover:text-[--text-primary] hover:bg-[--bg-panel]/50'
                                        }`}
                                >
                                    Models
                                </Link>
                                <Link
                                    href="/teams"
                                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive('/teams')
                                        ? 'bg-[--bg-panel] text-[--text-primary]'
                                        : 'text-[--text-secondary] hover:text-[--text-primary] hover:bg-[--bg-panel]/50'
                                        }`}
                                >
                                    Teams
                                </Link>
                                <Link
                                    href="/config"
                                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive('/config')
                                        ? 'bg-[--bg-panel] text-[--text-primary]'
                                        : 'text-[--text-secondary] hover:text-[--text-primary] hover:bg-[--bg-panel]/50'
                                        }`}
                                >
                                    Configuration
                                </Link>
                                <Link
                                    href="/channels"
                                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive('/channels')
                                        ? 'bg-[--bg-panel] text-[--text-primary]'
                                        : 'text-[--text-secondary] hover:text-[--text-primary] hover:bg-[--bg-panel]/50'
                                        }`}
                                >
                                    Channels
                                </Link>
                                <Link
                                    href="/monitor"
                                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive('/monitor')
                                        ? 'bg-[--bg-panel] text-[--text-primary]'
                                        : 'text-[--text-secondary] hover:text-[--text-primary] hover:bg-[--bg-panel]/50'
                                        }`}
                                >
                                    Monitor
                                </Link>
                            </div>
                        </div>
                    </div>
                    <div className="flex items-center gap-4">
                        <div className="h-2 w-2 rounded-full bg-[--accent-success] animate-pulse"></div>
                        <span className="text-xs text-[--text-muted] font-mono">SYSTEM ONLINE</span>
                    </div>
                </div>
            </div>
        </nav>
    );
}

