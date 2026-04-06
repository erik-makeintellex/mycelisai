import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Mycelis | AI Organizations",
  description: "Soma-primary AI Organizations with governed execution, memory continuity, reviews, and operator control.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning={true}>
      <body className="font-sans antialiased bg-cortex-bg text-cortex-text-main">
        {children}
      </body>
    </html>
  );
}
