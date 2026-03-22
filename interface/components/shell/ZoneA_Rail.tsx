"use client";

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Network, Settings, Home, Workflow, FolderCog, Brain, Activity, Eye, EyeOff, BookOpen, Building2 } from 'lucide-react';
import { useCortexStore } from '@/store/useCortexStore';

export function ZoneA() {
    const pathname = usePathname();
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const toggleAdvancedMode = useCortexStore((s) => s.toggleAdvancedMode);
    const [isHydrated, setIsHydrated] = useState(false);
    const [lastOrganization, setLastOrganization] = useState<{ id: string; name: string } | null>(null);

    useEffect(() => {
        const syncLastOrganization = () => {
            if (typeof window === 'undefined') {
                return;
            }
            const id = window.localStorage.getItem('mycelis-last-organization-id');
            const name = window.localStorage.getItem('mycelis-last-organization-name');
            setLastOrganization(id ? { id, name: name || 'Current Organization' } : null);
        };

        const handleLastOrganizationChanged = (event: Event) => {
            const detail = (event as CustomEvent<{ id: string; name: string }>).detail;
            if (!detail?.id) {
                syncLastOrganization();
                return;
            }
            setLastOrganization({ id: detail.id, name: detail.name || 'Current Organization' });
        };

        setIsHydrated(true);
        syncLastOrganization();
        window.addEventListener('mycelis:last-organization-changed', handleLastOrganizationChanged as EventListener);
        return () => {
            window.removeEventListener('mycelis:last-organization-changed', handleLastOrganizationChanged as EventListener);
        };
    }, [pathname]);

    const effectiveAdvancedMode = isHydrated ? advancedMode : false;
    const primaryNav = [
        { href: '/dashboard', icon: Home, label: 'AI Organization' },
        ...(lastOrganization ? [{ href: `/organizations/${lastOrganization.id}`, icon: Building2, label: 'Current Organization', title: lastOrganization.name }] : []),
        { href: '/automations', icon: Workflow, label: 'Automations' },
        { href: '/docs', icon: BookOpen, label: 'Docs' },
    ];
    const advancedNav = [
        { href: '/resources', icon: FolderCog, label: 'Resources' },
        { href: '/memory', icon: Brain, label: 'Memory' },
        { href: '/system', icon: Activity, label: 'System' },
    ];

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

            {/* 2. Soma-primary Navigation */}
            <div className="flex-1 flex flex-col py-4 gap-1 px-2">
                {primaryNav.map((item) => (
                    <NavItem key={item.href} href={item.href} icon={item.icon} label={item.label} title={item.title} />
                ))}
                {effectiveAdvancedMode && (
                    <div className="mt-3 space-y-1">
                        <div className="hidden px-2 py-1 text-[10px] font-mono uppercase tracking-[0.22em] text-cortex-text-muted/70 md:block">
                            Advanced
                        </div>
                        {advancedNav.map((item) => (
                            <NavItem key={item.href} href={item.href} icon={item.icon} label={item.label} />
                        ))}
                    </div>
                )}
            </div>

            {/* 3. Footer: Advanced Toggle + Settings */}
            <div className="p-2 border-t border-cortex-border space-y-1">
                <button
                    onClick={toggleAdvancedMode}
                    className="flex items-center justify-center md:justify-start w-full p-2.5 rounded-lg transition-all duration-200 text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg"
                    title={effectiveAdvancedMode ? 'Hide advanced panels' : 'Show advanced panels'}
                >
                    {effectiveAdvancedMode ? (
                        <EyeOff className="w-5 h-5 flex-shrink-0" />
                    ) : (
                        <Eye className="w-5 h-5 flex-shrink-0" />
                    )}
                    <span className="hidden md:block ml-3 text-sm font-medium">
                        {effectiveAdvancedMode ? 'Advanced: On' : 'Advanced: Off'}
                    </span>
                </button>
                <NavItem href="/settings" icon={Settings} label="Settings" />
            </div>
        </div>
    );
}

function NavItem({ icon: Icon, label, href, title }: { icon: any; label: string; href: string; title?: string }) {
    const pathname = usePathname();
    const isActive = pathname === href || pathname.startsWith(href + '/');

    return (
        <Link
            href={href}
            title={title ?? label}
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
