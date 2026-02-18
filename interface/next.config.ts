import type { NextConfig } from "next";

const isStaticExport = process.env.NEXT_PUBLIC_STATIC_EXPORT === "true";

const nextConfig: NextConfig = {
  output: isStaticExport ? "export" : undefined,
  images: {
    unoptimized: isStaticExport,
  },
  // Proxy /api/* and legacy admin paths to the Go backend.
  // Rewrites are not supported in static export mode.
  async rewrites() {
    if (isStaticExport) return [];

    const host = process.env.MYCELIS_API_HOST || "localhost";
    const port = process.env.MYCELIS_API_PORT || "8081";
    const backend = `http://${host}:${port}`;
    return [
      {
        source: "/api/:path*",
        destination: `${backend}/api/:path*`,
      },
      {
        source: "/admin/:path*",
        destination: `${backend}/admin/:path*`,
      },
      {
        source: "/agents",
        destination: `${backend}/agents`,
      },
      {
        source: "/healthz",
        destination: `${backend}/healthz`,
      },
    ];
  },
};

export default nextConfig;
