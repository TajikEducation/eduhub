"use client";

import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Award, GraduationCap, Mail, MessageSquare, Medal, Phone, Trophy, type LucideIcon } from "lucide-react";
import { C, FH, FB } from "@/lib/data";
import { useT } from "@/lib/i18n";
import { useGetApplicantQuery, useListApplicantAchievementsQuery } from "../api/applicantsApi";
import { useCreateConversationMutation } from "@/app/(site)/messages/api/chatApi";
import { useMeQuery } from "@/app/(site)/login/api/authApi";
import { getAccessToken } from "@/lib/authToken";
import { toast } from "sonner";

const ACH_TIER: Record<string, { icon: LucideIcon; color: string }> = {
  gold: { icon: Medal, color: C.gold }, silver: { icon: Medal, color: C.sub }, bronze: { icon: Medal, color: C.goldD }, special: { icon: Trophy, color: C.teal },
};

export default function ApplicantProfilePage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const t = useT();
  const { data: found, isLoading, isError } = useGetApplicantQuery(params.id);
  const { data: achData } = useListApplicantAchievementsQuery(params.id);
  const achievements = achData?.items ?? [];
  const { data: me } = useMeQuery(undefined, { skip: !getAccessToken() });
  const [createConversation] = useCreateConversationMutation();

  async function startChat() {
    if (!found) return;
    if (!me) { router.push("/login"); return; }
    try {
      const conv = await createConversation({ counterpart_type: "user", counterpart_id: found.user_id }).unwrap();
      router.push(`/messages?conv=${conv.id}`);
    } catch {
      toast.error(t({ ru: "Не удалось открыть чат", tg: "Чатро кушода натавонист" }));
    }
  }

  if (isLoading) {
    return <div style={{ padding: 40, color: C.muted, textAlign: "center" }}>{t({ ru: "Загрузка…", tg: "Боркунӣ…" })}</div>;
  }
  if (isError || !found) {
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

      <div className="eh-mobile-1col" style={{ display: "grid", gridTemplateColumns: "300px 1fr", gap: 24, alignItems: "start" }}>
        {/* LEFT CARD */}
        <div style={{ borderRadius: 18, overflow: "hidden", border: `1px solid ${C.border}`, background: C.s1 }}>
          <div style={{ height: 280, overflow: "hidden", position: "relative" }}>
            <img src={found.photo_url || "/logo.svg"} alt={found.name.ru} style={{ width: "100%", height: "100%", objectFit: "cover" }} />
            <div style={{ position: "absolute", inset: 0, background: `linear-gradient(180deg,transparent 50%,${C.overlay}F2 100%)` }} />
            <div style={{ position: "absolute", bottom: 16, left: 16, right: 16 }}>
              <p style={{ fontFamily: FH, fontWeight: 900, fontSize: 20, color: "#fff", lineHeight: 1.2 }}>{found.name.ru}</p>
              <p style={{ fontSize: 13.5, color: C.teal, fontFamily: FH, marginTop: 3 }}>{found.position.ru}</p>
            </div>
          </div>
          <div style={{ padding: 18, display: "flex", flexDirection: "column", gap: 8 }}>
            {found.hide_contacts && found.visibility !== "public" ? (
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
            <button onClick={startChat} style={{ marginTop: 8, width: "100%", padding: "10px", borderRadius: 12, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 700, fontSize: 13.5, display: "flex", alignItems: "center", justifyContent: "center", gap: 7, cursor: "pointer", border: "none" }}>
              <MessageSquare size={14} /> {t("common.write")}
            </button>
          </div>
        </div>

        {/* RIGHT */}
        <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
          <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 24 }}>
            <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text, marginBottom: 12 }}>{t("applicant.about")}</h2>
            {found.bio && <p style={{ fontSize: 14.5, color: C.sub, lineHeight: 1.75 }}>{found.bio.ru}</p>}
            {found.skills && found.skills.length > 0 && (
              <div style={{ display: "flex", gap: 8, flexWrap: "wrap", marginTop: 14 }}>
                {found.skills.map((s, i) => (
                  <span key={i} style={{ fontSize: 12, fontWeight: 600, padding: "5px 11px", borderRadius: 8, background: `${C.teal}18`, color: C.teal, fontFamily: FH }}>{s.ru}</span>
                ))}
              </div>
            )}
          </div>

          {found.education && found.education.length > 0 && (
            <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 24 }}>
              <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text, marginBottom: 16, display: "flex", alignItems: "center", gap: 8 }}>
                <GraduationCap size={18} style={{ color: C.teal }} /> {t("applicant.education")}
              </h2>
              <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                {found.education.map((ed, i) => (
                  <div key={i} style={{ display: "flex", gap: 14, alignItems: "flex-start" }}>
                    <div style={{ width: 8, height: 8, borderRadius: "50%", background: C.teal, marginTop: 6, flexShrink: 0 }} />
                    <p style={{ fontSize: 14, color: C.sub, lineHeight: 1.6 }}>{ed.ru}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {found.experience && found.experience.length > 0 && (
            <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 24 }}>
              <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text, marginBottom: 16, display: "flex", alignItems: "center", gap: 8 }}>
                <GraduationCap size={18} style={{ color: C.teal }} /> {t("applicant.experience")}
              </h2>
              <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                {found.experience.map((ex, i) => (
                  <div key={i} style={{ display: "flex", gap: 14, alignItems: "flex-start" }}>
                    <div style={{ width: 8, height: 8, borderRadius: "50%", background: C.teal, marginTop: 6, flexShrink: 0 }} />
                    <p style={{ fontSize: 14, color: C.sub, lineHeight: 1.6 }}>{ex.ru}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {achievements.length > 0 && (
            <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 24 }}>
              <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text, marginBottom: 16, display: "flex", alignItems: "center", gap: 8 }}>
                <Award size={18} style={{ color: C.gold }} /> {t("tab.achievements")}
              </h2>
              <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
                {achievements.map((ach) => {
                  const tier = ACH_TIER[ach.category ?? "special"] ?? ACH_TIER.special;
                  const TierIcon = tier.icon;
                  return (
                    <div key={ach.id} style={{ display: "flex", gap: 12, alignItems: "flex-start", padding: "12px 14px", borderRadius: 12, background: C.s2 }}>
                      <TierIcon size={20} style={{ color: tier.color, flexShrink: 0, marginTop: 1 }} />
                      <div>
                        <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text }}>{ach.title}</p>
                        <p style={{ fontSize: 12.5, color: C.sub }}>{ach.year ?? ""} {ach.description ? `· ${ach.description}` : ""}</p>
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
