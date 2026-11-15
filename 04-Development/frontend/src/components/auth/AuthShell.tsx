import Image from "next/image";
import type { ReactNode } from "react";

export default function AuthShell({ headline, children }: { headline: ReactNode; children: ReactNode }) {
  return (
    <div className="grid min-h-screen bg-white sm:bg-canvas lg:grid-cols-2 lg:bg-white">
      <aside className="flex items-center bg-ink px-6 py-5 text-white lg:flex-col lg:items-stretch lg:justify-between lg:gap-10 lg:p-24">
        <Image src="/Logo.png" alt="Hii Stranger" width={134} height={120} className="h-12 w-auto self-start lg:h-[120px]" priority />
        <h2 className="hidden text-6xl font-extrabold leading-[1.25] lg:block">{headline}</h2>
        <p className="hidden text-xl font-bold text-lime lg:block">cz we are not really strangers</p>
      </aside>
      <main className="flex items-start justify-center p-6 sm:pt-32 lg:items-center lg:p-12">
        <div className="w-full max-w-[455px]">{children}</div>
      </main>
    </div>
  );
}
