"use client";

import Link from "next/link";
import { ChevronRight, GraduationCap, UserRound } from "lucide-react";
import { C, FH } from "@/lib/data";
import { useT } from "@/lib/i18n";
import { useListApplicantsQuery } from "./api/applicantsApi";

export default function ApplicantsPage() {
  const t = useT();
  const { data, isLoading, isError } = useListApplicantsQuery();
  const published = data?.items ?? [];

  return (
    <div style={{ maxWidth: 1260, margin: "0 auto", padding: "28px 28px 80px" }}>
      <div style={{ display: "flex", gap: 8, marginBottom: 20 }}>
        <Link href="/vacancies" style={{ padding: "8px 16px", borderRadius: 10, background: C.s3, border: `1px solid ${C.border}`, color: C.sub, fontFamily: FH, fontWeight: 700, fontSize: 13, textDecoration: "none" }}>
          {t("nav.vacancies")}
        </Link>
        <span style={{ padding: "8px 16px", borderRadius: 10, background: `${C.teal}18`, border: `1px solid ${C.teal}44`, color: C.teal, fontFamily: FH, fontWeight: 700, fontSize: 13 }}>
          {t("tab.candidates")}
        </span>
      </div>

      <div style={{ marginBottom: 24 }}>
        <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(22px,3vw,30px)", color: C.text, letterSpacing: "-.02em" }}>
          {t("tab.candidates")}
        </h1>
        <p style={{ fontSize: 14, color: C.sub, marginTop: 4 }}>
          {t({ ru: "Опубликованные профили соискателей — педагогов и специалистов", tg: "Профилҳои нашршудаи ҷуяндагони кор — омӯзгорон ва мутахассисон" })}
        </p>
      </div>

      {isLoading && <p style={{ color: C.muted, fontSize: 14 }}>{t({ ru: "Загрузка…", tg: "Боркунӣ…" })}</p>}
      {isError && <p style={{ color: C.red, fontSize: 14 }}>{t({ ru: "Backend недоступен", tg: "Backend дастнорас аст" })}</p>}

      {!isLoading && published.length === 0 ? (
        <div style={{ padding: 56, borderRadius: 16, border: `1px dashed ${C.border}`, textAlign: "center", color: C.muted }}>
          <UserRound size={28} style={{ color: C.dim, margin: "0 auto 12px" }} />
          <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 17, color: C.text }}>{t("empty.candidates")}</p>
        </div>
      ) : (
        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          {published.map((a) => (
            <Link key={a.id} href={`/applicants/${a.id}`} style={{ display: "flex", alignItems: "center", gap: 16, borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: "16px 20px", textDecoration: "none" }}>
              <img src={a.photo_url || "/logo.svg"} alt="" style={{ width: 52, height: 52, borderRadius: "50%", objectFit: "cover", flexShrink: 0 }} />
              <div style={{ flex: 1, minWidth: 0 }}>
                <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 15, color: C.text }}>{a.name.ru}</p>
                <p style={{ fontSize: 13, color: C.teal, marginTop: 2 }}>{a.position.ru}</p>
                {a.education?.[0] && (
                  <p style={{ fontSize: 12.5, color: C.sub, marginTop: 4, display: "flex", alignItems: "center", gap: 5 }}>
                    <GraduationCap size={12} /> {a.education[0].ru}
                  </p>
                )}
              </div>
              <ChevronRight size={18} style={{ color: C.dim, flexShrink: 0 }} />
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
