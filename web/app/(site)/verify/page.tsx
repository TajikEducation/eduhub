"use client";

import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { ShieldCheck, Info } from "lucide-react";
import { C, FH } from "@/lib/data";
import { useAppState } from "@/lib/app-state";
import { useT } from "@/lib/i18n";

export default function VerifyPage() {
  return (
    <Suspense fallback={null}>
      <VerifyInner />
    </Suspense>
  );
}

function VerifyInner() {
  const t = useT();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { setRole } = useAppState();
  const next = searchParams.get("next") || "/onboarding";

  const [code, setCode] = useState("");

  function confirm(e: React.FormEvent) {
    e.preventDefault();
    if (code.trim().length !== 6) return;
    setRole("user");
    router.push(next);
  }

  return (
    <div style={{ maxWidth: 380, margin: "0 auto", padding: "80px 28px" }}>
      <div style={{ width: 52, height: 52, borderRadius: 14, background: `${C.teal}18`, display: "flex", alignItems: "center", justifyContent: "center", margin: "0 auto 18px" }}>
        <ShieldCheck size={24} style={{ color: C.teal }} />
      </div>
      <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 22, color: C.text, marginBottom: 8, textAlign: "center", letterSpacing: "-.02em" }}>
        {t({ ru: "Подтвердите номер", tg: "Рақамро тасдиқ кунед" })}
      </h1>
      <p style={{ fontSize: 13.5, color: C.sub, textAlign: "center", marginBottom: 28, lineHeight: 1.6 }}>
        {t({ ru: "Введите 6-значный код, отправленный по SMS или на email", tg: "Коди 6-рақамаро, ки тавассути SMS ё email фиристода шуд, ворид кунед" })}
      </p>

      <form onSubmit={confirm} style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <input
          value={code}
          onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
          placeholder="000000"
          inputMode="numeric"
          style={{ width: "100%", padding: "14px", borderRadius: 12, border: `1px solid ${C.border}`, background: C.s2, color: C.text, fontFamily: FH, fontWeight: 800, fontSize: 24, letterSpacing: "0.4em", textAlign: "center", outline: "none", boxSizing: "border-box" }}
        />
        <button type="submit" disabled={code.length !== 6} style={{ padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: code.length === 6 ? "pointer" : "default", opacity: code.length === 6 ? 1 : 0.5 }}>
          {t({ ru: "Подтвердить", tg: "Тасдиқ кардан" })}
        </button>
      </form>

      <div style={{ display: "flex", gap: 8, marginTop: 24, padding: "10px 14px", borderRadius: 10, background: `${C.gold}14`, border: `1px solid ${C.gold}33` }}>
        <Info size={14} style={{ color: C.gold, flexShrink: 0, marginTop: 1 }} />
        <p style={{ fontSize: 11.5, color: C.sub, lineHeight: 1.5 }}>
          {t({ ru: "Демо-режим: код реально не отправляется — введите любые 6 цифр.", tg: "Реҷаи демо: код воқеан фиристода намешавад — ягон 6 рақамро ворид кунед." })}
        </p>
      </div>
    </div>
  );
}
