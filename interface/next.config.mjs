/** @type {import('next').NextConfig} */
const nextConfig = {
    async rewrites() {
        return [
            {
                source: '/api/:path*',
                destination: 'http://localhost:8081/api/:path*', // Proxy to Backend
            },
            {
                source: '/admin/:path*',
                destination: 'http://localhost:8081/admin/:path*', // Proxy Governance
            },
            {
                source: '/agents',
                destination: 'http://localhost:8081/agents', // Proxy Agent Registry
            },
            {
                source: '/healthz',
                destination: 'http://localhost:8081/healthz', // Proxy Health Check
            },
        ]
    },
};

export default nextConfig;
