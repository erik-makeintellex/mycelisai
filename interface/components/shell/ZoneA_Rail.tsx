"use client";

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { Network, Settings, Home, FolderCog, Brain, Activity, Eye, EyeOff, BookOpen, Building2, Users, Radio, LogOut, PanelLeftClose, PanelLeftOpen } from 'lucide-react';
import { readLastOrganization, subscribeLastOrganizationChange } from '@/lib/lastOrganization';
import { useCortexStore } from '@/store/useCortexStore';

export function ZoneA() {
    const pathname = usePathname();
    const router = useRouter();
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const railCollapsed = useCortexStore((s) => s.railCollapsed);
    const toggleAdvancedMode = useCortexStore((s) => s.toggleAdvancedMode);
    const toggleRailCollapsed = useCortexStore((s) => s.toggleRailCollapsed);
    const setStatusDrawerOpen = useCortexStore((s) => s.setStatusDrawerOpen);
    const [isHydrated, setIsHydrated] = useState(false);
    const [lastOrganization, setLastOrganization] = useState<{ id: string; name: string } | null>(null);
    const [webRole, setWebRole] = useState<'admin' | 'standard'>('admin');

    useEffect(() => {
        const syncLastOrganization = () => {
            setLastOrganization(readLastOrganization());
        };

        setIsHydrated(true);
        syncLastOrganization();
        return subscribeLastOrganizationChange((organization) => {
            setLastOrganization(organization);
        });
    }, [pathname]);
    useEffect(() => {
        let cancelled = false;
        const request = fetch('/auth/session', { cache: 'no-store' });
        if (!request?.then) return () => {
            cancelled = true;
        };
        request.then((res) => (res.ok ? res.json() : null))
            .then((body) => {
                if (!cancelled && body?.data?.user?.role) setWebRole(body.data.user.role === 'admin' ? 'admin' : 'standard');
            })
            .catch(() => undefined);
        return () => {
            cancelled = true;
        };
    }, []);

    const isAdmin = webRole === 'admin';
    const effectiveAdvancedMode = isHydrated && isAdmin ? advancedMode : false;
    const currentOrganizationHref = lastOrganization ? `/organizations/${lastOrganization.id}` : null;
    const isCurrentOrganizationRoute =
        !!currentOrganizationHref &&
        (pathname === currentOrganizationHref || pathname?.startsWith(currentOrganizationHref + '/') === true);
    const primaryNav = [
        { href: '/dashboard', icon: Home, label: 'Soma', description: 'Ask first', testId: 'nav-dashboard' },
        { href: '/groups', icon: Users, label: 'Groups', description: 'Outputs', testId: 'nav-groups' },
        { href: '/resources', icon: FolderCog, label: 'Resources', description: 'Files & tools', testId: 'nav-resources' },
        ...(lastOrganization ? [{
            href: currentOrganizationHref!,
            icon: Building2,
            label: isCurrentOrganizationRoute ? 'Current Organization' : 'Return to Organization',
            title: lastOrganization.name,
            description: lastOrganization.name,
            testId: 'current-organization-nav',
        }] : []),
        { href: '/docs', icon: BookOpen, label: 'Docs', testId: 'nav-docs' },
    ];
    const advancedNav = [
        { href: '/activity', icon: Radio, label: 'Activity', description: 'Runs & bus', testId: 'nav-activity' },
        { href: '/memory', icon: Brain, label: 'Memory', testId: 'nav-memory' },
        { href: '/system', icon: Activity, label: 'System', testId: 'nav-system' },
    ];

    return (
        <div
            className={`bg-cortex-surface text-cortex-text-main flex flex-col border-r border-cortex-border z-50 flex-shrink-0 transition-all duration-300 ${
                railCollapsed ? 'w-16' : 'w-16 md:w-64'
            }`}
            data-testid="zone-a-rail"
            data-collapsed={railCollapsed ? 'true' : 'false'}
        >
            {/* 1. Identity / Logo → Home */}
            <div className="h-14 flex items-center gap-1 border-b border-cortex-border px-2">
                <Link
                    href="/"
                    className={`flex min-w-0 flex-1 items-center rounded-lg transition-colors hover:bg-cortex-bg/50 ${
                        railCollapsed ? 'justify-center px-0 py-1' : 'justify-center md:justify-start md:px-2 py-1'
                    }`}
                    title="Mycelis home"
                >
                    <div className="w-8 h-8 bg-cortex-primary rounded-lg flex items-center justify-center shadow-[0_4px_14px_0_rgba(75,78,109,0.28)]">
                        <Network className="w-5 h-5 text-white" />
                    </div>
                    <span className={`${railCollapsed ? 'hidden' : 'hidden md:block'} ml-3 font-bold text-sm tracking-widest uppercase text-cortex-text-muted`}>
                        Mycelis
                    </span>
                </Link>
                <button
                    type="button"
                    onClick={toggleRailCollapsed}
                    className="hidden h-8 w-8 items-center justify-center rounded-lg text-cortex-text-muted transition-colors hover:bg-cortex-bg hover:text-cortex-text-main md:flex"
                    title={railCollapsed ? 'Expand navigation' : 'Collapse navigation'}
                    aria-label={railCollapsed ? 'Expand navigation' : 'Collapse navigation'}
                    data-testid="rail-collapse-toggle"
                >
                    {railCollapsed ? <PanelLeftOpen className="h-4 w-4" /> : <PanelLeftClose className="h-4 w-4" />}
                </button>
            </div>

            {/* 2. Soma-primary Navigation */}
            <div className="flex-1 flex flex-col py-4 gap-1 px-2">
                {primaryNav.map((item) => (
                    <NavItem
                        key={item.href}
                        href={item.href}
                        icon={item.icon}
                        label={item.label}
                        title={item.title}
                        description={item.description}
                        testId={item.testId}
                        collapsed={railCollapsed}
                        onClick={item.href === currentOrganizationHref ? () => router.push(item.href) : undefined}
                    />
                ))}
                {effectiveAdvancedMode && (
                    <div className="mt-3 space-y-1">
                        <div className={`${railCollapsed ? 'hidden' : 'hidden md:block'} px-2 py-1 text-[10px] font-mono uppercase tracking-[0.22em] text-cortex-text-muted/70`}>
                            Admin tools
                        </div>
                        {advancedNav.map((item) => (
                            <NavItem key={item.href} href={item.href} icon={item.icon} label={item.label} testId={item.testId} collapsed={railCollapsed} />
                        ))}
                    </div>
                )}
            </div>

            {/* 3. Footer: Advanced Toggle + Settings */}
            <div className="p-2 border-t border-cortex-border space-y-1">
                {isAdmin && (
                    <button
                        onClick={toggleAdvancedMode}
                        className={`flex items-center justify-center ${railCollapsed ? '' : 'md:justify-start'} w-full p-2.5 rounded-lg transition-all duration-200 text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg`}
                        title={effectiveAdvancedMode ? 'Hide advanced panels' : 'Show advanced panels'}
                    >
                        {effectiveAdvancedMode ? (
                            <EyeOff className="w-5 h-5 flex-shrink-0" />
                        ) : (
                            <Eye className="w-5 h-5 flex-shrink-0" />
                        )}
                        <span className={`${railCollapsed ? 'hidden' : 'hidden md:block'} ml-3 text-sm font-medium`}>
                            {effectiveAdvancedMode ? 'Admin tools: On' : 'Admin tools: Off'}
                        </span>
                    </button>
                )}
                <button
                    type="button"
                    onClick={() => setStatusDrawerOpen(true)}
                    className={`flex items-center justify-center ${railCollapsed ? '' : 'md:justify-start'} w-full p-2.5 rounded-lg transition-all duration-200 text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg`}
                    title="Open status drawer"
                >
                    <Activity className="w-5 h-5 flex-shrink-0" />
                    <span className={`${railCollapsed ? 'hidden' : 'hidden md:block'} ml-3 text-sm font-medium`}>Status</span>
                </button>
                <NavItem href="/settings" icon={Settings} label="Settings" testId="nav-settings" collapsed={railCollapsed} />
                <form action="/auth/logout" method="post">
                    <button
                        type="submit"
                        className={`flex w-full items-center justify-center rounded-lg p-2.5 text-cortex-text-muted transition-all duration-200 hover:bg-cortex-bg hover:text-cortex-text-main ${railCollapsed ? '' : 'md:justify-start'}`}
                        title="Sign out"
                    >
                        <LogOut className="h-5 w-5 flex-shrink-0" />
                        <span className={`${railCollapsed ? 'hidden' : 'hidden md:block'} ml-3 text-sm font-medium`}>Sign out</span>
                    </button>
                </form>
            </div>
        </div>
    );
}

