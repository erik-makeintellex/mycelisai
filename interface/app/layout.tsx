import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { CommandRail } from "@/components/sov/CommandRail";
import { ActivityStream } from "@/components/sov/ActivityStream";
import { DecisionOverlay } from "@/components/sov/DecisionOverlay";
import { Header } from "@/components/layout/Header";

const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });
const mono = JetBrains_Mono({ subsets: ["latin"], variable: "--font-mono" });

export const metadata: Metadata = {
  title: "Cortex V6.1 | Mycelis",
  description: "Recursive Swarm Operating System",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${inter.variable} ${mono.variable} font-sans bg-[rgb(var(--background))] text-[rgb(var(--foreground))] antialiased overflow-hidden h-screen flex`}>
        {/* ZONE D: Decision Overlay (Z-Index High) */}
        <DecisionOverlay />

        {/* ZONE A: Command Rail (Immutable Left) */}
        <CommandRail />

        {/* Middle Canvas (Workspace + Header) */}
        <main className="flex-1 flex flex-col relative overflow-hidden bg-[rgb(var(--surface))] transition-all duration-300">
          {/* Header can act as the 'Top Rail' or context bar for the Workspace */}
          <Header />

          <div className="flex-1 flex overflow-hidden">
            {/* ZONE B: Active Workspace (Center Stage) */}
            <div className="flex-1 overflow-y-auto p-0 scrollbar-thin scrollbar-thumb-zinc-200 relative">
              {children}
            </div>

            {/* ZONE C: Activity Stream (Immutable Right) */}
            <ActivityStream />
          </div>
        </main>
      </body>
    </html>
  );
}
