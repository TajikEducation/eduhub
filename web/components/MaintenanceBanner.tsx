"use client";

import { AlertTriangle } from "lucide-react";
import { C, FH } from "@/lib/data";
import { useAppState } from "@/lib/app-state";
import { useT } from "@/lib/i18n";

export function MaintenanceBanner() {
  const t = useT();
  const { platformSettings } = useAppState();
  if (!platformSettings.maintenanceMode) return null;

  return (
    <div style={{ background: C.red, color: "#fff", padding: "8px 16px", textAlign: "center", display: "flex", alignItems: "center", justifyContent: "center", gap: 8, fontFamily: FH, fontWeight: 700, fontSize: 13 }}>
      <AlertTriangle size={14} />
      {t({ ru: "Ведутся технические работы — часть функций может быть недоступна", tg: "Корҳои техникӣ идома доранд — баъзе функсияҳо дастнорас буда метавонанд" })}
    </div>
  );
}
