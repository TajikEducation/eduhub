import type { Metadata } from "next";
import { Montserrat, Inter } from "next/font/google";
import { C } from "@/lib/data";
import { StoreProvider } from "@/lib/store/StoreProvider";
import { Toaster } from "@/shared/components/sonner";
import "./globals.css";

const montserrat = Montserrat({
  variable: "--font-montserrat",
  subsets: ["latin", "cyrillic"],
  weight: ["400", "500", "600", "700", "800", "900"],
});

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin", "cyrillic"],
  weight: ["400", "500", "600"],
});

export const metadata: Metadata = {
  title: "EduHub Tajikistan",
  description:
    "Национальная платформа поиска, сравнения и оценки образовательных учреждений Таджикистана.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ru" className={`${montserrat.variable} ${inter.variable}`}>
      <body style={{ minHeight: "100vh", background: C.bg, color: C.text }}>
        <StoreProvider>{children}</StoreProvider>
        <Toaster position="bottom-center" />
      </body>
    </html>
  );
}
