import Image from "next/image";

export default function Loading() {
  return (
    <div role="status" className="flex min-h-screen flex-col items-center justify-center gap-5 bg-black px-6 text-center text-white">
      <Image src="/Logo.png" alt="" width={180} height={160} priority />
      <div className="h-1 w-[min(454px,80vw)] animate-pulse bg-lime" />
      <p className="text-xl font-bold">Memuat Hii Stranger...</p>
      <p className="text-lg text-faint">Menyiapkan halaman, mohon tunggu sebentar.</p>
      <p className="absolute bottom-16 text-faint">© 2026 Hii Stranger</p>
    </div>
  );
}
