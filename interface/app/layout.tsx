
import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { Vitality } from "@/components/hud/Vitality";

const inter = Inter({ subsets: ["latin"], variable: "--font-sans" });
const mono = JetBrains_Mono({ subsets: ["latin"], variable: "--font-mono" });

export const metadata: Metadata = {
  title: "Mycelis V6 | The Grimoire",
  description: "Neural Organism Interface",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.variable} ${mono.variable} font-sans bg-slate-950 text-slate-100 h-screen w-screen overflow-hidden flex flex-col`}>
        {/* HEADER */}
        <header className="h-14 border-b border-slate-800 flex items-center justify-between px-4 bg-slate-900/50 backdrop-blur">
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-cyan-500 rounded-full shadow-[0_0_10px_#06b6d4]"></div>
            <h1 className="font-bold tracking-widest text-sm text-slate-300">MYCELIS <span className="text-cyan-500">GRIMOIRE</span></h1>
          </div>
          <Vitality />
        </header>

        {/* WORKSPACE */}
        <div className="flex-1 flex overflow-hidden">
          {/* LEFT PANE: LIBRARY */}
          <aside className="w-[250px] border-r border-slate-800 bg-slate-900/30 flex flex-col">
            <div className="p-3 border-b border-slate-800">
              <h2 className="text-xs font-mono text-slate-500 uppercase">Library</h2>
            </div>
            <div className="p-4 text-xs text-slate-600 italic">
              [Agents Configured]
            </div>
          </aside>

          {/* CENTER: CANVAS (Children) */}
          <main className="flex-1 relative bg-slate-950/50">
            {children}
          </main>
        </div>

        {/* BOTTOM PANE: TERMINAL */}
        <div className="h-[250px] border-t border-slate-800 bg-black/80 font-mono text-xs p-2 overflow-y-auto">
          <div className="text-slate-500 mb-1">Allows connection to Synaptic Injector...</div>
          <div className="text-green-500">$ _</div>
        </div>
      </body>
    </html>
  );
}
