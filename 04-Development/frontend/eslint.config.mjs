import { defineConfig, globalIgnores } from "eslint/config";
import nextVitals from "eslint-config-next/core-web-vitals";
import nextTs from "eslint-config-next/typescript";

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  {
    rules: {
      // Kode template masih banyak memakai any / variabel tak terpakai: tampilkan sebagai warning saja.
      "@typescript-eslint/no-explicit-any": "warn",
      "@typescript-eslint/no-unused-vars": "warn",
      "react/no-unescaped-entities": "warn",
      // Pola fetch-di-useEffect (reset state lalu muat data) dipakai di banyak komponen; tetap jalan benar.
      "react-hooks/set-state-in-effect": "warn",
    },
  },
  globalIgnores([
    ".next/**",
    "out/**",
    "build/**",
    "next-env.d.ts",
    // Folder referensi (source Next.js & agent skills), bukan bagian aplikasi.
    "rules-direction/**",
  ]),
]);

export default eslintConfig;
