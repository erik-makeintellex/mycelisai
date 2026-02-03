import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { Sidebar } from "@/components/layout/Sidebar";
import { Header } from "@/components/layout/Header";
import { Console } from "@/components/operator/Console";
// import { TelemetryDeck } from "@/components/shell/TelemetryDeck"; // Phase 3

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
      <body className={`${inter.variable} ${mono.variable} font-sans bg-[rgb(var(--background))] text-[rgb(var(--foreground))] antialiased overflow-hidden h-screen flex flex-col`}>
        <div className="flex-1 flex overflow-hidden">
          {/* ZONE A: Sidebar (Navigation Spine) */}
          <Sidebar />

          {/* ZONE B: Active Workspace */}
          <main className="flex-1 relative overflow-hidden flex flex-col bg-[rgb(var(--surface))] ml-64 transition-all duration-300">
            {/* Global Header */}
            <Header />

            {/* Workspace Content */}
            <div className="flex-1 overflow-y-auto p-0 scrollbar-thin scrollbar-thumb-zinc-200 pb-12">
              {children}
            </div>

            {/* ZONE C: Operator Console (Phase 2) */}
            <Console />
          </main>
        </div>
      </body>
    </html>
  );
}
