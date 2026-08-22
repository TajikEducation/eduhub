"use client";

import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Award, Briefcase, GraduationCap, Mail, MessageSquare, Medal, Phone, Send, Trophy, type LucideIcon } from "lucide-react";
import { C, FH, FB, SEED_APPLICANTS } from "@/lib/data";
import { useAppState } from "@/lib/app-state";
import { chatHref } from "@/lib/chat-window";
import { useLocale, useT } from "@/lib/i18n";

const ACH_TIER: Record<string, { icon: LucideIcon; color: string }> = {
  gold: { icon: Medal, color: C.gold }, silver: { icon: Medal, color: C.sub }, bronze: { icon: Medal, color: C.goldD }, special: { icon: Trophy, color: C.teal },
};
const DASH_INST_ID = 1;

export default function ApplicantProfilePage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const { applicant, role, myInstitution, hasResponded, addEmployerResponse } = useAppState();
  const { locale } = useLocale();
  const t = useT();

  // сначала база кандидатов, иначе — собственный профиль текущего пользователя;
  // "по отклику" тоже открывается по прямой ссылке (её присылают из чата/вакансии) — скрыт только draft
  const seedFound = SEED_APPLICANTS.find((a) => a.id === params.id);
  const found = seedFound ?? (applicant.id === params.id && applicant.visibility !== "draft" ? applicant : null);
  const instId = myInstitution?.id ?? DASH_INST_ID;
  const responded = found ? hasResponded(found.id, instId) : false;

  if (!found) {
    return (
      <div style={{ padding: 40, color: C.muted, textAlign: "center" }}>
        {t({ ru: "Профиль не найден или ещё не опубликован", tg: "Профил ёфт нашуд ё ҳанӯз нашр нашудааст" })}
      </div>
    );
  }

  return (
    <div style={{ maxWidth: 900, margin: "0 auto", padding: "28px 28px 80px", fontFamily: FB }}>
      <button onClick={() => router.push("/applicants")} style={{ display: "inline-flex", alignItems: "center", gap: 7, fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: C.teal, marginBottom: 24, padding: "8px 14px", borderRadius: 9, border: `1px solid ${C.teal}40`, background: `${C.teal}10`, cursor: "pointer" }}>
        <ArrowLeft size={15} /> {t("common.back")}
      </button>

      <div style={{ display: "grid", gridTemplateColumns: "300px 1fr", gap: 24, alignItems: "start" }}>
        {/* LEFT CARD */}
        <div style={{ borderRadius: 18, overflow: "hidden", border: `1px solid ${C.border}`, background: C.s1 }}>
          <div style={{ height: 280, overflow: "hidden", position: "relative" }}>
            <img src={found.photo} alt={found.name[locale]} style={{ width: "100%", height: "100%", objectFit: "cover" }} />
            <div style={{ position: "absolute", inset: 0, background: `linear-gradient(180deg,transparent 50%,${C.overlay}F2 100%)` }} />
            <div style={{ position: "absolute", bottom: 16, left: 16, right: 16 }}>
              <p style={{ fontFamily: FH, fontWeight: 900, fontSize: 20, color: "#fff", lineHeight: 1.2 }}>{found.name[locale]}</p>
              <p style={{ fontSize: 13.5, color: C.teal, fontFamily: FH, marginTop: 3 }}>{found.position[locale]}</p>
            </div>
          </div>
          <div style={{ padding: 18, display: "flex", flexDirection: "column", gap: 8 }}>
            {found.hideContacts ? (
              <p style={{ fontSize: 12.5, color: C.dim, lineHeight: 1.5 }}>
                {t({ ru: "Контакты скрыты — напишите в чат, соискатель ответит сам.", tg: "Тамос пинҳон аст — дар чат нависед, довталаб худаш ҷавоб медиҳад." })}
              </p>
            ) : (
              <>
                {found.email && (
                  <div style={{ display: "flex", alignItems: "center", gap: 8, fontSize: 13, color: C.sub }}>
                    <Mail size={14} style={{ color: C.teal, flexShrink: 0 }} />{found.email}
                  </div>
                )}
                {found.phone && (
                  <div style={{ display: "flex", alignItems: "center", gap: 8, fontSize: 13, color: C.sub }}>
                    <Phone size={14} style={{ color: C.teal, flexShrink: 0 }} />{found.phone}
                  </div>
                )}
              </>
            )}
            <button onClick={() => router.push(chatHref())} style={{ marginTop: 8, width: "100%", padding: "10px", borderRadius: 12, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 700, fontSize: 13.5, display: "flex", alignItems: "center", justifyContent: "center", gap: 7, cursor: "pointer", border: "none" }}>
              <MessageSquare size={14} /> {t("common.write")}
            </button>
            {role === "institution" && (
              <button
                onClick={() => { if (!responded) addEmployerResponse(found.id, instId, t({ ru: "Мы заинтересованы в вашем профиле — напишите нам в чат.", tg: "Мо ба профили шумо шавқмандем — дар чат ба мо нависед." })); }}
                disabled={responded}
                style={{ width: "100%", padding: "10px", borderRadius: 12, background: responded ? C.s2 : `${C.teal}18`, border: `1px solid ${responded ? C.border : C.teal + "44"}`, color: responded ? C.dim : C.teal, fontFamily: FH, fontWeight: 700, fontSize: 13.5, display: "flex", alignItems: "center", justifyContent: "center", gap: 7, cursor: responded ? "default" : "pointer" }}>
                <Send size={14} /> {responded ? t({ ru: "Вы откликнулись", tg: "Шумо ҷавоб додед" }) : t({ ru: "Откликнуться", tg: "Ҷавоб додан" })}
              </button>
            )}
          </div>
        </div>

        {/* RIGHT */}
        <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
          <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 24 }}>
            <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text, marginBottom: 12 }}>{t("applicant.about")}</h2>
            <p style={{ fontSize: 14.5, color: C.sub, lineHeight: 1.75 }}>{found.bio[locale]}</p>
            {found.skills.length > 0 && (
              <div style={{ display: "flex", gap: 8, flexWrap: "wrap", marginTop: 14 }}>
                {found.skills.map((s, i) => (
                  <span key={i} style={{ fontSize: 12, fontWeight: 600, padding: "5px 11px", borderRadius: 8, background: `${C.teal}18`, color: C.teal, fontFamily: FH }}>{s[locale]}</span>
                ))}
              </div>
            )}
          </div>

          {found.education.length > 0 && (
            <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 24 }}>
              <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text, marginBottom: 16, display: "flex", alignItems: "center", gap: 8 }}>
                <GraduationCap size={18} style={{ color: C.teal }} /> {t("applicant.education")}
              </h2>
              <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                {found.education.map((ed, i) => (
                  <div key={i} style={{ display: "flex", gap: 14, alignItems: "flex-start" }}>
                    <div style={{ width: 8, height: 8, borderRadius: "50%", background: C.teal, marginTop: 6, flexShrink: 0 }} />
                    <p style={{ fontSize: 14, color: C.sub, lineHeight: 1.6 }}>{ed[locale]}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {found.experience.length > 0 && (
            <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 24 }}>
              <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text, marginBottom: 16, display: "flex", alignItems: "center", gap: 8 }}>
                <Briefcase size={18} style={{ color: C.teal }} /> {t("applicant.experience")}
              </h2>
              <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                {found.experience.map((ex, i) => (
                  <div key={i} style={{ display: "flex", gap: 14, alignItems: "flex-start" }}>
                    <div style={{ width: 8, height: 8, borderRadius: "50%", background: C.teal, marginTop: 6, flexShrink: 0 }} />
                    <p style={{ fontSize: 14, color: C.sub, lineHeight: 1.6 }}>{ex[locale]}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {found.achievements.length > 0 && (
            <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 24 }}>
              <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text, marginBottom: 16, display: "flex", alignItems: "center", gap: 8 }}>
                <Award size={18} style={{ color: C.gold }} /> {t("tab.achievements")}
              </h2>
              <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
                {found.achievements.map((ach) => {
                  const tier = ACH_TIER[ach.category] ?? ACH_TIER.special;
                  const TierIcon = tier.icon;
                  return (
                    <div key={ach.id} style={{ display: "flex", gap: 12, alignItems: "flex-start", padding: "12px 14px", borderRadius: 12, background: C.s2 }}>
                      <TierIcon size={20} style={{ color: tier.color, flexShrink: 0, marginTop: 1 }} />
                      <div>
                        <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text }}>{ach.title[locale]}</p>
                        <p style={{ fontSize: 12.5, color: C.sub }}>{ach.year} · {ach.desc[locale]}</p>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
