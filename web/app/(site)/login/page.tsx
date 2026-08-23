"use client";

import { Suspense, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Phone, Lock, Info } from "lucide-react";
import { C, FH, FB } from "@/lib/data";
import { useAppState } from "@/lib/app-state";
import { useT } from "@/lib/i18n";

const inputStyle: React.CSSProperties = {
  width: "100%", padding: "12px 14px 12px 40px", borderRadius: 11, border: `1px solid ${C.border}`,
  background: C.s2, color: C.text, fontFamily: FB, fontSize: 14, outline: "none", boxSizing: "border-box",
};

export default function LoginPage() {
  return (
    <Suspense fallback={null}>
      <LoginInner />
    </Suspense>
  );
}

function LoginInner() {
  const t = useT();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { setRole } = useAppState();
  const next = searchParams.get("next") || "/account";

  const [phone, setPhone] = useState("");
  const [password, setPassword] = useState("");

  const canSubmit = phone.trim().length >= 6 && password.length >= 4;

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    setRole("user");
    router.push(next);
  }

  function loginWithGoogle() {
    setRole("user");
    router.push(next);
  }

  return (
    <div style={{ maxWidth: 400, margin: "0 auto", padding: "72px 28px 80px" }}>
      <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 26, color: C.text, marginBottom: 6, textAlign: "center", letterSpacing: "-.02em" }}>
        {t({ ru: "Вход", tg: "Воридшавӣ" })}
      </h1>
      <p style={{ fontSize: 13.5, color: C.sub, textAlign: "center", marginBottom: 28 }}>
        {t({ ru: "Рады видеть вас снова на EduHub", tg: "Хуш омадед боз ба EduHub" })}
      </p>

      <form onSubmit={submit} style={{ display: "flex", flexDirection: "column", gap: 14 }}>
        <div style={{ position: "relative" }}>
          <Phone size={16} style={{ position: "absolute", left: 13, top: "50%", transform: "translateY(-50%)", color: C.dim }} />
          <input value={phone} onChange={(e) => setPhone(e.target.value)} placeholder={t({ ru: "Номер телефона", tg: "Рақами телефон" })} style={inputStyle} />
        </div>
        <div style={{ position: "relative" }}>
          <Lock size={16} style={{ position: "absolute", left: 13, top: "50%", transform: "translateY(-50%)", color: C.dim }} />
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder={t({ ru: "Пароль", tg: "Парол" })} style={inputStyle} />
        </div>
        <button type="submit" disabled={!canSubmit} style={{ padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: canSubmit ? "pointer" : "default", opacity: canSubmit ? 1 : 0.5 }}>
          {t({ ru: "Войти", tg: "Ворид шудан" })}
        </button>
      </form>

      <div style={{ display: "flex", alignItems: "center", gap: 12, margin: "22px 0" }}>
        <div style={{ flex: 1, height: 1, background: C.border }} />
        <span style={{ fontSize: 12, color: C.dim, fontFamily: FH }}>{t({ ru: "или", tg: "ё" })}</span>
        <div style={{ flex: 1, height: 1, background: C.border }} />
      </div>

      <button onClick={loginWithGoogle} style={{ display: "flex", alignItems: "center", justifyContent: "center", gap: 10, width: "100%", padding: 12, borderRadius: 11, background: "transparent", border: `1px solid ${C.border}`, color: C.text, fontFamily: FH, fontWeight: 700, fontSize: 13.5, cursor: "pointer" }}>
        <span style={{ width: 20, height: 20, borderRadius: 6, background: "#fff", color: "#4285F4", display: "flex", alignItems: "center", justifyContent: "center", fontSize: 12, fontWeight: 900 }}>G</span>
        {t({ ru: "Войти через Google", tg: "Ворид тавассути Google" })}
      </button>

      <p style={{ fontSize: 13, color: C.sub, textAlign: "center", marginTop: 22 }}>
        {t({ ru: "Нет аккаунта?", tg: "Ҳисоб надоред?" })}{" "}
        <Link href={`/register${next !== "/account" ? `?next=${encodeURIComponent(next)}` : ""}`} style={{ color: C.teal, fontWeight: 700, textDecoration: "none" }}>
          {t({ ru: "Зарегистрироваться", tg: "Сабти ном шудан" })}
        </Link>
      </p>

      <div style={{ display: "flex", gap: 8, marginTop: 24, padding: "10px 14px", borderRadius: 10, background: `${C.gold}14`, border: `1px solid ${C.gold}33` }}>
        <Info size={14} style={{ color: C.gold, flexShrink: 0, marginTop: 1 }} />
        <p style={{ fontSize: 11.5, color: C.sub, lineHeight: 1.5 }}>
          {t({ ru: "Демо-режим прототипа: вход не проверяет реальные данные, просто открывает личный кабинет.", tg: "Реҷаи демо: воридшавӣ маълумоти воқеиро санҷида намекунад, танҳо кабинети шахсиро мекушояд." })}
        </p>
      </div>
    </div>
  );
}
