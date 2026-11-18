import type { Metadata, Viewport } from "next";
import localFont from "next/font/local";
import "./css/style.css";
import { Suspense } from "react";
import Providers from "./providers";
import { RouteProgress } from "@/components/ui/Loader";

// Inter variable (frontend/font); satu berkas mencakup bobot 100-900
const inter = localFont({
  src: "../../font/Inter-VariableFont_opsz,wght.ttf",
  weight: "100 900",
  variable: "--font-inter",
  display: "swap",
});

export const metadata: Metadata = {
  title: { default: "Hii Stranger — Streetwear Indonesia", template: "%s · Hii Stranger" },
  description:
    "Hii Stranger: streetwear heavyweight dengan rilis terbatas. Belanja kaos, hoodie, dan celana dengan pembayaran QRIS.",
};

export const viewport: Viewport = {
  themeColor: "#f4f4f1",
  width: "device-width",
  initialScale: 1,
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="id" data-scroll-behavior="smooth" className={`${inter.variable}`} suppressHydrationWarning>
      <body>
        <Suspense fallback={null}>
          <RouteProgress />
        </Suspense>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
