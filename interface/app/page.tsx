import { Activity, Cpu, HardDrive, Network } from "lucide-react";

export default function Home() {
  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight text-zinc-900">
          Operations Dashboard
        </h1>
        <div className="flex gap-2">
          <span className="px-2 py-1 bg-emerald-100 text-emerald-700 text-xs font-medium rounded-full border border-emerald-200">System Choice: Normal</span>
        </div>
      </div>

      {/* KPI Grid */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <MetricCard title="Active Agents" value="3" icon={BotIcon} trend="+1" />
        <MetricCard title="Memory Usage" value="42%" icon={HardDrive} trend="Stable" />
        <MetricCard title="Cognitive Load" value="12%" icon={Cpu} trend="Low" />
        <MetricCard title="Network Activity" value="1.2kb/s" icon={Network} trend="Idle" />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="p-6 rounded-xl border border-zinc-200 bg-white shadow-sm min-h-[300px]">
          <h3 className="text-sm font-semibold text-zinc-500 uppercase tracking-wider mb-4">Live Activity</h3>
          <div className="flex items-center justify-center h-full text-zinc-400">
            Map Visualization Placeholder
          </div>
        </div>

        <div className="p-6 rounded-xl border border-zinc-200 bg-white shadow-sm min-h-[300px]">
          <h3 className="text-sm font-semibold text-zinc-500 uppercase tracking-wider mb-4">Recent Events</h3>
          <div className="space-y-4">
            <EventRow time="10:00" msg="System Initialized" type="info" />
            <EventRow time="10:02" msg="Agent 'Architect' Provisioned" type="success" />
            <EventRow time="10:05" msg="SCIP Contract Validated" type="success" />
          </div>
        </div>
      </div>
    </div>
  );
}

function MetricCard({ title, value, icon: Icon, trend }: any) {
  return (
    <div className="p-4 rounded-xl border border-zinc-200 bg-white shadow-sm hover:shadow-md transition-shadow">
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-medium text-zinc-500 uppercase">{title}</span>
        <Icon className="w-4 h-4 text-zinc-400" />
      </div>
      <div className="text-2xl font-bold text-zinc-900">{value}</div>
      <div className="text-xs text-emerald-600 mt-1">{trend}</div>
    </div>
  )
}

function EventRow({ time, msg, type }: any) {
  return (
    <div className="flex items-center gap-3 text-sm">
      <span className="font-mono text-zinc-400 text-xs">{time}</span>
      <span className={`w-1.5 h-1.5 rounded-full ${type === 'success' ? 'bg-emerald-500' : 'bg-blue-500'}`}></span>
      <span className="text-zinc-700">{msg}</span>
    </div>
  )
}

function BotIcon(props: any) {
  return <Activity {...props} />
}