function NavItem({ icon: Icon, label, href, title, description, onClick, testId, collapsed = false }: { icon: any; label: string; href: string; title?: string; description?: string; onClick?: () => void; testId?: string; collapsed?: boolean }) {
    const pathname = usePathname();
    const isActive = pathname === href || pathname?.startsWith(href + '/') === true;
    const classes = `
        flex items-center justify-center ${collapsed ? '' : 'md:justify-start'} w-full p-2.5 rounded-lg transition-all duration-200
        ${isActive
            ? 'bg-cortex-primary text-cortex-bg shadow-[0_4px_14px_0_rgba(75,78,109,0.28)]'
            : 'text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg'
        }
    `;

    const content = (
        <>
            <Icon className="w-5 h-5 flex-shrink-0" />
            <span className={`${collapsed ? 'hidden' : 'hidden md:block'} ml-3 min-w-0`}>
                <span className="block truncate text-sm font-medium">{label}</span>
                {description ? (
                    <span className={`block truncate text-xs ${isActive ? 'text-cortex-bg/80' : 'text-cortex-text-muted/80'}`}>{description}</span>
                ) : null}
            </span>
        </>
    );

    if (onClick) {
        return (
            <button type="button" title={title ?? label} className={classes} onClick={onClick} data-testid={testId}>
                {content}
            </button>
        );
    }

    return (
        <Link href={href} title={title ?? label} className={classes} data-testid={testId}>
            {content}
        </Link>
    );
}
