import { fileURLToPath } from "node:url";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) },
  },
  test: {
    environment: "jsdom",
    include: ["src/**/*.test.{ts,tsx}"],
    setupFiles: ["./src/test/setup.ts"],
    // API_URL dibaca saat modul dimuat; garis miring di akhir sengaja untuk menguji pemangkasannya.
    env: { NEXT_PUBLIC_API_URL: "http://api.test/api/v1/" },
    // Berurutan: satu worker, file tes tidak paralel (hemat CPU).
    fileParallelism: false,
    maxWorkers: 1,
    restoreMocks: true,
    unstubGlobals: true,
  },
});
