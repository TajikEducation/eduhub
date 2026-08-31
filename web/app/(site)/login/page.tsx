"use client";

import { Suspense, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Mail, Lock, Info } from "lucide-react";
import { C, FH, FB } from "@/lib/data";
import { useAppState } from "@/lib/app-state";
import { useT } from "@/lib/i18n";
import { useLoginMutation } from "./api/authApi";
import { setTokens } from "@/lib/authToken";

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

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [login, { isLoading, error }] = useLoginMutation();

  const canSubmit = email.trim().length > 0 && password.length >= 4;

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    try {
      const res = await login({ email: email.trim(), password }).unwrap();
      setTokens(res.tokens.access_token, res.tokens.refresh_token);
      setRole(res.user.role);
      router.push(next);
    } catch {
      // ошибка уже в error через хук — рендерится ниже
    }
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
          <Mail size={16} style={{ position: "absolute", left: 13, top: "50%", transform: "translateY(-50%)", color: C.dim }} />
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder={t({ ru: "Email", tg: "Email" })} style={inputStyle} />
        </div>
        <div style={{ position: "relative" }}>
          <Lock size={16} style={{ position: "absolute", left: 13, top: "50%", transform: "translateY(-50%)", color: C.dim }} />
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder={t({ ru: "Пароль", tg: "Парол" })} style={inputStyle} />
        </div>
        {error && (
          <p style={{ fontSize: 12.5, color: C.red }}>
            {t({ ru: "Неверный email или пароль", tg: "Email ё парол нодуруст аст" })}
          </p>
        )}
        <button type="submit" disabled={!canSubmit || isLoading} style={{ padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: canSubmit ? "pointer" : "default", opacity: canSubmit && !isLoading ? 1 : 0.5 }}>
          {isLoading ? t({ ru: "Входим…", tg: "Воридшавӣ…" }) : t({ ru: "Войти", tg: "Ворид шудан" })}
        </button>
      </form>

      <p style={{ fontSize: 13, color: C.sub, textAlign: "center", marginTop: 22 }}>
        {t({ ru: "Нет аккаунта?", tg: "Ҳисоб надоред?" })}{" "}
        <Link href={`/register${next !== "/account" ? `?next=${encodeURIComponent(next)}` : ""}`} style={{ color: C.teal, fontWeight: 700, textDecoration: "none" }}>
          {t({ ru: "Зарегистрироваться", tg: "Сабти ном шудан" })}
        </Link>
      </p>

      <div style={{ display: "flex", gap: 8, marginTop: 24, padding: "10px 14px", borderRadius: 10, background: `${C.gold}14`, border: `1px solid ${C.gold}33` }}>
        <Info size={14} style={{ color: C.gold, flexShrink: 0, marginTop: 1 }} />
        <p style={{ fontSize: 11.5, color: C.sub, lineHeight: 1.5 }}>
          {t({ ru: "Вход через backend/internal/auth (email+пароль). Вход через Google пока не реализован.", tg: "Воридшавӣ тавассути backend/internal/auth (email+парол). Воридшавӣ тавассути Google ҳанӯз татбиқ нашудааст." })}
        </p>
      </div>
    </div>
  );
}
