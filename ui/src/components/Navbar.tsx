'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

export default function Navbar() {
    const pathname = usePathname();

    const isActive = (path: string) => pathname === path;

    return (
        <nav className="border-b border-zinc-700 bg-zinc-900/95 backdrop-blur-md sticky top-0 z-50">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="flex items-center justify-between h-16">
                    <div className="flex items-center">
                        <Link href="/" className="text-xl font-bold text-zinc-100 tracking-tight">
                            Mycelis<span className="text-zinc-400">AI</span>
                        </Link>
                        <div className="hidden md:block ml-10">
                            <div className="flex items-baseline space-x-4">
                                <Link
                                    href="/"
                                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive('/')
                                        ? 'bg-zinc-800 text-white'
                                        : 'text-zinc-400 hover:text-white hover:bg-zinc-800/50'
                                        }`}
                                >
                                    Dashboard
                                </Link>
                                <Link
                                    href="/agents"
                                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive('/agents')
                                        ? 'bg-zinc-800 text-white'
                                        : 'text-zinc-400 hover:text-white hover:bg-zinc-800/50'
                                        }`}
                                >
                                    Agents
                                </Link>
                                <Link
                                    href="/models"
                                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive('/models')
                                        ? 'bg-zinc-800 text-white'
                                        : 'text-zinc-400 hover:text-white hover:bg-zinc-800/50'
                                        }`}
                                >
                                    Models
                                </Link>
                                <Link
                                    href="/teams"
                                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive('/teams')
                                        ? 'bg-zinc-800 text-white'
                                        : 'text-zinc-400 hover:text-white hover:bg-zinc-800/50'
                                        }`}
                                >
                                    Teams
                                </Link>
                                <Link
                                    href="/config"
                                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive('/config')
                                        ? 'bg-zinc-800 text-white'
                                        : 'text-zinc-400 hover:text-white hover:bg-zinc-800/50'
                                        }`}
                                >
                                    Configuration
                                </Link>
                                <Link
                                    href="/monitor"
                                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${isActive('/monitor')
                                        ? 'bg-zinc-800 text-white'
                                        : 'text-zinc-400 hover:text-white hover:bg-zinc-800/50'
                                        }`}
                                >
                                    Monitor
                                </Link>
                            </div>
                        </div>
                    </div>
                    <div className="flex items-center gap-4">
                        <div className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse"></div>
                        <span className="text-xs text-zinc-500 font-mono">SYSTEM ONLINE</span>
                    </div>
                </div>
            </div>
        </nav>
    );
}
