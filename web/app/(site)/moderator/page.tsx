"use client";

import { useState } from "react";
import { CheckCircle, XCircle, ShieldCheck } from "lucide-react";
import { C, FH, FB, CATEGORY_META, REGION_LABEL } from "@/lib/data";
import { backendTypeToCategory } from "@/lib/backendTypes";
import { useT } from "@/lib/i18n";
import { useMeQuery } from "@/app/(site)/login/api/authApi";
import { getAccessToken } from "@/lib/authToken";
import {
  useListModerationInstitutionsQuery,
  useApproveInstitutionMutation,
  useRejectInstitutionMutation,
  type ModerationInstitution,
} from "./api/moderationApi";
import { toast } from "sonner";

function PendingCard({ inst }: { inst: ModerationInstitution }) {
  const t = useT();
  const [approve, { isLoading: approving }] = useApproveInstitutionMutation();
  const [reject, { isLoading: rejecting }] = useRejectInstitutionMutation();
  const [showReject, setShowReject] = useState(false);
  const [reasonText, setReasonText] = useState("");
  const meta = CATEGORY_META[backendTypeToCategory(inst.types[0])];

  async function doApprove() {
    try {
      await approve(inst.id).unwrap();
      toast.success(t({ ru: "Учреждение одобрено", tg: "Муассиса тасдиқ шуд" }));
    } catch {
      toast.error(t({ ru: "Не удалось одобрить", tg: "Тасдиқ нашуд" }));
    }
  }
  async function doReject() {
    if (!reasonText.trim()) return;
    try {
      await reject({ id: inst.id, reason_code: "other", reason_text: reasonText.trim() }).unwrap();
      toast.success(t({ ru: "Учреждение отклонено", tg: "Муассиса рад шуд" }));
      setShowReject(false); setReasonText("");
    } catch {
      toast.error(t({ ru: "Не удалось отклонить", tg: "Рад карда нашуд" }));
    }
  }

  return (
    <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 26 }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 18, gap: 12 }}>
        <div>
          <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text, marginBottom: 4 }}>{inst.name.ru}</h2>
          <p style={{ fontSize: 13, color: C.sub }}>{t(meta.label)} · {t(REGION_LABEL[inst.region as keyof typeof REGION_LABEL] ?? { ru: inst.region, tg: inst.region })}</p>
        </div>
        <span style={{ fontSize: 11.5, fontWeight: 700, padding: "4px 10px", borderRadius: 7, background: `${C.gold}22`, border: `1px solid ${C.gold}`, color: C.gold, fontFamily: FH, flexShrink: 0 }}>
          {t({ ru: "На рассмотрении", tg: "Дар баррасӣ" })}
        </span>
      </div>

      <div className="eh-mobile-1col" style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12, marginBottom: 18 }}>
        {[
          { l: t({ ru: "Адрес", tg: "Суроға" }), v: inst.address?.ru ?? "—" },
          { l: t({ ru: "Телефон", tg: "Телефон" }), v: inst.phone ?? "—" },
          { l: "Email", v: inst.email ?? "—" },
          { l: t({ ru: "Цена", tg: "Нарх" }), v: inst.price != null ? `${inst.price} ${t("common.perMonth")}` : "—" },
        ].map((f) => (
          <div key={f.l}>
            <p style={{ fontSize: 11, fontWeight: 700, color: C.dim, textTransform: "uppercase", letterSpacing: ".05em", fontFamily: FH, marginBottom: 3 }}>{f.l}</p>
            <p style={{ fontSize: 13.5, color: C.text }}>{f.v}</p>
          </div>
        ))}
      </div>

      {inst.description && <p style={{ fontSize: 13.5, color: C.sub, lineHeight: 1.6, marginBottom: 22 }}>{inst.description.ru}</p>}

      {!showReject ? (
        <div style={{ display: "flex", gap: 10 }}>
          <button onClick={doApprove} disabled={approving} style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", gap: 7, padding: 12, borderRadius: 11, background: C.ok, color: "#fff", fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer", opacity: approving ? 0.6 : 1 }}>
            <CheckCircle size={16} /> {t({ ru: "Одобрить", tg: "Тасдиқ" })}
          </button>
          <button onClick={() => setShowReject(true)} style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", gap: 7, padding: 12, borderRadius: 11, background: `${C.red}14`, border: `1px solid ${C.red}44`, color: C.red, fontFamily: FH, fontWeight: 800, fontSize: 13.5, cursor: "pointer" }}>
            <XCircle size={16} /> {t({ ru: "Отклонить", tg: "Рад кардан" })}
          </button>
        </div>
      ) : (
        <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          <textarea value={reasonText} onChange={(e) => setReasonText(e.target.value)} rows={2}
            placeholder={t({ ru: "Причина отклонения", tg: "Сабаби рад кардан" })}
            style={{ padding: "10px 14px", borderRadius: 10, border: `1px solid ${C.border}`, background: C.s2, color: C.text, fontFamily: FB, fontSize: 14, outline: "none", resize: "vertical" }} />
          <div style={{ display: "flex", gap: 10 }}>
            <button onClick={doReject} disabled={!reasonText.trim() || rejecting} style={{ flex: 1, padding: 11, borderRadius: 10, background: C.red, color: "#fff", fontFamily: FH, fontWeight: 700, fontSize: 13.5, border: "none", cursor: "pointer", opacity: (!reasonText.trim() || rejecting) ? 0.5 : 1 }}>
              {t({ ru: "Подтвердить отклонение", tg: "Тасдиқи рад" })}
            </button>
            <button onClick={() => setShowReject(false)} style={{ padding: "11px 18px", borderRadius: 10, background: C.s3, color: C.sub, fontFamily: FH, fontWeight: 700, fontSize: 13.5, border: "none", cursor: "pointer" }}>
              {t({ ru: "Отмена", tg: "Бекор" })}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

export default function ModeratorPage() {
  const t = useT();
  const { data: me, isLoading: meLoading } = useMeQuery(undefined, { skip: !getAccessToken() });
  const isModerator = me?.role === "moderator" || me?.role === "admin";
  const { data, isLoading, isError } = useListModerationInstitutionsQuery({ status: "pending" }, { skip: !isModerator });
  const pending = data?.items ?? [];

  if (meLoading) {
    return <div style={{ padding: 60, textAlign: "center", color: C.muted, fontFamily: FB }}>{t({ ru: "Загрузка…", tg: "Боркунӣ…" })}</div>;
  }

  if (!isModerator) {
    return (
      <div style={{ maxWidth: 480, margin: "0 auto", padding: "60px 28px", textAlign: "center", fontFamily: FB }}>
        <ShieldCheck size={26} style={{ color: C.dim, marginBottom: 12 }} />
        <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 17, color: C.text, marginBottom: 8 }}>
          {t({ ru: "Доступ только для модераторов", tg: "Дастрасӣ танҳо барои модераторон" })}
        </p>
        <p style={{ fontSize: 13.5, color: C.muted }}>
          {t({ ru: "Войдите в аккаунт с ролью модератора или администратора.", tg: "Бо аккаунти дорои нақши модератор ё маъмур ворид шавед." })}
        </p>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: 720, margin: "0 auto", padding: "40px 28px 80px", fontFamily: FB }}>
      <div style={{ textAlign: "center", marginBottom: 28 }}>
        <ShieldCheck size={26} style={{ color: C.teal, marginBottom: 10 }} />
        <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(22px,3vw,28px)", color: C.text, marginBottom: 8 }}>
          {t({ ru: "Модерация заявок", tg: "Модератсияи аризаҳо" })}
        </h1>
        <p style={{ fontSize: 13, color: C.dim, maxWidth: 480, margin: "0 auto" }}>
          {t({ ru: "Очередь заявок на регистрацию учреждений, ожидающих модерации.", tg: "Навбати аризаҳои сабти муассисаҳо, ки мунтазири модератсиянд." })}
        </p>
      </div>

      {isLoading && <p style={{ color: C.muted, fontSize: 14, textAlign: "center" }}>{t({ ru: "Загрузка…", tg: "Боркунӣ…" })}</p>}
      {isError && <p style={{ color: C.red, fontSize: 14, textAlign: "center" }}>{t({ ru: "Backend недоступен", tg: "Backend дастнорас аст" })}</p>}

      {!isLoading && pending.length === 0 && (
        <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 40, textAlign: "center" }}>
          <p style={{ fontSize: 14.5, color: C.sub }}>{t({ ru: "Нет заявок на рассмотрении", tg: "Аризае барои баррасӣ нест" })}</p>
        </div>
      )}

      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        {pending.map((inst) => (
          <PendingCard key={inst.id} inst={inst} />
        ))}
      </div>
    </div>
  );
}
