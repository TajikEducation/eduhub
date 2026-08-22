"use client";

import { useState } from "react";
import Link from "next/link";
import { Target, ShieldCheck, MapPin, Mail, ArrowRight, Users, Building2, Briefcase, MessagesSquare, Wallet, SearchX, UserCheck, BadgeCheck, Send, CheckCircle2 } from "lucide-react";
import { C, FH, FB, PHOTOS, REGION_ORDER, REGION_LABEL } from "@/lib/data";
import { SubjectMotifs } from "@/components/SubjectMotifs";
import { useT } from "@/lib/i18n";

const TOPICS = [
  { v: "parent", ru: "Я родитель/студент", tg: "Ман волидайн/донишҷӯ ҳастам" },
  { v: "applicant", ru: "Я соискатель", tg: "Ман ҷуяндаи кор ҳастам" },
  { v: "institution", ru: "Я представляю учреждение", tg: "Ман намояндаи муассиса ҳастам" },
  { v: "other", ru: "Другое", tg: "Дигар" },
] as const;

function ContactForm() {
  const t = useT();
  const [topic, setTopic] = useState<string>(TOPICS[0].v);
  const [text, setText] = useState("");
  const [contact, setContact] = useState("");
  const [honeypot, setHoneypot] = useState(""); // .claude/rules/security.md — базовая защита от ботов на публичной форме
  const [sent, setSent] = useState(false);

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (honeypot.trim() || !text.trim()) return; // заполненный honeypot = бот, тихо игнорируем
    setSent(true);
  }

  if (sent) {
    return (
      <div style={{ borderRadius: 16, border: `1px solid ${C.ok}44`, background: `${C.ok}12`, padding: 24, textAlign: "center" }}>
        <CheckCircle2 size={22} style={{ color: C.ok, marginBottom: 8 }} />
        <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14.5, color: C.text }}>
          {t({ ru: "Вопрос отправлен, спасибо!", tg: "Савол фиристода шуд, ташаккур!" })}
        </p>
      </div>
    );
  }

  return (
    <form onSubmit={submit} style={{ position: "relative", borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 24, display: "flex", flexDirection: "column", gap: 14 }}>
      <input value={honeypot} onChange={(e) => setHoneypot(e.target.value)} name="website" tabIndex={-1} autoComplete="off" style={{ position: "absolute", left: -9999, width: 1, height: 1, opacity: 0 }} aria-hidden="true" />
      <div>
        <label style={{ display: "block", fontSize: 11.5, fontWeight: 700, color: C.dim, textTransform: "uppercase", letterSpacing: ".05em", fontFamily: FH, marginBottom: 6 }}>
          {t({ ru: "Кто вы?", tg: "Шумо кистед?" })}
        </label>
        <select value={topic} onChange={(e) => setTopic(e.target.value)} style={{ width: "100%", padding: "10px 13px", borderRadius: 10, border: `1px solid ${C.border}`, background: C.s2, color: C.text, fontFamily: FB, fontSize: 13.5, outline: "none", boxSizing: "border-box" }}>
          {TOPICS.map((o) => <option key={o.v} value={o.v}>{t({ ru: o.ru, tg: o.tg })}</option>)}
        </select>
      </div>
      <div>
        <label style={{ display: "block", fontSize: 11.5, fontWeight: 700, color: C.dim, textTransform: "uppercase", letterSpacing: ".05em", fontFamily: FH, marginBottom: 6 }}>
          {t({ ru: "Ваш вопрос", tg: "Саволи шумо" })}
        </label>
        <textarea value={text} onChange={(e) => setText(e.target.value)} rows={4} style={{ width: "100%", padding: "10px 13px", borderRadius: 10, border: `1px solid ${C.border}`, background: C.s2, color: C.text, fontFamily: FB, fontSize: 13.5, outline: "none", boxSizing: "border-box", resize: "vertical" }} />
      </div>
      <div>
        <label style={{ display: "block", fontSize: 11.5, fontWeight: 700, color: C.dim, textTransform: "uppercase", letterSpacing: ".05em", fontFamily: FH, marginBottom: 6 }}>
          {t({ ru: "Email или телефон (необязательно)", tg: "Email ё телефон (ихтиёрӣ)" })}
        </label>
        <input value={contact} onChange={(e) => setContact(e.target.value)} style={{ width: "100%", padding: "10px 13px", borderRadius: 10, border: `1px solid ${C.border}`, background: C.s2, color: C.text, fontFamily: FB, fontSize: 13.5, outline: "none", boxSizing: "border-box" }} />
      </div>
      <button type="submit" disabled={!text.trim()} style={{ display: "flex", alignItems: "center", justifyContent: "center", gap: 8, padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer", opacity: text.trim() ? 1 : 0.5 }}>
        <Send size={15} /> {t({ ru: "Отправить", tg: "Фиристодан" })}
      </button>
    </form>
  );
}

const PROBLEMS = [
  { icon: SearchX, title: { ru: "Информация разбросана", tg: "Маълумот пароканда аст" }, desc: { ru: "Instagram школы, чужие отзывы в VK-группах, советы знакомых — вместо одного места, где можно сравнить варианты рядом друг с другом.", tg: "Instagram-и мактаб, шарҳҳо дар гурӯҳҳои VK, маслиҳати шиносон — ба ҷои як ҷое, ки вариантҳоро паҳлуи ҳам муқоиса кардан мумкин аст." } },
  { icon: MessagesSquare, title: { ru: "Отзывам сложно доверять", tg: "Ба шарҳҳо бовар кардан душвор аст" }, desc: { ru: "Кто угодно может написать что угодно о любой школе — без проверки, что этот человек вообще имеет к ней отношение.", tg: "Ҳар кас метавонад дар бораи ҳар мактаб чизе нависад — бе санҷиши он, ки ин шахс воқеан ба он алоқаманд аст." } },
  { icon: Wallet, title: { ru: "Полная стоимость всплывает потом", tg: "Арзиши пурра баъдтар зоҳир мешавад" }, desc: { ru: "Обучение — это ещё развозка, питание, кружки, форма. Родитель часто узнаёт реальную сумму только после того, как ребёнок уже зачислен.", tg: "Таълим — ин боз расонидан, ғизо, машғулиятҳо, либос. Волидайн аксар вақт маблағи воқеиро танҳо пас аз қабули фарзанд мефаҳмад." } },
];

const RATING_CATEGORIES = [
  { ru: "Качество обучения", tg: "Сифати таълим" },
  { ru: "Условия и помещения", tg: "Шароит ва биноҳо" },
  { ru: "Безопасность", tg: "Бехатарӣ" },
  { ru: "Питание", tg: "Ғизо" },
  { ru: "Развозка", tg: "Расонидан" },
  { ru: "Цена/качество", tg: "Нарх/сифат" },
  { ru: "Участие родителей", tg: "Иштироки волидайн" },
  { ru: "Инклюзивность", tg: "Фарогирӣ" },
];

const TRUST_PILLARS = [
  { icon: ShieldCheck, title: { ru: "Лицензия МОН РТ", tg: "Литсензияи ВМО ҶТ" }, desc: { ru: "Каждое учреждение проверяется по номеру лицензии перед публикацией профиля", tg: "Ҳар як муассиса пеш аз нашри профил аз рӯи рақами литсензия санҷида мешавад" } },
  { icon: UserCheck, title: { ru: "Привязка к ребёнку", tg: "Алоқаманд ба фарзанд" }, desc: { ru: "Отзыв может оставить только родитель, чей ребёнок структурно связан с учреждением", tg: "Шарҳро танҳо волидайне гузошта метавонад, ки фарзандаш бо муассиса алоқаманд аст" } },
  { icon: BadgeCheck, title: { ru: "Верификация владельца профиля", tg: "Тасдиқи соҳиби профил" }, desc: { ru: "Представитель учреждения подтверждает полномочия документом до публикации", tg: "Намояндаи муассиса пеш аз нашр ваколати худро бо ҳуҷҷат тасдиқ мекунад" } },
];

const AUDIENCE_LINKS = [
  { icon: Users, label: { ru: "Родителям и студентам", tg: "Ба волидайн ва донишҷӯён" }, href: "/guide/parents" },
  { icon: Briefcase, label: { ru: "Соискателям", tg: "Ба ҷуяндагони кор" }, href: "/guide/applicants" },
  { icon: Building2, label: { ru: "Учреждениям", tg: "Ба муассисаҳо" }, href: "/guide/institutions" },
];

export default function CompanyPage() {
  const t = useT();

  return (
    <div>
      {/* ── HERO ── */}
      <div style={{ position: "relative", overflow: "hidden", padding: "72px 28px 56px", textAlign: "center" }}>
        <img src={PHOTOS.campus2} alt="" style={{ position: "absolute", inset: 0, width: "100%", height: "100%", objectFit: "cover", opacity: 0.3 }} />
        <div style={{ position: "absolute", inset: 0, background: `linear-gradient(180deg, ${C.overlay}99 0%, ${C.overlay}CC 60%, ${C.overlay} 100%)` }} />
        <SubjectMotifs />
        <div style={{ position: "relative", maxWidth: 700, margin: "0 auto" }}>
          <span style={{ display: "inline-flex", padding: "6px 14px", borderRadius: 999, border: `1px solid ${C.teal}66`, background: `${C.teal}22`, fontSize: 12.5, fontWeight: 700, color: C.teal, fontFamily: FH, marginBottom: 18 }}>
            {t("nav.about")}
          </span>
          <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(28px,4vw,42px)", color: "#fff", marginBottom: 16, letterSpacing: "-.02em", lineHeight: 1.15 }}>
            {t({ ru: "Национальная платформа образования Таджикистана", tg: "Платформаи миллии маорифи Тоҷикистон" })}
          </h1>
          <p style={{ fontSize: 15, color: "rgba(255,255,255,.75)", lineHeight: 1.7 }}>
            {t({
              ru: "Сегодня выбор сада, школы или вуза — это чаты, соцсети и советы знакомых. Мы строим EduHub, чтобы у каждой семьи в Таджикистане был один надёжный каталог с объективным рейтингом и честными отзывами.",
              tg: "Имрӯз интихоби боғча, мактаб ё донишгоҳ — ин чатҳо, шабакаҳои иҷтимоӣ ва маслиҳати шиносҳо. Мо EduHub-ро месозем, то ҳар оила дар Тоҷикистон як каталоги боэътимод бо рейтинги объективӣ ва шарҳҳои ростқавлона дошта бошад.",
            })}
          </p>
        </div>
      </div>

      {/* ── PROBLEM ── */}
      <div style={{ maxWidth: 1000, margin: "0 auto", padding: "56px 28px 0" }}>
        <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 22, color: C.text, marginBottom: 24 }}>
          {t({ ru: "Проблема, которую мы решаем", tg: "Мушкиле, ки мо ҳал мекунем" })}
        </h2>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3,1fr)", gap: 16 }}>
          {PROBLEMS.map((p) => (
            <div key={p.title.ru} style={{ borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: 22 }}>
              <p.icon size={20} style={{ color: C.red, marginBottom: 12 }} />
              <h3 style={{ fontFamily: FH, fontWeight: 700, fontSize: 15, color: C.text, marginBottom: 8 }}>{t(p.title)}</h3>
              <p style={{ fontSize: 13, color: C.sub, lineHeight: 1.6 }}>{t(p.desc)}</p>
            </div>
          ))}
        </div>
      </div>

      {/* ── PRODUCT SHOWCASE ── */}
      <div style={{ maxWidth: 1000, margin: "0 auto", padding: "56px 28px 0" }}>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 36, alignItems: "center" }}>
          <div>
            <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 22, color: C.text, marginBottom: 12 }}>
              {t({ ru: "Одно место вместо десятка вкладок", tg: "Як ҷой ба ҷои даҳ варақа" })}
            </h2>
            <p style={{ fontSize: 14, color: C.sub, lineHeight: 1.7 }}>
              {t({
                ru: "Фильтры по району, цене и программе обучения. Рейтинг по 8 объективным метрикам. Отзывы только от родителей с подтверждённой связью с учреждением. Прямой чат — без звонков и ожидания.",
                tg: "Филтрҳо аз рӯи ноҳия, нарх ва барномаи таълим. Рейтинг аз рӯи 8 меъёри объективӣ. Шарҳҳо танҳо аз волидайне, ки алоқаи тасдиқшуда доранд. Чати бевосита — бе занг ва интизорӣ.",
              })}
            </p>
          </div>
          <div style={{ borderRadius: 16, overflow: "hidden", border: `1px solid ${C.border}`, boxShadow: "0 20px 50px rgba(0,0,0,.35)" }}>
            <img src="/guide/parents-1-search.png" alt="" style={{ width: "100%", display: "block" }} />
          </div>
        </div>
      </div>

      {/* ── HOW RATING WORKS ── */}
      <div style={{ maxWidth: 1000, margin: "0 auto", padding: "56px 28px 0" }}>
        <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 22, color: C.text, marginBottom: 6, display: "flex", alignItems: "center", gap: 9 }}>
          <Target size={20} style={{ color: C.teal }} /> {t({ ru: "Как считается рейтинг", tg: "Рейтинг чӣ гуна ҳисоб карда мешавад" })}
        </h2>
        <p style={{ fontSize: 14, color: C.sub, marginBottom: 24, maxWidth: 640 }}>
          {t({ ru: "Каждое учреждение оценивается родителями и студентами по объективным категориям — не общий балл «нравится/не нравится», а разбивка по конкретным критериям.", tg: "Ҳар муассиса аз ҷониби волидайн ва донишҷӯён аз рӯи категорияҳои объективӣ баҳогузорӣ мешавад — на балли умумии «маъқул/номаъқул», балки тақсимот аз рӯи меъёрҳои мушаххас." })}
        </p>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(4,1fr)", gap: 12 }}>
          {RATING_CATEGORIES.map((c) => (
            <div key={c.ru} style={{ borderRadius: 12, border: `1px solid ${C.border}`, background: C.s1, padding: "14px 16px", fontSize: 13.5, fontWeight: 600, color: C.text, fontFamily: FH }}>
              {t(c)}
            </div>
          ))}
        </div>
      </div>

      {/* ── TRANSPARENCY ── */}
      <div style={{ maxWidth: 1000, margin: "0 auto", padding: "48px 28px 0" }}>
        <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 28, display: "flex", gap: 18, alignItems: "flex-start" }}>
          <div style={{ width: 44, height: 44, borderRadius: 12, background: `${C.ok}18`, display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
            <ShieldCheck size={21} style={{ color: C.ok }} />
          </div>
          <div>
            <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 17, color: C.text, marginBottom: 8 }}>
              {t({ ru: "Как мы модерируем отзывы", tg: "Мо шарҳҳоро чӣ гуна модератсия мекунем" })}
            </h3>
            <p style={{ fontSize: 14, color: C.sub, lineHeight: 1.65 }}>
              {t({
                ru: "Отзывы проверяются на спам и оскорбления, но не редактируются по содержанию и не удаляются по просьбе учреждения. Учреждение может публично ответить на отзыв в своём кабинете — это тоже видно всем.",
                tg: "Шарҳҳо аз назари спам ва таҳқир санҷида мешаванд, аммо аз рӯи мазмун таҳрир намешаванд ва бо дархости муассиса нест намешаванд. Муассиса метавонад ба шарҳ дар кабинети худ оммавӣ ҷавоб диҳад — ин низ ба ҳама намоён аст.",
              })}
            </p>
          </div>
        </div>
      </div>

      {/* ── TRUST PILLARS ── */}
      <div style={{ maxWidth: 1000, margin: "0 auto", padding: "48px 28px 0" }}>
        <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 22, color: C.text, marginBottom: 6 }}>
          {t({ ru: "Почему нам можно верить", tg: "Чаро ба мо бовар кардан мумкин аст" })}
        </h2>
        <p style={{ fontSize: 14, color: C.sub, marginBottom: 20, maxWidth: 640 }}>
          {t({ ru: "Рейтинг нельзя купить: платный тариф учреждения влияет только на видимость в поиске, но никогда — на значение оценки.", tg: "Рейтингро харидан мумкин нест: тарифи пулакии муассиса танҳо ба намоён будан таъсир мерасонад, аммо ҳеҷ гоҳ — ба арзиши баҳо." })}
        </p>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3,1fr)", gap: 14 }}>
          {TRUST_PILLARS.map((p) => (
            <div key={p.title.ru} style={{ borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: 22 }}>
              <div style={{ width: 40, height: 40, borderRadius: 11, background: `${C.teal}18`, display: "flex", alignItems: "center", justifyContent: "center", marginBottom: 12 }}>
                <p.icon size={18} style={{ color: C.teal }} />
              </div>
              <h3 style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text, marginBottom: 6 }}>{t(p.title)}</h3>
              <p style={{ fontSize: 12.5, color: C.sub, lineHeight: 1.55 }}>{t(p.desc)}</p>
            </div>
          ))}
        </div>
      </div>

      {/* ── REGIONS ── */}
      <div style={{ maxWidth: 1000, margin: "0 auto", padding: "48px 28px 0" }}>
        <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 22, color: C.text, marginBottom: 6, display: "flex", alignItems: "center", gap: 9 }}>
          <MapPin size={20} style={{ color: C.teal }} /> {t({ ru: "География", tg: "Ҷуғрофия" })}
        </h2>
        <p style={{ fontSize: 14, color: C.sub, marginBottom: 20, maxWidth: 640 }}>
          {t({ ru: "Начали с Душанбе, расширяемся по регионам Таджикистана.", tg: "Аз Душанбе оғоз кардем, дар минтақаҳои Тоҷикистон паҳн мешавем." })}
        </p>
        <div style={{ display: "flex", gap: 10, flexWrap: "wrap" }}>
          {REGION_ORDER.map((r) => (
            <span key={r} style={{ padding: "8px 16px", borderRadius: 999, border: `1px solid ${C.border}`, background: C.s1, fontSize: 13.5, fontWeight: 600, color: C.text, fontFamily: FH }}>
              {t(REGION_LABEL[r])}
            </span>
          ))}
        </div>
      </div>

      {/* ── AUDIENCE LINKS ── */}
      <div style={{ maxWidth: 1000, margin: "0 auto", padding: "48px 28px 0" }}>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3,1fr)", gap: 14 }}>
          {AUDIENCE_LINKS.map((a) => (
            <Link key={a.href} href={a.href} style={{ display: "flex", alignItems: "center", gap: 10, borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "16px 18px", textDecoration: "none" }}>
              <a.icon size={18} style={{ color: C.teal, flexShrink: 0 }} />
              <span style={{ flex: 1, fontSize: 13.5, fontWeight: 700, color: C.text, fontFamily: FH }}>{t(a.label)}</span>
              <ArrowRight size={14} style={{ color: C.dim }} />
            </Link>
          ))}
        </div>
      </div>

      {/* ── CONTACT / FEEDBACK FORM ── */}
      <div style={{ maxWidth: 560, margin: "0 auto", padding: "56px 28px 80px" }}>
        <div style={{ textAlign: "center", marginBottom: 24 }}>
          <Mail size={24} style={{ color: C.teal, marginBottom: 12 }} />
          <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text, marginBottom: 8 }}>
            {t({ ru: "Есть вопрос, которого нет в гайдах?", tg: "Саволе доред, ки дар дастурҳо нест?" })}
          </h2>
          <p style={{ fontSize: 13.5, color: C.sub }}>
            {t({ ru: "Напишите нам напрямую — отвечаем в течение 1-2 рабочих дней.", tg: "Мустақим ба мо нависед — дар давоми 1-2 рӯзи корӣ ҷавоб медиҳем." })}
          </p>
        </div>
        <ContactForm />
        <div style={{ textAlign: "center", marginTop: 20, display: "flex", flexDirection: "column", gap: 6 }}>
          <a href="mailto:info@eduhub.tj" style={{ fontSize: 13, color: C.dim, fontFamily: FH, textDecoration: "none" }}>info@eduhub.tj</a>
          <Link href="/about" style={{ fontSize: 13, color: C.dim, fontFamily: FH, textDecoration: "none" }}>
            {t({ ru: "Вы инвестор или партнёр?", tg: "Шумо сармоягузор ё шарик ҳастед?" })} →
          </Link>
        </div>
      </div>
    </div>
  );
}
