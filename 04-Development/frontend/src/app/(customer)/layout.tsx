import Footer from "@/components/layout/Footer";
import Navbar from "@/components/layout/Navbar";

export default function CustomerLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col">
      <Navbar />
      <main className="mx-auto w-full max-w-360 flex-1 px-6 py-6 sm:px-9 sm:py-8 lg:px-14 lg:py-12">{children}</main>
      <Footer />
    </div>
  );
}
