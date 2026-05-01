import type { NextConfig } from "next";
import path from "path";

const isStaticExport = process.env.NEXT_PUBLIC_STATIC_EXPORT === "true";
const projectRoot = path.resolve(__dirname, "..");

const nextConfig: NextConfig = {
  output: isStaticExport ? "export" : "standalone",
  outputFileTracingRoot: projectRoot,
  outputFileTracingIncludes: {
    "/docs-api/[slug]": [
      "./README.md",
      "../README.md",
      "./.state/V8_DEV_STATE.md",
      "../.state/V8_DEV_STATE.md",
      "./.state/V7_DEV_STATE.md",
      "../.state/V7_DEV_STATE.md",
      "./architecture/v8-2.md",
      "../architecture/v8-2.md",
      "./architecture/mycelis-architecture-v7.md",
      "../architecture/mycelis-architecture-v7.md",
      "./docs/**/*",
      "../docs/**/*",
    ],
  },
  images: {
    unoptimized: isStaticExport,
  },
  experimental: {
    cpus: 1,
    staticGenerationMaxConcurrency: 1,
  },
  allowedDevOrigins: ["127.0.0.1", "localhost", "::1"],
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
