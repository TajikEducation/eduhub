"use client";

import { useRouter, useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Wallet, Clock, CheckCircle, MessageSquare, Building2, ChevronRight } from "lucide-react";
import { C, FH, CATEGORY_META } from "@/lib/data";
import { useT } from "@/lib/i18n";
import { useGetVacancyQuery } from "../api/vacanciesApi";
import { backendTypeToCategory } from "@/lib/backendTypes";
import { useCreateConversationMutation } from "@/app/(site)/messages/api/chatApi";
import { useMeQuery } from "@/app/(site)/login/api/authApi";
import { getAccessToken } from "@/lib/authToken";
import { useApplyToVacancyMutation, useListMyApplicationsQuery } from "@/app/(site)/applicants/api/applicantsApi";
import { toast } from "sonner";

export default function VacancyDetailPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const t = useT();
  const { data: me } = useMeQuery(undefined, { skip: !getAccessToken() });
  const [createConversation] = useCreateConversationMutation();
  const [applyToVacancy, { isLoading: applying }] = useApplyToVacancyMutation();
  const { data: myAppsData } = useListMyApplicationsQuery(undefined, { skip: !me });
  const applied = myAppsData?.vacancy_ids.includes(params.id) ?? false;

  const { data: vacancy, isLoading, isError } = useGetVacancyQuery(params.id);

  if (isLoading) {
    return <div style={{ maxWidth: 720, margin: "0 auto", padding: "80px 28px", textAlign: "center", color: C.muted }}>{t({ ru: "Загрузка…", tg: "Боркунӣ…" })}</div>;
  }
  if (isError || !vacancy) {
    return (
      <div style={{ maxWidth: 720, margin: "0 auto", padding: "80px 28px", textAlign: "center" }}>
        <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text }}>{t({ ru: "Вакансия не найдена", tg: "Ҷои холӣ ёфт нашуд" })}</p>
        <Link href="/vacancies" style={{ color: C.teal, fontFamily: FH, fontWeight: 700, fontSize: 14, marginTop: 12, display: "inline-block" }}>
          {t({ ru: "← Все вакансии", tg: "← Ҳамаи ҷойҳои холӣ" })}
        </Link>
      </div>
    );
  }

  const inst = vacancy.institution;
  const meta = CATEGORY_META[backendTypeToCategory(inst.types[0])];
  const Icon = meta.icon;

  return (
    <div style={{ maxWidth: 780, margin: "0 auto", padding: "28px 28px 80px" }}>
      <button onClick={() => router.back()} style={{ display: "flex", alignItems: "center", gap: 6, background: "none", border: "none", color: C.sub, fontFamily: FH, fontWeight: 600, fontSize: 13.5, cursor: "pointer", marginBottom: 20, padding: 0 }}>
        <ArrowLeft size={16} /> {t({ ru: "Назад", tg: "Бозгашт" })}
      </button>

      <Link href={`/institutions/${inst.id}`} style={{ display: "flex", alignItems: "center", gap: 12, borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: 12, textDecoration: "none", marginBottom: 24 }}>
        <div style={{ width: 48, height: 48, borderRadius: 10, overflow: "hidden", position: "relative", flexShrink: 0, background: meta.color }}>
          {inst.cover_photo_s3_key && <img src={inst.cover_photo_s3_key} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />}
        </div>
        <div style={{ flex: 1, minWidth: 0 }}>
          <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14.5, color: C.text }}>{t(inst.name)}</p>
          <p style={{ fontSize: 12.5, color: C.sub, display: "flex", alignItems: "center", gap: 5, marginTop: 2 }}>
            <Icon size={12} /> {t(meta.label)}{inst.district ? ` · ${inst.district}` : ""}
          </p>
        </div>
        <ChevronRight size={16} style={{ color: C.dim, flexShrink: 0 }} />
      </Link>

      <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 28 }}>
        <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(20px,2.6vw,26px)", color: C.text, marginBottom: 12, letterSpacing: "-.02em" }}>
          {t(vacancy.title)}
        </h1>
        <div style={{ display: "flex", gap: 16, flexWrap: "wrap", marginBottom: 20 }}>
          {vacancy.salary_from && (
            <span style={{ display: "flex", alignItems: "center", gap: 6, fontSize: 13.5, color: C.text }}>
              <Wallet size={14} style={{ color: C.teal }} /> {vacancy.salary_from}–{vacancy.salary_to} {t("common.perMonth")}
            </span>
          )}
          <span style={{ display: "flex", alignItems: "center", gap: 6, fontSize: 13.5, color: C.text }}>
            <Clock size={14} style={{ color: C.teal }} /> {t(vacancy.employment)}
          </span>
        </div>
        <p style={{ fontSize: 14.5, color: C.sub, lineHeight: 1.75, marginBottom: 20 }}>{t(vacancy.description)}</p>
        {vacancy.requirements && vacancy.requirements.length > 0 && (
          <div style={{ display: "flex", flexDirection: "column", gap: 8, marginBottom: 24 }}>
            {vacancy.requirements.map((r, i) => (
              <div key={i} style={{ display: "flex", alignItems: "flex-start", gap: 8, fontSize: 13.5, color: C.sub }}>
                <CheckCircle size={14} style={{ color: C.ok, flexShrink: 0, marginTop: 2 }} /> {t(r)}
              </div>
            ))}
          </div>
        )}
        <div style={{ display: "flex", gap: 10, flexWrap: "wrap" }}>
          <button
            disabled={applying}
            onClick={async () => {
              if (!me) { router.push("/login"); return; }
              if (!applied) {
                try {
                  await applyToVacancy(vacancy.id).unwrap();
                } catch {
                  toast.error(t({ ru: "Не удалось откликнуться — сначала создайте резюме в личном кабинете", tg: "Ҷавоб дода нашуд — аввал дар кабинети шахсӣ резюме созед" }));
                  router.push("/account");
                  return;
                }
              }
              try {
                const conv = await createConversation({ counterpart_type: "institution", counterpart_id: inst.id }).unwrap();
                router.push(`/messages?conv=${conv.id}`);
              } catch {
                toast.error(t({ ru: "Не удалось открыть чат", tg: "Чатро кушода натавонист" }));
              }
            }}
            style={{ display: "flex", alignItems: "center", gap: 8, padding: "11px 22px", borderRadius: 10, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 700, fontSize: 13.5, border: "none", cursor: "pointer", opacity: applying ? 0.6 : 1 }}>
            <MessageSquare size={14} /> {applied ? t({ ru: "Вы откликнулись — открыть чат", tg: "Шумо ҷавоб додаед — чатро кушоед" }) : t({ ru: "Откликнуться", tg: "Ҷавоб додан" })}
          </button>
          <Link href={`/institutions/${inst.id}`}
            style={{ display: "flex", alignItems: "center", gap: 8, padding: "11px 22px", borderRadius: 10, background: "transparent", border: `1px solid ${C.border}`, color: C.text, fontFamily: FH, fontWeight: 700, fontSize: 13.5, textDecoration: "none" }}>
            <Building2 size={14} /> {t({ ru: "Перейти на профиль учреждения", tg: "Ба профили муассиса гузаред" })}
          </Link>
        </div>
      </div>
    </div>
  );
}
