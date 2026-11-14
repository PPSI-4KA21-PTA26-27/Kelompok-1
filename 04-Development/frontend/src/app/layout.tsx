import type { Metadata, Viewport } from "next";
import { Bricolage_Grotesque, Geist } from "next/font/google";
import "./css/style.css";
import { Suspense } from "react";
import Providers from "./providers";
import { RouteProgress } from "@/components/ui/Loader";

const geist = Geist({ subsets: ["latin"], variable: "--font-geist", display: "swap" });
const bricolage = Bricolage_Grotesque({
  subsets: ["latin"],
  weight: ["600", "700", "800"],
  variable: "--font-bricolage",
  display: "swap",
});

export const metadata: Metadata = {
  title: { default: "Hii Stranger — Streetwear Indonesia", template: "%s · Hii Stranger" },
  description:
    "Hii Stranger: streetwear heavyweight dengan rilis terbatas. Belanja kaos, hoodie, dan celana dengan pembayaran QRIS.",
};

export const viewport: Viewport = {
  themeColor: "#fbf9f5",
  width: "device-width",
  initialScale: 1,
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="id" className={`${geist.variable} ${bricolage.variable}`} suppressHydrationWarning>
      <body>
        <Suspense fallback={null}>
          <RouteProgress />
        </Suspense>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
