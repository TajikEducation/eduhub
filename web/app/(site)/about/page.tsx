"use client";

import { TrendingUp, Users, Building2, Wallet, Target, Map, Mail, ArrowRight, CheckCircle2, Briefcase, ShieldCheck, BadgeCheck, UserCheck } from "lucide-react";
import { C, FH, FB, PHOTOS } from "@/lib/data";
import { SubjectMotifs } from "@/components/SubjectMotifs";
import { useReveal, revealStyle } from "@/lib/useReveal";
import { useT } from "@/lib/i18n";

const MARKET = [
  { l: { ru: "TAM · весь рынок", tg: "TAM · тамоми бозор" }, v: "1.5M", d: { ru: "Семьи с детьми + студенты по стране", tg: "Оилаҳо бо кӯдакон + донишҷӯён дар кишвар" } },
  { l: { ru: "SAM · доступный рынок", tg: "SAM · бозори дастрас" }, v: "200K", d: { ru: "Семьи в Душанбе (Фаза 1)", tg: "Оилаҳо дар Душанбе (Фазаи 1)" } },
  { l: { ru: "SOM · год 1", tg: "SOM · соли 1" }, v: "50K", d: { ru: "Активных родителей + 2K учреждений", tg: "Волидайни фаъол + 2K муассиса" } },
];

const SIDES = [
  {
    icon: Users, color: C.teal,
    title: { ru: "Родители и студенты", tg: "Волидайн ва донишҷӯён" },
    price: { ru: "Полностью бесплатно", tg: "Комилан ройгон" },
    points: [
      { ru: "Централизованный поиск с фильтрами", tg: "Ҷустуҷӯи марказонидашуда бо филтрҳо" },
      { ru: "Верифицированные отзывы и 8 метрик рейтинга", tg: "Шарҳҳои санҷидашуда ва 8 меъёри рейтинг" },
      { ru: "Полная финансовая прозрачность", tg: "Шаффофияти пурраи молиявӣ" },
    ],
  },
  {
    icon: Building2, color: C.teal,
    title: { ru: "Образовательные учреждения", tg: "Муассисаҳои таълимӣ" },
    price: { ru: "Free · Pro $30/мес · Enterprise $100/мес", tg: "Free · Pro $30/моҳ · Enterprise $100/моҳ" },
    points: [
      { ru: "Профиль с фото, видео и полной информацией", tg: "Профил бо акс, видео ва маълумоти пурра" },
      { ru: "Аналитика источников трафика и просмотров", tg: "Таҳлили манбаъҳои трафик ва дидан" },
      { ru: "Публикация вакансий и новостей", tg: "Нашри ҷойҳои холӣ ва хабарҳо" },
    ],
  },
  {
    icon: Briefcase, color: C.teal,
    title: { ru: "Соискатели-педагоги", tg: "Ҷуяндагони кор" },
    price: { ru: "Полностью бесплатно", tg: "Комилан ройгон" },
    points: [
      { ru: "Вакансии с фильтрами по должности и району", tg: "Ҷойҳои холӣ бо филтр аз рӯи вазифа ва ноҳия" },
      { ru: "Профиль учреждения как работодателя", tg: "Профили муассиса ҳамчун корфармо" },
      { ru: "Прямой контакт в чате", tg: "Тамоси бевосита дар чат" },
    ],
  },
];

const ROADMAP = [
  { phase: { ru: "MVP", tg: "MVP" }, period: { ru: "Мес. 1–3", tg: "Моҳи 1–3" }, result: { ru: "Запуск в Душанбе, 100 школ, 1K DAU", tg: "Оғоз дар Душанбе, 100 мактаб, 1K DAU" } },
  { phase: { ru: "Core", tg: "Core" }, period: { ru: "Мес. 4–6", tg: "Моҳи 4–6" }, result: { ru: "8 метрик, аналитика, премиум, 500 школ, 5K DAU", tg: "8 меъёр, таҳлил, премиум, 500 мактаб, 5K DAU" } },
  { phase: { ru: "Premium", tg: "Premium" }, period: { ru: "Мес. 7–9", tg: "Моҳи 7–9" }, result: { ru: "Приложения, вакансии, 1.5K школ, 10K DAU", tg: "Барномаҳо, ҷойҳои холӣ, 1.5K мактаб, 10K DAU" } },
  { phase: { ru: "Scale", tg: "Scale" }, period: { ru: "Мес. 10–12", tg: "Моҳи 10–12" }, result: { ru: "Расширение по городам, интеграция с МОН, 2K школ, 15K DAU", tg: "Васеъшавӣ дар шаҳрҳо, ҳамгироӣ бо ВМО, 2K мактаб, 15K DAU" } },
];

