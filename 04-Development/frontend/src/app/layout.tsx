import type { Metadata, Viewport } from "next";
import { Inter } from "next/font/google";
import "./css/style.css";
import { Suspense } from "react";
import Providers from "./providers";
import { RouteProgress } from "@/components/ui/Loader";

const inter = Inter({
  subsets: ["latin"],
  weight: ["400", "500", "600", "700", "800", "900"],
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
    <html lang="id" className={`${inter.variable}`} suppressHydrationWarning>
      <body>
        <Suspense fallback={null}>
          <RouteProgress />
        </Suspense>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
