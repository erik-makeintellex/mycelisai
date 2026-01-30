'use client';

import { useState, useEffect } from 'react';
import { API_BASE_URL } from '@/config';

interface HealthStatus {
    status: string;
    timestamp: string;
    components: {
        database: string;
        nats: string;
        api: string;
    };
    environment: {
        nats_url: string;
        database_url: string;
        backend_host: string;
    };
}

export default function MonitorPage() {
    const [health, setHealth] = useState<HealthStatus | null>(null);
    const [agents, setAgents] = useState<any[]>([]);
    const [teams, setTeams] = useState<any[]>([]);
    const [services, setServices] = useState<any[]>([]);
    const [loading, setLoading] = useState(true);

    const fetchData = async () => {
        try {
            const [healthRes, agentsRes, teamsRes, servicesRes] = await Promise.all([
                fetch(`${API_BASE_URL}/health`),
                fetch(`${API_BASE_URL}/agents`),
                fetch(`${API_BASE_URL}/teams`),
                fetch(`${API_BASE_URL}/services`)
            ]);

            if (healthRes.ok) setHealth(await healthRes.json());
            if (agentsRes.ok) setAgents(await agentsRes.json());
            if (teamsRes.ok) setTeams(await teamsRes.json());
            if (servicesRes.ok) setServices(await servicesRes.json());
        } catch (error) {
            console.error('Failed to fetch monitor data:', error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 5000); // Refresh every 5s
        return () => clearInterval(interval);
    }, []);

    if (loading && !health) {
        return <div className="p-8 text-zinc-400">Loading system status...</div>;
    }

    return (
        <div className="space-y-8">
            <div>
                <h1 className="text-3xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-emerald-400 to-cyan-400 mb-2">
                    System Monitor
                </h1>
                <p className="text-zinc-400">Real-time environment state and configuration.</p>
            </div>

            {/* System Health Cards */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <StatusCard
                    title="Overall Health"
                    status={health?.status || 'unknown'}
                    details={health?.timestamp}
                />
                <StatusCard
                    title="Database"
                    status={health?.components.database === 'connected' ? 'healthy' : 'error'}
                    details={health?.components.database}
                />
                <StatusCard
                    title="Messaging (NATS)"
                    status={health?.components.nats === 'connected' ? 'healthy' : 'error'}
                    details={health?.components.nats}
                />
            </div>

            {/* Resource Overview */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <CountCard title="Active Agents" count={agents.length} href="/agents" />
                <CountCard title="Active Teams" count={teams.length} href="/teams" />
                <CountCard title="Registered Services" count={services.length} href="/services" />
            </div>

            {/* Environment Config */}
            <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
                <h2 className="text-xl font-semibold text-zinc-200 mb-4">Environment Configuration</h2>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                    <ConfigItem label="NATS URL" value={health?.environment.nats_url} />
                    <ConfigItem label="Database" value={health?.environment.database_url} />
                    <ConfigItem label="Backend Host" value={health?.environment.backend_host} />
                    <ConfigItem label="API Endpoint" value={API_BASE_URL} />
                </div>
            </div>
        </div>
    );
}

function StatusCard({ title, status, details }: { title: string, status: string, details?: string }) {
    const isHealthy = status === 'healthy' || status === 'connected';
    return (
        <div className={`p-6 rounded-xl border ${isHealthy ? 'border-emerald-500/20 bg-emerald-500/5' : 'border-red-500/20 bg-red-500/5'}`}>
            <div className="flex justify-between items-start mb-2">
                <h3 className="font-semibold text-zinc-200">{title}</h3>
                <div className={`w-3 h-3 rounded-full ${isHealthy ? 'bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.5)]' : 'bg-red-500 shadow-[0_0_10px_rgba(239,68,68,0.5)]'}`} />
            </div>
            <p className={`text-2xl font-bold ${isHealthy ? 'text-emerald-400' : 'text-red-400'} capitalize`}>
                {status}
            </p>
            {details && <p className="text-xs text-zinc-500 mt-2 font-mono">{details}</p>}
        </div>
    );
}

function CountCard({ title, count, href }: { title: string, count: number, href: string }) {
    return (
        <a href={href} className="block p-6 rounded-xl border border-zinc-800 bg-zinc-900/50 hover:border-zinc-700 transition-colors">
            <h3 className="text-zinc-400 text-sm font-medium mb-2">{title}</h3>
            <p className="text-4xl font-bold text-zinc-100">{count}</p>
        </a>
    );
}

function ConfigItem({ label, value }: { label: string, value?: string }) {
    return (
        <div className="flex justify-between items-center py-2 border-b border-zinc-800 last:border-0">
            <span className="text-zinc-500">{label}</span>
            <code className="text-zinc-300 bg-zinc-950 px-2 py-1 rounded">{value || 'N/A'}</code>
        </div>
    );
}