const TRUST_PILLARS = [
  {
    icon: ShieldCheck,
    title: { ru: "Лицензия МОН РТ", tg: "Литсензияи ВМО ҶТ" },
    desc: { ru: "Каждое учреждение проверяется по номеру лицензии перед публикацией профиля", tg: "Ҳар як муассиса пеш аз нашри профил аз рӯи рақами литсензия санҷида мешавад" },
  },
  {
    icon: UserCheck,
    title: { ru: "Привязка к ребёнку", tg: "Алоқаманд ба фарзанд" },
    desc: { ru: "Отзыв может оставить только родитель, чей ребёнок структурно связан с учреждением", tg: "Шарҳро танҳо волидайне гузошта метавонад, ки фарзандаш бо муассиса алоқаманд аст" },
  },
  {
    icon: BadgeCheck,
    title: { ru: "Верификация владельца профиля", tg: "Тасдиқи соҳиби профил" },
    desc: { ru: "Представитель учреждения подтверждает полномочия документом до публикации", tg: "Намояндаи муассиса пеш аз нашр ваколати худро бо ҳуҷҷат тасдиқ мекунад" },
  },
];

const VISION = [
  { y: { ru: "Год 1", tg: "Соли 1" }, market: { ru: "Таджикистан", tg: "Тоҷикистон" }, goal: "15K DAU · 2K учреждений · $120K" },
  { y: { ru: "Год 2", tg: "Соли 2" }, market: { ru: "+ Узбекистан", tg: "+ Ӯзбекистон" }, goal: "50K DAU · $400K+ MRR" },
  { y: { ru: "Год 3", tg: "Соли 3" }, market: { ru: "+ Кыргызстан, Казахстан", tg: "+ Қирғизистон, Қазоқистон" }, goal: "200K DAU · раунд Series A" },
  { y: { ru: "Год 4–5", tg: "Соли 4–5" }, market: { ru: "Весь СНГ", tg: "Тамоми ИДМ" }, goal: "1M+ DAU · устойчивая прибыльность" },
];

