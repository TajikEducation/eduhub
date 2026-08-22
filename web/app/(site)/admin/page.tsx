"use client";

import { useState } from "react";
import Link from "next/link";
import { LayoutDashboard, Settings, Building2, Star, Users, Briefcase, ShieldAlert, ArrowRight } from "lucide-react";
import { C, FH, FB, INSTITUTIONS, ALL_STAFF, VACANCIES, REVIEWS } from "@/lib/data";
import { useAppState } from "@/lib/app-state";
import { useT } from "@/lib/i18n";

type Tab = "overview" | "settings";

const inputStyle: React.CSSProperties = {
  width: "100%", padding: "10px 13px", borderRadius: 10, border: `1px solid ${C.border}`,
  background: C.s2, color: C.text, fontFamily: FB, fontSize: 13.5, outline: "none", boxSizing: "border-box",
};

function FormField({ label, children }: { label: React.ReactNode; children: React.ReactNode }) {
  return (
    <div style={{ marginBottom: 14 }}>
      <label style={{ display: "block", fontSize: 11.5, fontWeight: 700, color: C.dim, textTransform: "uppercase", letterSpacing: ".05em", fontFamily: FH, marginBottom: 6 }}>{label}</label>
      {children}
    </div>
  );
}

export default function AdminPage() {
  const t = useT();
  const { myInstitution, platformSettings, setPlatformSettings } = useAppState();
  const [tab, setTab] = useState<Tab>("overview");
  const [proPrice, setProPrice] = useState(platformSettings.tierPrices.pro);
  const [entPrice, setEntPrice] = useState(platformSettings.tierPrices.enterprise);
  const [maintenance, setMaintenance] = useState(platformSettings.maintenanceMode);
  const [saved, setSaved] = useState(false);

  const pendingCount = myInstitution?.status === "pending" ? 1 : 0;
  // + собственная заявка пользователя, если она уже не в статичном сид-массиве (новая регистрация)
  const ownExtra = myInstitution && !INSTITUTIONS.some((i) => i.id === myInstitution.id) ? 1 : 0;
  const totalInstitutions = INSTITUTIONS.length + ownExtra;

  function saveSettings(e: React.FormEvent) {
    e.preventDefault();
    setPlatformSettings({ tierPrices: { pro: proPrice, enterprise: entPrice }, maintenanceMode: maintenance });
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  }

  return (
    <div style={{ display: "flex", fontFamily: FB, background: C.bg, color: C.text, minHeight: "calc(100vh - 64px)" }}>
      <aside style={{ width: 240, flexShrink: 0, borderRight: `1px solid ${C.border}`, padding: "20px 14px" }}>
        <p style={{ padding: "0 10px", fontSize: 11, fontWeight: 700, color: C.dim, textTransform: "uppercase", letterSpacing: ".06em", marginBottom: 8 }}>
          {t({ ru: "Администратор платформы", tg: "Маъмури платформа" })}
        </p>
        <nav style={{ display: "flex", flexDirection: "column", gap: 2 }}>
          <button onClick={() => setTab("overview")} style={{ display: "flex", alignItems: "center", gap: 10, padding: "10px 12px", borderRadius: 10, fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: tab === "overview" ? C.teal : C.sub, background: tab === "overview" ? `${C.teal}18` : "transparent", border: "none", cursor: "pointer", textAlign: "left" }}>
            <LayoutDashboard size={16} /> {t({ ru: "Обзор", tg: "Хулоса" })}
          </button>
          <button onClick={() => setTab("settings")} style={{ display: "flex", alignItems: "center", gap: 10, padding: "10px 12px", borderRadius: 10, fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: tab === "settings" ? C.teal : C.sub, background: tab === "settings" ? `${C.teal}18` : "transparent", border: "none", cursor: "pointer", textAlign: "left" }}>
            <Settings size={16} /> {t({ ru: "Настройки платформы", tg: "Танзимоти платформа" })}
          </button>
        </nav>
      </aside>

      <div style={{ flex: 1, padding: "28px 32px 80px" }}>
        {tab === "overview" && (
          <div>
            <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24, marginBottom: 22 }}>{t({ ru: "Обзор платформы", tg: "Хулосаи платформа" })}</h1>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(4,1fr)", gap: 14, marginBottom: 24 }}>
              {[
                { l: t({ ru: "Учреждений", tg: "Муассисаҳо" }), v: totalInstitutions, icon: Building2 },
                { l: t("common.reviews"), v: REVIEWS.length, icon: Star },
                { l: t({ ru: "Персонала", tg: "Кормандон" }), v: ALL_STAFF.length, icon: Users },
                { l: t({ ru: "Вакансий", tg: "Ҷойҳои холӣ" }), v: VACANCIES.length, icon: Briefcase },
              ].map((s) => (
                <div key={s.l} style={{ borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: 18 }}>
                  <s.icon size={16} style={{ color: C.teal, marginBottom: 8 }} />
                  <p style={{ fontFamily: FH, fontWeight: 900, fontSize: 26, color: C.text }}>{s.v}</p>
                  <p style={{ fontSize: 12.5, color: C.sub, marginTop: 2 }}>{s.l}</p>
                </div>
              ))}
            </div>

            <Link href="/moderator" style={{ display: "flex", alignItems: "center", justifyContent: "space-between", borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: 18, textDecoration: "none" }}>
              <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                <div style={{ width: 38, height: 38, borderRadius: 10, background: pendingCount ? `${C.gold}18` : `${C.ok}18`, display: "flex", alignItems: "center", justifyContent: "center" }}>
                  <ShieldAlert size={18} style={{ color: pendingCount ? C.gold : C.ok }} />
                </div>
                <div>
                  <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text }}>
                    {pendingCount ? t({ ru: "1 заявка на модерации", tg: "1 ариза дар модератсия" }) : t({ ru: "Заявок на модерации нет", tg: "Ариза нест" })}
                  </p>
                  <p style={{ fontSize: 12, color: C.dim }}>{t({ ru: "Перейти к очереди модератора", tg: "Ба навбати модератор гузаред" })}</p>
                </div>
              </div>
              <ArrowRight size={16} style={{ color: C.dim }} />
            </Link>
          </div>
        )}

        {tab === "settings" && (
          <div>
            <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24, marginBottom: 22 }}>{t({ ru: "Настройки платформы", tg: "Танзимоти платформа" })}</h1>
            <form onSubmit={saveSettings} style={{ maxWidth: 420, borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 24 }}>
              <FormField label={t({ ru: "Тариф Pro, $/мес", tg: "Тарифи Pro, $/моҳ" })}>
                <input type="number" min={0} value={proPrice} onChange={(e) => setProPrice(Number(e.target.value))} style={inputStyle} />
              </FormField>
              <FormField label={t({ ru: "Тариф Enterprise, $/мес", tg: "Тарифи Enterprise, $/моҳ" })}>
                <input type="number" min={0} value={entPrice} onChange={(e) => setEntPrice(Number(e.target.value))} style={inputStyle} />
              </FormField>

              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "10px 0", marginBottom: 14 }}>
                <span style={{ fontSize: 13.5, color: C.text, fontFamily: FH, fontWeight: 600 }}>{t({ ru: "Технические работы (баннер на сайте)", tg: "Корҳои техникӣ (баннер дар сомона)" })}</span>
                <button type="button" onClick={() => setMaintenance((v) => !v)} style={{ width: 40, height: 22, borderRadius: 999, position: "relative", background: maintenance ? C.red : C.s3, border: "none", cursor: "pointer", flexShrink: 0 }}>
                  <span style={{ position: "absolute", top: 3, width: 16, height: 16, borderRadius: "50%", background: "#fff", left: maintenance ? 21 : 3, transition: "left .15s" }} />
                </button>
              </div>

              <button type="submit" style={{ width: "100%", padding: 12, borderRadius: 10, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                {saved ? t({ ru: "Сохранено ✓", tg: "Нигоҳ дошта шуд ✓" }) : t({ ru: "Сохранить", tg: "Нигоҳ доштан" })}
              </button>
            </form>
          </div>
        )}
      </div>
    </div>
  );
}
