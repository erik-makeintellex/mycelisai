'use client';

import Link from 'next/link';
import { useEventStream } from '@/hooks/useEventStream';
import { useEffect, useState } from 'react';
import { API_BASE_URL } from '@/config';
import { motion } from 'framer-motion';
import { Activity, Zap, Users } from 'lucide-react';
import StatsCard from '@/components/dashboard/StatsCard';
import StatusPanel from '@/components/dashboard/StatusPanel';
import EventFeed from '@/components/dashboard/EventFeed';

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

  return (
    <div className="space-y-12">
      {/* Hero Section */}
      <motion.header
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6 }}
        className="border-b border-[--border-subtle] pb-8"
      >
        <h1 className="text-5xl font-extrabold tracking-tight">
          <span className="bg-gradient-to-r from-[--text-primary] via-[--accent-info] to-purple-400 bg-clip-text text-transparent">
            Mycelis Service Network
          </span>
        </h1>
        <p className="text-[--text-secondary] mt-3 text-lg">
          Orchestrate your AI agent teams with precision.
        </p>

        <div className="mt-6 flex gap-4">
          <Link
            href="/agents"
            className="group px-6 py-3 bg-transparent border-2 border-[--accent-info] text-[--accent-info] rounded-lg hover:shadow-[0_0_20px_var(--accent-glow)] transition-all duration-300 font-medium relative overflow-hidden"
          >
            <span className="relative z-10">Manage Agents</span>
            <div className="absolute inset-0 bg-[--accent-info]/10 opacity-0 group-hover:opacity-100 transition-opacity duration-300"></div>
          </Link>
          <Link
            href="/config"
            className="group px-6 py-3 bg-transparent border-2 border-[--border-subtle] text-[--text-secondary] rounded-lg hover:border-[--accent-info] hover:text-[--accent-info] hover:shadow-[0_0_20px_var(--accent-glow)] transition-all duration-300 font-medium"
          >
            System Config
          </Link>
        </div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.3, duration: 0.5 }}
          className="mt-6 p-4 bg-[--bg-secondary] border border-[--border-subtle] rounded-lg"
        >
          <h3 className="text-sm font-semibold text-[--text-primary] uppercase tracking-wider mb-2 flex items-center gap-2">
            <Zap className="h-4 w-4 text-[--accent-warn]" />
            Getting Started
          </h3>
          <ol className="list-decimal list-inside text-[--text-secondary] space-y-1 text-sm">
            <li>
              <Link href="/models" className="text-[--accent-info] hover:underline">
                Register AI Models
              </Link>{' '}
              (e.g., OpenAI, Ollama) to power your agents.
            </li>
            <li>
              <Link href="/agents" className="text-[--accent-info] hover:underline">
                Create Agents
              </Link>{' '}
              and define their capabilities and prompts.
            </li>
            <li>
              <Link href="/teams" className="text-[--accent-info] hover:underline">
                Create Teams
              </Link>{' '}
              to group agents and assign them to channels.
            </li>
          </ol>
        </motion.div>
      </motion.header>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <StatsCard title="Total Events" value={stats.totalEvents} icon={Activity} delay={0} />
        <StatsCard title="Events / Sec" value={stats.eventsPerSecond} icon={Zap} delay={0.1} />
        <StatusPanel health={health} delay={0.2} />
      </div>

      {/* Live Event Feed */}
      <EventFeed events={events} channel="sensors" />
    </div>
  );
}
