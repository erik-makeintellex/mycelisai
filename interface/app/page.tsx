import Image from "next/image";
import SystemStatus from "@/components/SystemStatus";
import ApprovalDeck from "@/components/ApprovalDeck";
import LogStream from "@/components/LogStream";

export default function Home() {
  return (
    <main className="h-screen w-full bg-slate-900 text-slate-50 p-6 font-mono grid grid-rows-[auto_1fr_400px] gap-6 overflow-hidden">

      {/* 1. Header Area */}
      <header className="flex justify-between items-center border-b border-slate-800 pb-6">
        <div className="flex items-center gap-4">
          <div className="relative w-12 h-12">
            <Image
              src="/logo.png"
              alt="Mycelis Logo"
              fill
              className="object-contain"
              priority
            />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-white">
              MYCELIS <span className="text-blue-500">CORTEX</span>
            </h1>
            <div className="flex items-center gap-2">
              <span className="w-1.5 h-1.5 bg-blue-500 rounded-full animate-pulse"></span>
              <p className="text-[10px] text-slate-500 uppercase tracking-widest font-semibold">
                Neural Infrastructure v3.0
              </p>
            </div>
          </div>
        </div>
        <SystemStatus />
      </header>

      {/* 2. Middle: Governance Deck */}
      <section className="min-h-0 flex flex-col">
        <div className="flex items-center justify-between mb-3 px-1">
          <h3 className="text-xs font-bold text-slate-500 uppercase tracking-widest flex items-center gap-2">
            Governance Control
          </h3>
        </div>
        <div className="flex-1 min-h-0 bg-slate-900/50 rounded-xl border border-slate-800/50 p-1">
          <ApprovalDeck />
        </div>
      </section>

      {/* 3. Bottom: Log Stream */}
      <section className="min-h-0 flex flex-col">
        <h3 className="text-xs font-bold text-slate-500 uppercase tracking-widest mb-3 px-1">
          Neural Activity Stream
        </h3>
        <div className="flex-1 min-h-0 shadow-2xl shadow-blue-900/10">
          <LogStream />
        </div>
      </section>

    </main>
  );
}
