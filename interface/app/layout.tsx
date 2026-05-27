import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Mycelis | AI Organizations",
  description: "Soma-primary AI Organizations with governed execution, memory continuity, reviews, and operator control.",
};

const themeBootScript = `
(() => {
  try {
    const raw = window.localStorage.getItem("mycelis-user-settings");
    const parsed = raw ? JSON.parse(raw) : {};
    let theme = parsed && typeof parsed.theme === "string" ? parsed.theme : "";
    if (theme !== "aero-light" && theme !== "midnight-cortex" && theme !== "system") {
      theme = "aero-light";
    }
    if (theme === "system") {
      theme = window.matchMedia("(prefers-color-scheme: dark)").matches ? "midnight-cortex" : "aero-light";
    }
    document.documentElement.dataset.theme = theme;
  } catch {
    document.documentElement.dataset.theme = "aero-light";
  }
})();
`;

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning={true}>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeBootScript }} />
      </head>
      <body className="font-sans antialiased bg-cortex-bg text-cortex-text-main">
        {children}
      </body>
    </html>
  );
}
