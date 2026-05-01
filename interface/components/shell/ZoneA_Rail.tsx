"use client";

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { Network, Settings, Home, FolderCog, Brain, Activity, Eye, EyeOff, BookOpen, Building2, Users, Radio } from 'lucide-react';
import { readLastOrganization, subscribeLastOrganizationChange } from '@/lib/lastOrganization';
import { useCortexStore } from '@/store/useCortexStore';

export function ZoneA() {
    const pathname = usePathname();
    const router = useRouter();
    const advancedMode = useCortexStore((s) => s.advancedMode);
    const toggleAdvancedMode = useCortexStore((s) => s.toggleAdvancedMode);
    const setStatusDrawerOpen = useCortexStore((s) => s.setStatusDrawerOpen);
    const [isHydrated, setIsHydrated] = useState(false);
    const [lastOrganization, setLastOrganization] = useState<{ id: string; name: string } | null>(null);

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

    const effectiveAdvancedMode = isHydrated ? advancedMode : false;
    const currentOrganizationHref = lastOrganization ? `/organizations/${lastOrganization.id}` : null;
    const isCurrentOrganizationRoute =
        !!currentOrganizationHref &&
        (pathname === currentOrganizationHref || pathname?.startsWith(currentOrganizationHref + '/') === true);
    const primaryNav = [
        { href: '/dashboard', icon: Home, label: 'Soma', description: 'Home', testId: 'nav-dashboard' },
        { href: '/groups', icon: Users, label: 'Groups', description: 'Focused lanes', testId: 'nav-groups' },
        { href: '/activity', icon: Radio, label: 'Activity', description: 'Runs & bus', testId: 'nav-activity' },
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
        { href: '/resources', icon: FolderCog, label: 'Resources', testId: 'nav-resources' },
        { href: '/memory', icon: Brain, label: 'Memory', testId: 'nav-memory' },
        { href: '/system', icon: Activity, label: 'System', testId: 'nav-system' },
    ];

    return (
        <div className="w-16 md:w-64 bg-cortex-surface text-cortex-text-main flex flex-col border-r border-cortex-border z-50 flex-shrink-0 transition-all duration-300">
            {/* 1. Identity / Logo → Home */}
            <Link href="/" className="h-14 flex items-center justify-center md:justify-start md:px-4 border-b border-cortex-border hover:bg-cortex-bg/50 transition-colors">
                <div className="w-8 h-8 bg-cortex-primary rounded-lg flex items-center justify-center shadow-[0_4px_14px_0_rgba(75,78,109,0.28)]">
                    <Network className="w-5 h-5 text-white" />
                </div>
                <span className="hidden md:block ml-3 font-bold text-sm tracking-widest uppercase text-cortex-text-muted">
                    Mycelis
                </span>
            </Link>

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
                        onClick={item.href === currentOrganizationHref ? () => router.push(item.href) : undefined}
                    />
                ))}
                {effectiveAdvancedMode && (
                    <div className="mt-3 space-y-1">
                        <div className="hidden px-2 py-1 text-[10px] font-mono uppercase tracking-[0.22em] text-cortex-text-muted/70 md:block">
                            Advanced
                        </div>
                        {advancedNav.map((item) => (
                            <NavItem key={item.href} href={item.href} icon={item.icon} label={item.label} testId={item.testId} />
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
                <button
                    type="button"
                    onClick={() => setStatusDrawerOpen(true)}
                    className="flex items-center justify-center md:justify-start w-full p-2.5 rounded-lg transition-all duration-200 text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg"
                    title="Open status drawer"
                >
                    <Activity className="w-5 h-5 flex-shrink-0" />
                    <span className="hidden md:block ml-3 text-sm font-medium">Status</span>
                </button>
                <NavItem href="/settings" icon={Settings} label="Settings" testId="nav-settings" />
            </div>
        </div>
    );
}

function NavItem({ icon: Icon, label, href, title, description, onClick, testId }: { icon: any; label: string; href: string; title?: string; description?: string; onClick?: () => void; testId?: string }) {
    const pathname = usePathname();
    const isActive = pathname === href || pathname?.startsWith(href + '/') === true;
    const classes = `
        flex items-center justify-center md:justify-start w-full p-2.5 rounded-lg transition-all duration-200
        ${isActive
            ? 'bg-cortex-primary text-cortex-bg shadow-[0_4px_14px_0_rgba(75,78,109,0.28)]'
            : 'text-cortex-text-muted hover:text-cortex-text-main hover:bg-cortex-bg'
        }
    `;

    const content = (
        <>
            <Icon className="w-5 h-5 flex-shrink-0" />
            <span className="hidden md:block ml-3 min-w-0">
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
