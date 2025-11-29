'use client';

import Link from 'next/link';
import { useEventStream } from '@/hooks/useEventStream';

export default function Home() {
  const { events, stats } = useEventStream('sensors'); // Listening to 'sensors' channel for demo

  return (
    <div className="space-y-12">
      <header className="border-b border-zinc-700 pb-8">
        <h1 className="text-4xl font-bold text-zinc-100 tracking-tight">
          Mycelis Service Network
        </h1>
        <p className="text-zinc-400 mt-2 text-lg">
          Orchestrate your AI agent teams with precision.
        </p>

        <div className="mt-6 p-4 bg-zinc-800/50 border border-zinc-700 rounded-lg">
          <h3 className="text-sm font-semibold text-zinc-200 uppercase tracking-wider mb-2">Getting Started</h3>
          <ol className="list-decimal list-inside text-zinc-400 space-y-1 text-sm">
            <li><Link href="/models" className="text-emerald-400 hover:underline">Register AI Models</Link> (e.g., OpenAI, Ollama) to power your agents.</li>
            <li><Link href="/agents" className="text-emerald-400 hover:underline">Create Agents</Link> and define their capabilities and prompts.</li>
            <li><Link href="/teams" className="text-emerald-400 hover:underline">Create Teams</Link> to group agents and assign them to channels.</li>
          </ol>
        </div>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="p-6 border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg">
          <h3 className="text-sm font-medium text-zinc-400 uppercase tracking-wider">Total Events</h3>
          <p className="text-3xl font-bold mt-2 text-zinc-100">{stats.totalEvents}</p>
        </div>
        <div className="p-6 border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg">
          <h3 className="text-sm font-medium text-zinc-400 uppercase tracking-wider">Events / Sec</h3>
          <p className="text-3xl font-bold mt-2 text-zinc-100">{stats.eventsPerSecond}</p>
        </div>
        <div className="p-6 border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg">
          <h3 className="text-sm font-medium text-zinc-400 uppercase tracking-wider">System Status</h3>
          <div className="flex items-center gap-2 mt-2">
            <div className={`h-3 w-3 rounded-full ${stats.isConnected ? 'bg-emerald-500' : 'bg-red-500'}`}></div>
            <p className="text-3xl font-bold text-zinc-100">{stats.isConnected ? 'Online' : 'Disconnected'}</p>
          </div>
        </div>
      </div>

      <div className="border border-zinc-700 rounded-xl bg-zinc-800 shadow-lg overflow-hidden">
        <div className="p-4 border-b border-zinc-700">
          <h3 className="text-lg font-semibold text-zinc-200">Live Event Feed (Channel: sensors)</h3>
        </div>
        <div className="h-64 overflow-y-auto p-4 space-y-2 font-mono text-sm">
          {events.length === 0 ? (
            <p className="text-zinc-500 italic">Waiting for events...</p>
          ) : (
            events.map((event, idx) => (
              <div key={idx} className="flex gap-4 text-zinc-300 border-b border-zinc-700/50 pb-2 last:border-0">
                <span className="text-zinc-500">{new Date(event.timestamp * 1000).toLocaleTimeString()}</span>
                <span className="text-emerald-400 font-bold">{event.event_type}</span>
                <span className="text-zinc-400 truncate">{JSON.stringify(event.payload)}</span>
              </div>
            ))
          )}
        </div>
      </div>

      <div className="flex gap-4">
        <Link
          href="/agents"
          className="px-6 py-3 bg-zinc-100 text-zinc-900 rounded-lg hover:bg-white transition-colors font-medium shadow-lg shadow-zinc-900/20"
        >
          Manage Agents
        </Link>
        <Link
          href="/config"
          className="px-6 py-3 border border-zinc-600 text-zinc-300 rounded-lg hover:bg-zinc-700 transition-colors font-medium"
        >
          System Config
        </Link>
      </div>
    </div>
  );
}
