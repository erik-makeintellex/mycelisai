import type { Metadata } from "next";
import { Source_Sans_3, JetBrains_Mono } from "next/font/google";
import "./globals.css";

const ui = Source_Sans_3({
  subsets: ["latin"],
  variable: "--font-ui",
  display: "swap",
});
const mono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-code",
  display: "swap",
});

export const metadata: Metadata = {
  title: "Cortex V6.2 | Mycelis",
  description: "Recursive Swarm Operating System",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning={true}>
      <body className={`${ui.variable} ${mono.variable} font-sans antialiased bg-cortex-bg text-cortex-text-main`}>
        {children}
      </body>
    </html>
  );
}
