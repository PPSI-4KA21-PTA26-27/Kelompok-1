import Footer from "@/components/layout/Footer";
import MaintenanceGuard from "@/components/layout/MaintenanceGuard";
import Navbar from "@/components/layout/Navbar";
import VerifyGuard from "@/components/layout/VerifyGuard";

export default function CustomerLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col">
      <MaintenanceGuard />
      <VerifyGuard />
      <Navbar />
      <main className="mx-auto w-full max-w-[1440px] flex-1 px-6 py-6 sm:px-9 sm:py-8 lg:px-14 lg:py-12">{children}</main>
      <Footer />
    </div>
  );
}
