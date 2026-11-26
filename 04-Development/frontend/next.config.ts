import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  cacheComponents: true,
  partialPrefetching: true,
  // Gambar produk & banner diunggah admin ke Cloudinary (lihat backend/README.md).
  images: {
    remotePatterns: [
      { protocol: "https", hostname: "res.cloudinary.com" },
      ...(process.env.NODE_ENV !== "production" ? [{ protocol: "http" as const, hostname: "localhost" }] : []),
    ],
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
