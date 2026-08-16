import type { NextConfig } from "next";

const nextConfig: NextConfig = {
    allowedDevOrigins: ["192.168.1.50"],
    experimental: {
        serverActions: {
            bodySizeLimit: "10mb",
        }
    },
    images: {
        remotePatterns: [
            {
                protocol: 'https',
                hostname: 's3.perditum.com',
                pathname: '/**',
            },
        ],
    },
};

export default nextConfig;
