import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  experimental: {
    agentFeedback: true,
  },
  cacheComponents: true,
  partialPrefetching: true,
  // Gambar produk & banner diunggah admin ke Cloudinary (lihat backend/README.md).
  images: {
    remotePatterns: [
      { protocol: "https", hostname: "res.cloudinary.com" },
      { protocol: "http", hostname: "localhost" },
    ],
  },
  // Rute lama tetap bisa dibuka dari tautan yang sudah tersebar.
  async redirects() {
    return [
      { source: "/shop-with-sidebar", destination: "/shop", permanent: true },
      { source: "/shop-details", has: [{ type: "query", key: "id", value: "(?<id>.*)" }], destination: "/product/:id", permanent: true },
      { source: "/blog", destination: "/journal", permanent: true },
    ];
  },
  turbopack: {
    rules: {
      "*.css": {
        loaders: ["@tailwindcss/turbopack"],
        as: "*.css",
      },
    },
  },
};

export default nextConfig;
