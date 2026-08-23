"use client";

import Link from "next/link";
import { UserRoundPen, Baby, Building2, ArrowRight } from "lucide-react";
import { C, FH } from "@/lib/data";
import { useT } from "@/lib/i18n";

const CHOICES = [
  { icon: UserRoundPen, title: { ru: "Дополнить профиль", tg: "Профилро пурра кунед" }, desc: { ru: "Фото, о себе — пригодится, если решите откликаться на вакансии", tg: "Расм, дар бораи худ — агар ба ҷойҳои холӣ ҷавоб диҳед, лозим мешавад" }, href: "/account" },
  { icon: Baby, title: { ru: "Добавить детей", tg: "Фарзандонро илова кунед" }, desc: { ru: "Привяжите ребёнка к учреждению — это нужно, чтобы оставлять отзывы", tg: "Фарзандро ба муассиса пайваст кунед — барои гузоштани шарҳ лозим аст" }, href: "/account" },
  { icon: Building2, title: { ru: "Подать заявку на учреждение", tg: "Дархост барои муассиса пешниҳод кунед" }, desc: { ru: "Вы представляете сад, школу, центр или вуз? Зарегистрируйте профиль", tg: "Шумо намояндаи боғча, мактаб, марказ ё донишгоҳ ҳастед? Профилро сабт кунед" }, href: "/register-institution" },
];

export default function OnboardingPage() {
  const t = useT();

  return (
    <div style={{ maxWidth: 720, margin: "0 auto", padding: "72px 28px 80px" }}>
      <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(22px,3vw,30px)", color: C.text, marginBottom: 8, textAlign: "center", letterSpacing: "-.02em" }}>
        {t({ ru: "Аккаунт создан. Что дальше?", tg: "Ҳисоб сохта шуд. Баъд чӣ?" })}
      </h1>
      <p style={{ fontSize: 14, color: C.sub, textAlign: "center", marginBottom: 36 }}>
        {t({ ru: "Можно пропустить и сделать это позже из личного кабинета", tg: "Метавонед гузаред ва баъдтар аз кабинети шахсӣ иҷро кунед" })}
      </p>

      <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {CHOICES.map((c) => (
          <Link key={c.href + c.title.ru} href={c.href} style={{ display: "flex", alignItems: "center", gap: 16, borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: "18px 20px", textDecoration: "none" }}>
            <div style={{ width: 44, height: 44, borderRadius: 12, background: `${C.teal}18`, display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
              <c.icon size={20} style={{ color: C.teal }} />
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 15, color: C.text, marginBottom: 3 }}>{t(c.title)}</p>
              <p style={{ fontSize: 12.5, color: C.sub, lineHeight: 1.5 }}>{t(c.desc)}</p>
            </div>
            <ArrowRight size={16} style={{ color: C.dim, flexShrink: 0 }} />
          </Link>
        ))}
      </div>

      <div style={{ textAlign: "center", marginTop: 28 }}>
        <Link href="/" style={{ fontSize: 13, color: C.dim, fontFamily: FH, textDecoration: "none" }}>
          {t({ ru: "Позже, на главную →", tg: "Баъдтар, ба саҳифаи асосӣ →" })}
        </Link>
      </div>
    </div>
  );
}
