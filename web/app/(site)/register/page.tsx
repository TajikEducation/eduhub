"use client";

import { Suspense, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Phone, Mail, Lock, Info } from "lucide-react";
import { C, FH, FB } from "@/lib/data";
import { useT } from "@/lib/i18n";

const inputStyle: React.CSSProperties = {
  width: "100%", padding: "12px 14px 12px 40px", borderRadius: 11, border: `1px solid ${C.border}`,
  background: C.s2, color: C.text, fontFamily: FB, fontSize: 14, outline: "none", boxSizing: "border-box",
};

export default function RegisterPage() {
  return (
    <Suspense fallback={null}>
      <RegisterInner />
    </Suspense>
  );
}

function RegisterInner() {
  const t = useT();
  const router = useRouter();
  const searchParams = useSearchParams();
  const next = searchParams.get("next") || "/onboarding";

  const [phone, setPhone] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [consent, setConsent] = useState(false);

  const canSubmit = phone.trim().length >= 6 && password.length >= 4 && consent;

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    router.push(`/verify?next=${encodeURIComponent(next)}`);
  }

  function registerWithGoogle() {
    if (!consent) return;
    router.push(`/verify?next=${encodeURIComponent(next)}`);
  }

  return (
    <div style={{ maxWidth: 400, margin: "0 auto", padding: "72px 28px 80px" }}>
      <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 26, color: C.text, marginBottom: 6, textAlign: "center", letterSpacing: "-.02em" }}>
        {t({ ru: "Регистрация", tg: "Сабти ном" })}
      </h1>
      <p style={{ fontSize: 13.5, color: C.sub, textAlign: "center", marginBottom: 28 }}>
        {t({ ru: "Один аккаунт — для поиска учреждений, отзывов, вакансий и регистрации своего учреждения", tg: "Як ҳисоб — барои ҷустуҷӯи муассисаҳо, шарҳҳо, ҷойҳои холӣ ва сабти муассисаи худ" })}
      </p>

      <form onSubmit={submit} style={{ display: "flex", flexDirection: "column", gap: 14 }}>
        <div style={{ position: "relative" }}>
          <Phone size={16} style={{ position: "absolute", left: 13, top: "50%", transform: "translateY(-50%)", color: C.dim }} />
          <input value={phone} onChange={(e) => setPhone(e.target.value)} placeholder={t({ ru: "Номер телефона", tg: "Рақами телефон" })} style={inputStyle} />
        </div>
        <div style={{ position: "relative" }}>
          <Mail size={16} style={{ position: "absolute", left: 13, top: "50%", transform: "translateY(-50%)", color: C.dim }} />
          <input value={email} onChange={(e) => setEmail(e.target.value)} placeholder={t({ ru: "Email (необязательно)", tg: "Email (ихтиёрӣ)" })} style={inputStyle} />
        </div>
        <div style={{ position: "relative" }}>
          <Lock size={16} style={{ position: "absolute", left: 13, top: "50%", transform: "translateY(-50%)", color: C.dim }} />
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder={t({ ru: "Пароль", tg: "Парол" })} style={inputStyle} />
        </div>

        <label style={{ display: "flex", alignItems: "flex-start", gap: 9, fontSize: 12, color: C.sub, lineHeight: 1.5, cursor: "pointer" }}>
          <input type="checkbox" checked={consent} onChange={(e) => setConsent(e.target.checked)} style={{ marginTop: 2, flexShrink: 0 }} />
          {t({
            ru: "Согласен(на) на обработку персональных данных в соответствии с Законом Республики Таджикистан №1537",
            tg: "Розӣ ҳастам, ки маълумоти шахсии ман мутобиқи Қонуни Ҷумҳурии Тоҷикистон №1537 коркард шавад",
          })}
        </label>

        <button type="submit" disabled={!canSubmit} style={{ padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: canSubmit ? "pointer" : "default", opacity: canSubmit ? 1 : 0.5 }}>
          {t({ ru: "Зарегистрироваться", tg: "Сабти ном шудан" })}
        </button>
      </form>

      <div style={{ display: "flex", alignItems: "center", gap: 12, margin: "22px 0" }}>
        <div style={{ flex: 1, height: 1, background: C.border }} />
        <span style={{ fontSize: 12, color: C.dim, fontFamily: FH }}>{t({ ru: "или", tg: "ё" })}</span>
        <div style={{ flex: 1, height: 1, background: C.border }} />
      </div>

      <button onClick={registerWithGoogle} disabled={!consent} style={{ display: "flex", alignItems: "center", justifyContent: "center", gap: 10, width: "100%", padding: 12, borderRadius: 11, background: "transparent", border: `1px solid ${C.border}`, color: C.text, fontFamily: FH, fontWeight: 700, fontSize: 13.5, cursor: consent ? "pointer" : "default", opacity: consent ? 1 : 0.5 }}>
        <span style={{ width: 20, height: 20, borderRadius: 6, background: "#fff", color: "#4285F4", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 12, fontWeight: 900 }}>G</span>
        {t({ ru: "Продолжить с Google", tg: "Идома бо Google" })}
      </button>

      <p style={{ fontSize: 13, color: C.sub, textAlign: "center", marginTop: 22 }}>
        {t({ ru: "Уже есть аккаунт?", tg: "Ҳисоб доред?" })}{" "}
        <Link href="/login" style={{ color: C.teal, fontWeight: 700, textDecoration: "none" }}>
          {t({ ru: "Войти", tg: "Ворид шудан" })}
        </Link>
      </p>

      <div style={{ display: "flex", gap: 8, marginTop: 24, padding: "10px 14px", borderRadius: 10, background: `${C.gold}14`, border: `1px solid ${C.gold}33` }}>
        <Info size={14} style={{ color: C.gold, flexShrink: 0, marginTop: 1 }} />
        <p style={{ fontSize: 11.5, color: C.sub, lineHeight: 1.5 }}>
          {t({ ru: "Демо-режим прототипа: SMS/email с кодом реально не отправляется, следующий шаг принимает любой код.", tg: "Реҷаи демо: SMS/email бо код воқеан фиристода намешавад, қадами оянда ҳар кодро қабул мекунад." })}
        </p>
      </div>
    </div>
  );
}
