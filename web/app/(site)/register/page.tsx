"use client";

import { Suspense, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Mail, Lock, User, Info } from "lucide-react";
import { C, FH, FB } from "@/lib/data";
import { useAppState } from "@/lib/app-state";
import { useT } from "@/lib/i18n";
import { useRegisterMutation } from "../login/api/authApi";
import { setTokens } from "@/lib/authToken";

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
  const { setRole } = useAppState();
  const next = searchParams.get("next") || "/onboarding";

  const [displayName, setDisplayName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [consent, setConsent] = useState(false);
  const [register, { isLoading, error }] = useRegisterMutation();

  const canSubmit = email.trim().length > 0 && password.length >= 8 && consent;

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    try {
      const res = await register({ email: email.trim(), password, display_name: displayName.trim() || undefined }).unwrap();
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
        {t({ ru: "Регистрация", tg: "Сабти ном" })}
      </h1>
      <p style={{ fontSize: 13.5, color: C.sub, textAlign: "center", marginBottom: 28 }}>
        {t({ ru: "Один аккаунт — для поиска учреждений, отзывов, вакансий и регистрации своего учреждения", tg: "Як ҳисоб — барои ҷустуҷӯи муассисаҳо, шарҳҳо, ҷойҳои холӣ ва сабти муассисаи худ" })}
      </p>

      <form onSubmit={submit} style={{ display: "flex", flexDirection: "column", gap: 14 }}>
        <div style={{ position: "relative" }}>
          <User size={16} style={{ position: "absolute", left: 13, top: "50%", transform: "translateY(-50%)", color: C.dim }} />
          <input value={displayName} onChange={(e) => setDisplayName(e.target.value)} placeholder={t({ ru: "Имя (необязательно)", tg: "Ном (ихтиёрӣ)" })} style={inputStyle} />
        </div>
        <div style={{ position: "relative" }}>
          <Mail size={16} style={{ position: "absolute", left: 13, top: "50%", transform: "translateY(-50%)", color: C.dim }} />
          <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder={t({ ru: "Email", tg: "Email" })} style={inputStyle} />
        </div>
        <div style={{ position: "relative" }}>
          <Lock size={16} style={{ position: "absolute", left: 13, top: "50%", transform: "translateY(-50%)", color: C.dim }} />
          <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder={t({ ru: "Пароль (минимум 8 символов)", tg: "Парол (ҳадди ақал 8 аломат)" })} style={inputStyle} />
        </div>

        <label style={{ display: "flex", alignItems: "flex-start", gap: 9, fontSize: 12, color: C.sub, lineHeight: 1.5, cursor: "pointer" }}>
          <input type="checkbox" checked={consent} onChange={(e) => setConsent(e.target.checked)} style={{ marginTop: 2, flexShrink: 0 }} />
          {t({
            ru: "Согласен(на) на обработку персональных данных в соответствии с Законом Республики Таджикистан №1537",
            tg: "Розӣ ҳастам, ки маълумоти шахсии ман мутобиқи Қонуни Ҷумҳурии Тоҷикистон №1537 коркард шавад",
          })}
        </label>

        {error && (
          <p style={{ fontSize: 12.5, color: C.red }}>
            {"status" in error && error.status === 409
              ? t({ ru: "Этот email уже зарегистрирован", tg: "Ин email аллакай сабт шудааст" })
              : t({ ru: "Не удалось зарегистрироваться", tg: "Сабти ном ба анҷом нарасид" })}
          </p>
        )}

        <button type="submit" disabled={!canSubmit || isLoading} style={{ padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: canSubmit ? "pointer" : "default", opacity: canSubmit && !isLoading ? 1 : 0.5 }}>
          {isLoading ? t({ ru: "Регистрируем…", tg: "Сабт мешавад…" }) : t({ ru: "Зарегистрироваться", tg: "Сабти ном шудан" })}
        </button>
      </form>

      <p style={{ fontSize: 13, color: C.sub, textAlign: "center", marginTop: 22 }}>
        {t({ ru: "Уже есть аккаунт?", tg: "Ҳисоб доред?" })}{" "}
        <Link href="/login" style={{ color: C.teal, fontWeight: 700, textDecoration: "none" }}>
          {t({ ru: "Войти", tg: "Ворид шудан" })}
        </Link>
      </p>

      <div style={{ display: "flex", gap: 8, marginTop: 24, padding: "10px 14px", borderRadius: 10, background: `${C.gold}14`, border: `1px solid ${C.gold}33` }}>
        <Info size={14} style={{ color: C.gold, flexShrink: 0, marginTop: 1 }} />
        <p style={{ fontSize: 11.5, color: C.sub, lineHeight: 1.5 }}>
          {t({ ru: "Регистрация через backend/internal/auth. Email-подтверждение и вход через Google пока не реализованы — аккаунт активен сразу.", tg: "Сабти ном тавассути backend/internal/auth. Тасдиқи email ва воридшавӣ тавассути Google ҳанӯз татбиқ нашудаанд — ҳисоб дарҳол фаъол мешавад." })}
        </p>
      </div>
    </div>
  );
}
