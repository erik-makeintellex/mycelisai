'use client';

import Link from 'next/link';
import { useEventStream } from '@/hooks/useEventStream';
import { useEffect, useState } from 'react';
import { API_BASE_URL } from '@/config';

interface HealthStatus {
  status: string;
  components: {
    database: string;
    nats: string;
    ollama?: string;
    api: string;
  };
}

export default function Home() {
  const { events, stats } = useEventStream('sensors');
  const [health, setHealth] = useState<HealthStatus | null>(null);

  useEffect(() => {
    const fetchHealth = async () => {
      try {
        const res = await fetch(`${API_BASE_URL}/health`);
        if (res.ok) {
          setHealth(await res.json());
        }
      } catch (e) {
        console.error("Health check failed", e);
      }
    };
    fetchHealth();
    const interval = setInterval(fetchHealth, 30000);
    return () => clearInterval(interval);
  }, []);

  const getStatusColor = (status: string) => {
    if (status === 'connected' || status === 'healthy') return 'bg-[--accent-success]';
    if (status === 'degraded') return 'bg-[--accent-warn]';
    return 'bg-red-500';
  };

  return (
    <div className="space-y-12">
      <header className="border-b border-[--border-light] pb-8">
        <h1 className="text-4xl font-bold text-[--text-primary] tracking-tight">
          Mycelis Service Network
        </h1>
        <p className="text-[--text-secondary] mt-2 text-lg">
          Orchestrate your AI agent teams with precision.
        </p>

        <div className="mt-6 p-4 bg-[--bg-panel] border border-[--border-light] rounded-lg">
          <h3 className="text-sm font-semibold text-[--text-primary] uppercase tracking-wider mb-2">Getting Started</h3>
          <ol className="list-decimal list-inside text-[--text-secondary] space-y-1 text-sm">
            <li><Link href="/models" className="text-[--accent-link] hover:underline">Register AI Models</Link> (e.g., OpenAI, Ollama) to power your agents.</li>
            <li><Link href="/agents" className="text-[--accent-link] hover:underline">Create Agents</Link> and define their capabilities and prompts.</li>
            <li><Link href="/teams" className="text-[--accent-link] hover:underline">Create Teams</Link> to group agents and assign them to channels.</li>
          </ol>
        </div>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="p-6 border border-[--border-light] rounded-xl bg-[--bg-panel] shadow-lg">
          <h3 className="text-sm font-medium text-[--text-secondary] uppercase tracking-wider">Total Events</h3>
          <p className="text-3xl font-bold mt-2 text-[--text-primary]">{stats.totalEvents}</p>
        </div>
        <div className="p-6 border border-[--border-light] rounded-xl bg-[--bg-panel] shadow-lg">
          <h3 className="text-sm font-medium text-[--text-secondary] uppercase tracking-wider">Events / Sec</h3>
          <p className="text-3xl font-bold mt-2 text-[--text-primary]">{stats.eventsPerSecond}</p>
        </div>
        <div className="p-6 border border-[--border-light] rounded-xl bg-[--bg-panel] shadow-lg">
          <h3 className="text-sm font-medium text-[--text-secondary] uppercase tracking-wider">System Status</h3>
          <div className="flex flex-col gap-2 mt-2">
            <div className="flex items-center justify-between">
              <span className="text-[--text-secondary] text-sm">API</span>
              <div className="flex items-center gap-2">
                <div className={`h-2 w-2 rounded-full ${health ? getStatusColor(health.components.api) : 'bg-[--text-muted]'}`}></div>
                <span className="text-[--text-primary] text-sm">{health?.components.api || 'Checking...'}</span>
              </div>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-[--text-secondary] text-sm">Database</span>
              <div className="flex items-center gap-2">
                <div className={`h-2 w-2 rounded-full ${health ? getStatusColor(health.components.database) : 'bg-[--text-muted]'}`}></div>
                <span className="text-[--text-primary] text-sm">{health?.components.database || 'Checking...'}</span>
              </div>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-[--text-secondary] text-sm">NATS</span>
              <div className="flex items-center gap-2">
                <div className={`h-2 w-2 rounded-full ${health ? getStatusColor(health.components.nats) : 'bg-[--text-muted]'}`}></div>
                <span className="text-[--text-primary] text-sm">{health?.components.nats || 'Checking...'}</span>
              </div>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-[--text-secondary] text-sm">Ollama</span>
              <div className="flex items-center gap-2">
                <div className={`h-2 w-2 rounded-full ${health ? getStatusColor(health.components.ollama || 'unknown') : 'bg-[--text-muted]'}`}></div>
                <span className="text-[--text-primary] text-sm">{health?.components.ollama || 'Checking...'}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="border border-[--border-light] rounded-xl bg-[--bg-panel] shadow-lg overflow-hidden">
        <div className="p-4 border-b border-[--border-light]">
          <h3 className="text-lg font-semibold text-[--text-primary]">Live Event Feed (Channel: sensors)</h3>
        </div>
        <div className="h-64 overflow-y-auto p-4 space-y-2 font-mono text-sm">
          {events.length === 0 ? (
            <p className="text-[--text-muted] italic">Waiting for events...</p>
          ) : (
            events.map((event, idx) => (
              <div key={idx} className="flex gap-4 text-[--text-secondary] border-b border-[--border-light] pb-2 last:border-0">
                <span className="text-[--text-muted]">{new Date(event.timestamp * 1000).toLocaleTimeString()}</span>
                <span className="text-[--accent-success] font-bold">{event.event_type}</span>
                <span className="text-[--text-secondary] truncate">{JSON.stringify(event.payload)}</span>
              </div>
            ))
          )}
        </div>
      </div>

      <div className="flex gap-4">
        <Link
          href="/agents"
          className="px-6 py-3 bg-[--btn-primary-bg] text-[--bg-primary] rounded-lg hover:opacity-90 transition-opacity font-medium shadow-lg"
        >
          Manage Agents
        </Link>
        <Link
          href="/config"
          className="px-6 py-3 bg-[--btn-secondary-bg] text-[--text-primary] border border-[--border-light] rounded-lg hover:bg-[--border-light] transition-colors font-medium"
        >
          System Config
        </Link>
      </div>
    </div>
  );
}

