/** @type {import('next').NextConfig} */
const isStaticExport = process.env.NEXT_PUBLIC_STATIC_EXPORT === 'true';

const nextConfig = {
    output: isStaticExport ? 'export' : undefined,
    images: {
        unoptimized: isStaticExport,
    },
    // Only enable rewrites if NOT exporting (rewrites are not supported in static export)
    async rewrites() {
        if (isStaticExport) return [];

        const host = process.env.MYCELIS_API_HOST || 'localhost';
        const port = process.env.MYCELIS_API_PORT || '8081';
        const backend = `http://${host}:${port}`;
        return [
            {
                source: '/api/:path*',
                destination: `${backend}/api/:path*`, // Proxy to Backend
            },
            {
                source: '/admin/:path*',
                destination: `${backend}/admin/:path*`, // Proxy Governance
            },
            {
                source: '/agents',
                destination: `${backend}/agents`, // Proxy Agent Registry
            },
            {
                source: '/healthz',
                destination: `${backend}/healthz`, // Proxy Health Check
            },
        ]
    },
};

export default nextConfig;
