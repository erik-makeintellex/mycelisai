import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { Sidebar } from "@/components/shell/Sidebar";
import { TelemetryDeck } from "@/components/shell/TelemetryDeck";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });
const mono = JetBrains_Mono({ subsets: ["latin"], variable: "--font-mono" });

export const metadata: Metadata = {
  title: "Cortex V6 | Mycelis",
  description: "Advanced Agentic Command Console",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${inter.variable} ${mono.variable} font-sans bg-zinc-50 text-zinc-900 antialiased overflow-hidden h-screen flex flex-col`}>
        {/* Main Application Container - Full Height */}
        <div className="flex-1 flex overflow-hidden">
          {/* ZONE A: Sidebar */}
          <aside className="w-64 h-full flex-shrink-0 z-20">
            <Sidebar />
          </aside>

          {/* ZONE B: Active Workspace */}
          <main className="flex-1 relative overflow-hidden flex flex-col bg-white">
            {/* Workspace Content */}
            <div className="flex-1 overflow-y-auto p-6 scrollbar-thin scrollbar-thumb-zinc-200">
              {children}
            </div>

            {/* ZONE C: Telemetry Deck (Pinned to bottom of Main) */}
            <div className="flex-shrink-0 z-10 w-full">
              <TelemetryDeck />
            </div>
          </main>
        </div>
      </body>
    </html>
  );
}
