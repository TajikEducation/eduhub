"use client";

import { CheckCircle, XCircle, ShieldCheck } from "lucide-react";
import { C, FH, FB, CATEGORY_META, REGION_LABEL } from "@/lib/data";
import { useAppState } from "@/lib/app-state";
import { useT } from "@/lib/i18n";

export default function ModeratorPage() {
  const t = useT();
  const { myInstitution, setInstitutionStatus } = useAppState();

  const pending = myInstitution?.status === "pending" ? myInstitution : null;

  return (
    <div style={{ maxWidth: 720, margin: "0 auto", padding: "40px 28px 80px", fontFamily: FB }}>
      <div style={{ textAlign: "center", marginBottom: 28 }}>
        <ShieldCheck size={26} style={{ color: C.teal, marginBottom: 10 }} />
        <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(22px,3vw,28px)", color: C.text, marginBottom: 8 }}>
          {t({ ru: "Модерация заявок", tg: "Модератсияи аризаҳо" })}
        </h1>
        <p style={{ fontSize: 13, color: C.dim, maxWidth: 480, margin: "0 auto" }}>
          {t({ ru: "Показана заявка текущего браузера — в этом прототипе нет сервера и других пользователей, реальная многопользовательская очередь потребует бэкенда.", tg: "Нишон дода мешавад аризаи браузери ҷорӣ — дар ин прототип сервер ва корбарони дигар нест." })}
        </p>
      </div>

      {!pending && (
        <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 40, textAlign: "center" }}>
          <p style={{ fontSize: 14.5, color: C.sub }}>{t({ ru: "Нет заявок на рассмотрении", tg: "Аризае барои баррасӣ нест" })}</p>
        </div>
      )}

      {pending && (
        <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 26 }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 18, gap: 12 }}>
            <div>
              <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text, marginBottom: 4 }}>{pending.name.ru}</h2>
              <p style={{ fontSize: 13, color: C.sub }}>{t(CATEGORY_META[pending.tk].label)} · {t(REGION_LABEL[pending.region])}</p>
            </div>
            <span style={{ fontSize: 11.5, fontWeight: 700, padding: "4px 10px", borderRadius: 7, background: `${C.gold}22`, border: `1px solid ${C.gold}`, color: C.gold, fontFamily: FH, flexShrink: 0 }}>
              {t({ ru: "На рассмотрении", tg: "Дар баррасӣ" })}
            </span>
          </div>

          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12, marginBottom: 18 }}>
            {[
              { l: t({ ru: "Адрес", tg: "Суроға" }), v: pending.address.ru },
              { l: t({ ru: "Телефон", tg: "Телефон" }), v: pending.phone },
              { l: "Email", v: pending.email },
              { l: t({ ru: "Цена", tg: "Нарх" }), v: `${pending.price} ${t("common.perMonth")}` },
            ].map((f) => (
              <div key={f.l}>
                <p style={{ fontSize: 11, fontWeight: 700, color: C.dim, textTransform: "uppercase", letterSpacing: ".05em", fontFamily: FH, marginBottom: 3 }}>{f.l}</p>
                <p style={{ fontSize: 13.5, color: C.text }}>{f.v}</p>
              </div>
            ))}
          </div>

          <p style={{ fontSize: 13.5, color: C.sub, lineHeight: 1.6, marginBottom: 22 }}>{pending.description.ru}</p>

          <div style={{ display: "flex", gap: 10 }}>
            <button onClick={() => setInstitutionStatus("approved")} style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", gap: 7, padding: 12, borderRadius: 11, background: C.ok, color: "#fff", fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
              <CheckCircle size={16} /> {t({ ru: "Одобрить", tg: "Тасдиқ" })}
            </button>
            <button onClick={() => setInstitutionStatus("rejected")} style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", gap: 7, padding: 12, borderRadius: 11, background: `${C.red}14`, border: `1px solid ${C.red}44`, color: C.red, fontFamily: FH, fontWeight: 800, fontSize: 13.5, cursor: "pointer" }}>
              <XCircle size={16} /> {t({ ru: "Отклонить", tg: "Рад кардан" })}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
