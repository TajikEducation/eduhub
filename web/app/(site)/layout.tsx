import { NavBar } from "@/components/NavBar";
import { Footer } from "@/components/Footer";
import { MaintenanceBanner } from "@/components/MaintenanceBanner";
import { AppStateProvider } from "@/lib/app-state";

export default function SiteLayout({ children }: { children: React.ReactNode }) {
  return (
    <AppStateProvider>
      <NavBar />
      <MaintenanceBanner />
      {children}
      <Footer />
    </AppStateProvider>
  );
}
