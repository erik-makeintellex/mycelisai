'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { clsx } from 'clsx';

export default function Navbar() {
    const pathname = usePathname();

    const isActive = (path: string) => pathname === path;

    const navLinks = [
        { href: '/', label: 'Dashboard' },
        { href: '/agents', label: 'Agents' },
        { href: '/models', label: 'Models' },
        { href: '/teams', label: 'Teams' },
        { href: '/config', label: 'Configuration' },
        { href: '/channels', label: 'Channels' },
        { href: '/monitor', label: 'Monitor' },
    ];

    return (
        <nav className="border-b border-[--border-subtle] bg-[--bg-glass] backdrop-blur-xl sticky top-0 z-50">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="flex items-center justify-between h-16">
                    <div className="flex items-center">
                        <Link href="/" className="text-xl font-bold tracking-tight">
                            <span className="text-[--text-primary]">Mycelis</span>
                            <span className="bg-gradient-to-r from-[--accent-info] to-purple-500 bg-clip-text text-transparent">AI</span>
                        </Link>
                        <div className="hidden md:block ml-10">
                            <div className="flex items-baseline space-x-2">
                                {navLinks.map((link) => (
                                    <Link
                                        key={link.href}
                                        href={link.href}
                                        className={clsx(
                                            'px-3 py-2 rounded-lg text-sm font-medium transition-all duration-300 relative',
                                            isActive(link.href)
                                                ? 'text-[--text-primary] bg-[--accent-info]/10'
                                                : 'text-[--text-secondary] hover:text-[--text-primary]'
                                        )}
                                    >
                                        {link.label}
                                        {isActive(link.href) && (
                                            <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-[--accent-info] shadow-[0_0_8px_var(--accent-info)]" />
                                        )}
                                    </Link>
                                ))}
                            </div>
                        </div>
                    </div>
                    <div className="flex items-center gap-3 group relative">
                        <div className="h-2 w-2 rounded-full bg-[--accent-success] animate-pulse shadow-[0_0_6px_var(--accent-success)]"></div>
                        <span className="text-xs text-[--text-secondary] font-mono tracking-wide">SYSTEM ONLINE</span>
                        <div className="absolute hidden group-hover:block top-full right-0 mt-2 px-3 py-2 bg-[--bg-secondary] border border-[--border-subtle] rounded-lg text-xs text-[--text-secondary] whitespace-nowrap">
                            All Systems Operational
                        </div>
                    </div>
                </div>
            </div>
        </nav>
    );
}
