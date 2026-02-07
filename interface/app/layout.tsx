import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";

import { CommandRail } from "@/components/shell/CommandRail";
import { ActivityStream } from "@/components/shell/ActivityStream";
import { Workspace } from "@/components/shell/Workspace";
import { DecisionOverlay } from "@/components/shell/DecisionOverlay";

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
      <body className={`${inter.variable} ${mono.variable} font-sans bg-zinc-50 text-zinc-900 antialiased overflow-hidden h-screen flex`}>
        {/* ZONE D: Decision Overlay (Z-Index High) */}
        <DecisionOverlay />

        {/* ZONE A: Command Rail (Immutable Left) */}
        <CommandRail />

        {/* ZONE B: Workspace (Fluid Center) */}
        <Workspace>
          {children}
        </Workspace>

        {/* ZONE C: Activity Stream (Immutable Right) */}
        <ActivityStream />
      </body>
    </html>
  );
}
