"use client";

import { useRouter, useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Wallet, Clock, CheckCircle, MessageSquare, Building2, ChevronRight } from "lucide-react";
import { C, FH, VACANCIES, INSTITUTIONS, CATEGORY_META } from "@/lib/data";
import { chatHref } from "@/lib/chat-window";
import { useAppState } from "@/lib/app-state";
import { useT } from "@/lib/i18n";

export default function VacancyDetailPage() {
  const router = useRouter();
  const params = useParams();
  const t = useT();
  const { hasApplied, addApplication } = useAppState();

  const vacancy = VACANCIES.find(v => v.id === params.id);
  const inst = vacancy ? INSTITUTIONS.find(i => i.id === vacancy.instId) : undefined;

  if (!vacancy) {
    return (
      <div style={{ maxWidth: 720, margin: "0 auto", padding: "80px 28px", textAlign: "center" }}>
        <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text }}>{t({ ru: "Вакансия не найдена", tg: "Ҷои холӣ ёфт нашуд" })}</p>
        <Link href="/vacancies" style={{ color: C.teal, fontFamily: FH, fontWeight: 700, fontSize: 14, marginTop: 12, display: "inline-block" }}>
          {t({ ru: "← Все вакансии", tg: "← Ҳамаи ҷойҳои холӣ" })}
        </Link>
      </div>
    );
  }

  const meta = inst ? CATEGORY_META[inst.tk] : null;
  const Icon = meta?.icon;

  return (
    <div style={{ maxWidth: 780, margin: "0 auto", padding: "28px 28px 80px" }}>
      <button onClick={() => router.back()} style={{ display: "flex", alignItems: "center", gap: 6, background: "none", border: "none", color: C.sub, fontFamily: FH, fontWeight: 600, fontSize: 13.5, cursor: "pointer", marginBottom: 20, padding: 0 }}>
        <ArrowLeft size={16} /> {t({ ru: "Назад", tg: "Бозгашт" })}
      </button>

      {inst && (
        <Link href={`/institutions/${inst.id}`} style={{ display: "flex", alignItems: "center", gap: 12, borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: 12, textDecoration: "none", marginBottom: 24 }}>
          <div style={{ width: 48, height: 48, borderRadius: 10, overflow: "hidden", position: "relative", flexShrink: 0, background: inst.color }}>
            {inst.coverPhoto && <img src={inst.coverPhoto} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />}
          </div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14.5, color: C.text }}>{t(inst.name)}</p>
            <p style={{ fontSize: 12.5, color: C.sub, display: "flex", alignItems: "center", gap: 5, marginTop: 2 }}>
              {Icon && <Icon size={12} />} {t(meta!.label)} · {inst.area}
            </p>
          </div>
          <ChevronRight size={16} style={{ color: C.dim, flexShrink: 0 }} />
        </Link>
      )}

      <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 28 }}>
        <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(20px,2.6vw,26px)", color: C.text, marginBottom: 12, letterSpacing: "-.02em" }}>
          {t(vacancy.title)}
        </h1>
        <div style={{ display: "flex", gap: 16, flexWrap: "wrap", marginBottom: 20 }}>
          {vacancy.salaryFrom && (
            <span style={{ display: "flex", alignItems: "center", gap: 6, fontSize: 13.5, color: C.text }}>
              <Wallet size={14} style={{ color: C.teal }} /> {vacancy.salaryFrom}–{vacancy.salaryTo} {t("common.perMonth")}
            </span>
          )}
          <span style={{ display: "flex", alignItems: "center", gap: 6, fontSize: 13.5, color: C.text }}>
            <Clock size={14} style={{ color: C.teal }} /> {t(vacancy.employment)}
          </span>
        </div>
        <p style={{ fontSize: 14.5, color: C.sub, lineHeight: 1.75, marginBottom: 20 }}>{t(vacancy.description)}</p>
        <div style={{ display: "flex", flexDirection: "column", gap: 8, marginBottom: 24 }}>
          {vacancy.requirements.map((r, i) => (
            <div key={i} style={{ display: "flex", alignItems: "flex-start", gap: 8, fontSize: 13.5, color: C.sub }}>
              <CheckCircle size={14} style={{ color: C.ok, flexShrink: 0, marginTop: 2 }} /> {t(r)}
            </div>
          ))}
        </div>
        <div style={{ display: "flex", gap: 10, flexWrap: "wrap" }}>
          <button
            onClick={() => { addApplication(vacancy.id); router.push(chatHref(vacancy.instId)); }}
            style={{ display: "flex", alignItems: "center", gap: 8, padding: "11px 22px", borderRadius: 10, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 700, fontSize: 13.5, border: "none", cursor: "pointer" }}>
            <MessageSquare size={14} /> {hasApplied(vacancy.id) ? t({ ru: "Вы откликнулись — открыть чат", tg: "Шумо ҷавоб додаед — чатро кушоед" }) : t({ ru: "Откликнуться", tg: "Ҷавоб додан" })}
          </button>
          {inst && (
            <Link href={`/institutions/${inst.id}`}
              style={{ display: "flex", alignItems: "center", gap: 8, padding: "11px 22px", borderRadius: 10, background: "transparent", border: `1px solid ${C.border}`, color: C.text, fontFamily: FH, fontWeight: 700, fontSize: 13.5, textDecoration: "none" }}>
              <Building2 size={14} /> {t({ ru: "Перейти на профиль учреждения", tg: "Ба профили муассиса гузаред" })}
            </Link>
          )}
        </div>
      </div>
    </div>
  );
}