export default function AboutPage() {
  const t = useT();
  const { ref: statsRef, visible: statsVisible } = useReveal<HTMLDivElement>();
  const { ref: sidesRef, visible: sidesVisible } = useReveal<HTMLDivElement>();
  const { ref: roadmapRef, visible: roadmapVisible } = useReveal<HTMLDivElement>();

  return (
    <div>
      {/* ── HERO ── */}
      <div style={{ position: "relative", overflow: "hidden", padding: "72px 28px 56px", textAlign: "center" }}>
        <img src={PHOTOS.graduation} alt="" style={{ position: "absolute", inset: 0, width: "100%", height: "100%", objectFit: "cover", opacity: 0.22 }} />
        <div style={{ position: "absolute", inset: 0, background: `linear-gradient(180deg, ${C.overlay}CC 0%, ${C.overlay}EE 55%, ${C.overlay} 100%)` }} />
        <SubjectMotifs />
        <div style={{ position: "relative", maxWidth: 720, margin: "0 auto" }}>
          <span style={{ display: "inline-flex", padding: "6px 14px", borderRadius: 999, border: `1px solid ${C.teal}66`, background: `${C.teal}22`, fontSize: 12.5, fontWeight: 700, color: C.teal, fontFamily: FH, marginBottom: 18 }}>
            {t({ ru: "Для партнёров и инвесторов", tg: "Барои шарикон ва сармоягузорон" })}
          </span>
          <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(28px,4vw,44px)", color: "#fff", marginBottom: 14, letterSpacing: "-.02em", lineHeight: 1.15 }}>
            {t({ ru: "«Google Maps для образования» в Таджикистане", tg: "«Google Maps барои маориф» дар Тоҷикистон" })}
          </h1>
          <p style={{ fontSize: 15, color: "rgba(255,255,255,.72)", lineHeight: 1.65 }}>
            {t({
              ru: "Первая национальная платформа для поиска, сравнения и оценки всех типов образовательных учреждений — от детских садов до университетов. Трёхсторонний рынок без конкурентов в нише, с планом расширения по всему СНГ.",
              tg: "Аввалин платформаи миллӣ барои ҷустуҷӯ, муқоиса ва бақои ҳама намудҳои муассисаҳои таълимӣ — аз боғчаҳо то донишгоҳҳо. Бозори сеҷониба бе рақибон дар ин соҳа, бо нақшаи васеъшавӣ дар тамоми ИДМ.",
            })}
          </p>
        </div>
      </div>

      {/* ── KEY STATS ── */}
      <div ref={statsRef} style={{ maxWidth: 1100, margin: "0 auto", padding: "0 28px 56px", ...revealStyle(statsVisible) }}>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(4,1fr)", gap: 14 }}>
          {[
            { icon: Users, l: { ru: "DAU (год 1)", tg: "DAU (соли 1)" }, v: "15 000" },
            { icon: Building2, l: { ru: "Активных учреждений", tg: "Муассисаҳои фаъол" }, v: "2 000+" },
            { icon: Wallet, l: { ru: "Доход за год 1", tg: "Даромад дар соли 1" }, v: "$136K" },
            { icon: TrendingUp, l: { ru: "Точка окупаемости", tg: "Нуқтаи баргашт" }, v: "18–24 мес." },
          ].map((s) => (
            <div key={s.l.ru} style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 20 }}>
              <s.icon size={18} style={{ color: C.teal, marginBottom: 10 }} />
              <p style={{ fontFamily: FH, fontWeight: 900, fontSize: 22, color: C.text }}>{s.v}</p>
              <p style={{ fontSize: 12.5, color: C.sub, marginTop: 3 }}>{t(s.l)}</p>
            </div>
          ))}
        </div>
      </div>

      {/* ── MARKET SIZE ── */}
      <div style={{ maxWidth: 1100, margin: "0 auto", padding: "0 28px 56px" }}>
        <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 20, color: C.text, marginBottom: 18, display: "flex", alignItems: "center", gap: 8 }}>
          <Target size={19} style={{ color: C.teal }} /> {t({ ru: "Размер рынка", tg: "Андозаи бозор" })}
        </h2>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3,1fr)", gap: 14 }}>
          {MARKET.map((m) => (
            <div key={m.l.ru} style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 22 }}>
              <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 12.5, color: C.teal, textTransform: "uppercase", letterSpacing: ".04em", marginBottom: 8 }}>{t(m.l)}</p>
              <p style={{ fontFamily: FH, fontWeight: 900, fontSize: 32, color: C.text, marginBottom: 6 }}>{m.v}</p>
              <p style={{ fontSize: 13, color: C.sub, lineHeight: 1.5 }}>{t(m.d)}</p>
            </div>
          ))}
        </div>
      </div>

      {/* ── TRUST / METHODOLOGY ── */}
      <div style={{ maxWidth: 1100, margin: "0 auto", padding: "0 28px 56px" }}>
        <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 20, color: C.text, marginBottom: 6, display: "flex", alignItems: "center", gap: 8 }}>
          <ShieldCheck size={19} style={{ color: C.teal }} /> {t({ ru: "Методология проверки", tg: "Методологияи санҷиш" })}
        </h2>
        <p style={{ fontSize: 13.5, color: C.sub, marginBottom: 20, maxWidth: 640, lineHeight: 1.6 }}>
          {t({
            ru: "Рейтинг нельзя купить: платный тариф учреждения влияет только на видимость в поиске, но никогда — на значение оценки. Оценка формируется только 8 независимыми метриками из верифицированных отзывов.",
            tg: "Рейтингро харидан мумкин нест: тарифи пулакии муассиса танҳо ба намоён будан дар ҷустуҷӯ таъсир мерасонад, аммо ҳеҷ гоҳ — ба арзиши баҳо. Баҳо танҳо аз 8 меъёри мустақил аз шарҳҳои тасдиқшуда ташкил мешавад.",
          })}
        </p>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3,1fr)", gap: 14 }}>
          {TRUST_PILLARS.map((p) => (
            <div key={p.title.ru} style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 22 }}>
              <div style={{ width: 40, height: 40, borderRadius: 11, background: `${C.teal}18`, display: "flex", alignItems: "center", justifyContent: "center", marginBottom: 12 }}>
                <p.icon size={18} style={{ color: C.teal }} />
              </div>
              <h3 style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text, marginBottom: 6 }}>{t(p.title)}</h3>
              <p style={{ fontSize: 12.5, color: C.sub, lineHeight: 1.55 }}>{t(p.desc)}</p>
            </div>
          ))}
        </div>
      </div>

      {/* ── THREE SIDES ── */}
      <div ref={sidesRef} style={{ maxWidth: 1100, margin: "0 auto", padding: "0 28px 56px", ...revealStyle(sidesVisible) }}>
        <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 20, color: C.text, marginBottom: 18 }}>
          {t({ ru: "Трёхсторонняя модель", tg: "Модели сеҷониба" })}
        </h2>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3,1fr)", gap: 14 }}>
          {SIDES.map((s) => (
            <div key={s.title.ru} style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 24 }}>
              <div style={{ width: 42, height: 42, borderRadius: 12, background: `${s.color}18`, display: "flex", alignItems: "center", justifyContent: "center", marginBottom: 14 }}>
                <s.icon size={19} style={{ color: s.color }} />
              </div>
              <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 15.5, color: C.text, marginBottom: 6 }}>{t(s.title)}</h3>
              <p style={{ fontSize: 12, color: C.teal, fontFamily: FH, fontWeight: 700, marginBottom: 14 }}>{t(s.price)}</p>
              <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                {s.points.map((p) => (
                  <div key={p.ru} style={{ display: "flex", alignItems: "flex-start", gap: 7, fontSize: 12.5, color: C.sub, lineHeight: 1.5 }}>
                    <CheckCircle2 size={13} style={{ color: C.ok, flexShrink: 0, marginTop: 2 }} /> {t(p)}
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* ── ROADMAP ── */}
      <div ref={roadmapRef} style={{ maxWidth: 1100, margin: "0 auto", padding: "0 28px 56px", ...revealStyle(roadmapVisible) }}>
        <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 20, color: C.text, marginBottom: 18, display: "flex", alignItems: "center", gap: 8 }}>
          <Map size={19} style={{ color: C.teal }} /> {t({ ru: "Дорожная карта — год 1", tg: "Нақшаи роҳ — соли 1" })}
        </h2>
        <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: "6px 24px" }}>
          {ROADMAP.map((r, i) => (
            <div key={r.phase.ru} style={{ display: "flex", alignItems: "center", gap: 18, padding: "18px 0", borderTop: i > 0 ? `1px solid ${C.border}` : "none" }}>
              <div style={{ width: 90, flexShrink: 0 }}>
                <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 14, color: C.teal }}>{t(r.phase)}</p>
                <p style={{ fontSize: 11.5, color: C.dim, marginTop: 2 }}>{t(r.period)}</p>
              </div>
              <p style={{ fontSize: 13.5, color: C.text, lineHeight: 1.5 }}>{t(r.result)}</p>
            </div>
          ))}
        </div>

        <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 16, color: C.text, margin: "32px 0 14px" }}>
          {t({ ru: "Видение на 5 лет", tg: "Дурнамо барои 5 сол" })}
        </h3>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(4,1fr)", gap: 12 }}>
          {VISION.map((v) => (
            <div key={v.y.ru} style={{ borderRadius: 12, border: `1px solid ${C.border}`, background: C.s1, padding: 16 }}>
              <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 12.5, color: C.teal, marginBottom: 6 }}>{t(v.y)}</p>
              <p style={{ fontSize: 13, color: C.text, fontWeight: 600, marginBottom: 4 }}>{t(v.market)}</p>
              <p style={{ fontSize: 11.5, color: C.sub, lineHeight: 1.5 }}>{v.goal}</p>
            </div>
          ))}
        </div>
      </div>

      {/* ── CONTACT ── */}
      <div style={{ maxWidth: 720, margin: "0 auto", padding: "0 28px 80px", textAlign: "center" }}>
        <div style={{ borderRadius: 18, border: `1px solid ${C.teal}33`, background: `${C.teal}0d`, padding: "40px 32px" }}>
          <Mail size={26} style={{ color: C.teal, marginBottom: 14 }} />
          <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 20, color: C.text, marginBottom: 10 }}>
            {t({ ru: "Хотите узнать больше?", tg: "Мехоҳед бештар донед?" })}
          </h2>
          <p style={{ fontSize: 14, color: C.sub, marginBottom: 22, lineHeight: 1.6 }}>
            {t({ ru: "Полный бизнес-план доступен для партнёров, спонсоров и инвесторов по запросу.", tg: "Нақшаи пурраи бизнес барои шарикон, сарпарастон ва сармоягузорон бо дархост дастрас аст." })}
          </p>
          <a href="mailto:partners@eduhub.tj" style={{ display: "inline-flex", alignItems: "center", gap: 8, padding: "13px 26px", borderRadius: 12, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, textDecoration: "none" }}>
            {t({ ru: "Написать команде", tg: "Ба даста нависед" })} <ArrowRight size={15} />
          </a>
        </div>
      </div>
    </div>
  );
}
