"use client";

import { useState } from "react";
import Link from "next/link";
import { Target, ShieldCheck, MapPin, Mail, ArrowRight, Users, Building2, Briefcase, MessagesSquare, Wallet, SearchX, UserCheck, BadgeCheck, Send, CheckCircle2, Languages, Heart, Eye, Search, Scale, Lock } from "lucide-react";
import { C, FH, FB, PHOTOS, REGION_ORDER, REGION_LABEL } from "@/lib/data";
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
  const [touched, setTouched] = useState(false);
  const [sent, setSent] = useState(false);

  const textError = touched && !text.trim();

  function submit(e: React.FormEvent) {
    e.preventDefault();
    setTouched(true);
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
    <form onSubmit={submit} noValidate style={{ position: "relative", borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 24, display: "flex", flexDirection: "column", gap: 14 }}>
      <input value={honeypot} onChange={(e) => setHoneypot(e.target.value)} name="website" tabIndex={-1} autoComplete="off" style={{ position: "absolute", left: -9999, width: 1, height: 1, opacity: 0 }} aria-hidden="true" />
      <div>
        <label style={{ display: "block", fontSize: 11.5, fontWeight: 700, color: C.sub, textTransform: "uppercase", letterSpacing: ".05em", fontFamily: FH, marginBottom: 6 }}>
          {t({ ru: "Кто вы?", tg: "Шумо кистед?" })}
        </label>
        <select value={topic} onChange={(e) => setTopic(e.target.value)} style={{ width: "100%", padding: "10px 13px", borderRadius: 10, border: `1px solid ${C.border}`, background: C.s2, color: C.text, fontFamily: FB, fontSize: 13.5, outline: "none", boxSizing: "border-box" }}>
          {TOPICS.map((o) => <option key={o.v} value={o.v}>{t({ ru: o.ru, tg: o.tg })}</option>)}
        </select>
      </div>
      <div>
        <label style={{ display: "block", fontSize: 11.5, fontWeight: 700, color: C.sub, textTransform: "uppercase", letterSpacing: ".05em", fontFamily: FH, marginBottom: 6 }}>
          {t({ ru: "Ваш вопрос", tg: "Саволи шумо" })}
        </label>
        <textarea value={text} onChange={(e) => setText(e.target.value)} onBlur={() => setTouched(true)} rows={4}
          style={{ width: "100%", padding: "10px 13px", borderRadius: 10, border: `1px solid ${textError ? C.red : C.border}`, background: C.s2, color: C.text, fontFamily: FB, fontSize: 13.5, outline: "none", boxSizing: "border-box", resize: "vertical" }} />
        {textError && (
          <p style={{ fontSize: 12, color: C.red, marginTop: 6 }}>
            {t({ ru: "Напишите вопрос перед отправкой", tg: "Пеш аз фиристодан саволро нависед" })}
          </p>
        )}
      </div>
      <div>
        <label style={{ display: "block", fontSize: 11.5, fontWeight: 700, color: C.sub, textTransform: "uppercase", letterSpacing: ".05em", fontFamily: FH, marginBottom: 6 }}>
          {t({ ru: "Email или телефон (необязательно)", tg: "Email ё телефон (ихтиёрӣ)" })}
        </label>
        <input value={contact} onChange={(e) => setContact(e.target.value)} style={{ width: "100%", padding: "10px 13px", borderRadius: 10, border: `1px solid ${C.border}`, background: C.s2, color: C.text, fontFamily: FB, fontSize: 13.5, outline: "none", boxSizing: "border-box" }} />
      </div>
      <button type="submit" style={{ display: "flex", alignItems: "center", justifyContent: "center", gap: 8, padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer" }}>
        <Send size={15} /> {t({ ru: "Отправить", tg: "Фиристодан" })}
      </button>
    </form>
  );
}

const PROBLEMS = [
  { icon: SearchX, title: { ru: "Информация разбросана", tg: "Маълумот пароканда аст" }, desc: { ru: "Instagram школы, чужие отзывы в VK-группах, советы знакомых — вместо одного места для сравнения.", tg: "Instagram-и мактаб, шарҳҳо дар гурӯҳҳои VK, маслиҳати шиносон — ба ҷои як ҷои муқоиса." } },
  { icon: MessagesSquare, title: { ru: "Отзывам сложно доверять", tg: "Ба шарҳҳо бовар кардан душвор аст" }, desc: { ru: "Кто угодно может написать что угодно — без проверки, что человек вообще имеет отношение к школе.", tg: "Ҳар кас метавонад чизе нависад — бе санҷиши он, ки шахс воқеан алоқаманд аст." } },
  { icon: Wallet, title: { ru: "Полная стоимость всплывает потом", tg: "Арзиши пурра баъдтар зоҳир мешавад" }, desc: { ru: "Развозка, питание, кружки, форма — родитель узнаёт сумму целиком уже после зачисления.", tg: "Расонидан, ғизо, машғулиятҳо, либос — волидайн маблағро пас аз қабул мефаҳмад." } },
];

const PRINCIPLES = [
  { icon: Eye, title: { ru: "Верификация превыше всего", tg: "Тасдиқ бештар аз ҳама" }, desc: { ru: "Меньше отзывов, но проверенных — доверие важнее объёма.", tg: "Камтар шарҳ, аммо тасдиқшуда — эътимод муҳимтар аз ҳаҷм." } },
  { icon: Wallet, title: { ru: "Полная финансовая прозрачность", tg: "Шаффофияти пурраи молиявӣ" }, desc: { ru: "Цена, развозка, питание, скидки — видны в профиле сразу.", tg: "Нарх, расонидан, ғизо, тахфифҳо — дарҳол дар профил намоён." } },
  { icon: Languages, title: { ru: "Локально для Таджикистана", tg: "Маҳаллӣ барои Тоҷикистон" }, desc: { ru: "Таджикский и русский с начала, все 5 регионов страны.", tg: "Тоҷикӣ ва русӣ аз ибтидо, ҳамаи 5 минтақаи кишвар." } },
  { icon: Heart, title: { ru: "Бесплатно для семей", tg: "Ройгон барои оилаҳо" }, desc: { ru: "Поиск, отзывы, чат — родитель не платит ни за что.", tg: "Ҷустуҷӯ, шарҳҳо, чат — волидайн барои ҳеҷ чиз пул намедиҳад." } },
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
  { icon: UserCheck, title: { ru: "Привязка к ребёнку", tg: "Алоқаманд ба фарзанд" }, desc: { ru: "Отзыв оставляет только родитель ребёнка, который учится в учреждении, закончил его или перевёлся в другое — связь должна быть реальной, а не обязательно текущей", tg: "Шарҳро танҳо волидайне мегузорад, ки фарзандаш дар муассиса таҳсил мекунад, онро хатм кардааст ё ба муассисаи дигар гузаштааст — алоқа бояд воқеӣ бошад, на ҳатман феълӣ" } },
  { icon: BadgeCheck, title: { ru: "Верификация владельца профиля", tg: "Тасдиқи соҳиби профил" }, desc: { ru: "Представитель учреждения подтверждает полномочия документом до публикации", tg: "Намояндаи муассиса пеш аз нашр ваколати худро бо ҳуҷҷат тасдиқ мекунад" } },
  { icon: Scale, title: { ru: "Справедливый разбор спора", tg: "Баррасии одилонаи баҳс" }, desc: { ru: "Учреждение может оспорить отзыв — модератор отвечает не позже чем за 72 часа", tg: "Муассиса метавонад шарҳро баҳс кунад — модератор на дертар аз 72 соат ҷавоб медиҳад" } },
  { icon: Lock, title: { ru: "Приватность соискателя", tg: "Махфияти ҷуяндаи кор" }, desc: { ru: "Контакты педагога скрыты от учреждения, пока сам не ответит в чате", tg: "Тамосҳои омӯзгор аз муассиса пинҳон аст, то худаш дар чат ҷавоб надиҳад" } },
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
      <style jsx>{`
        .eh-grid-3 { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; }
        .eh-grid-4 { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px; }
        .eh-grid-5 { display: grid; grid-template-columns: repeat(3, 1fr); gap: 14px; }
        .eh-grid-rating { display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px; }
        .eh-showcase { display: grid; grid-template-columns: 1fr 1fr; gap: 36px; align-items: center; }
        .eh-philosophy { display: grid; grid-template-columns: 1fr 1.2fr; gap: 40px; }
        @media (max-width: 760px) {
          .eh-grid-3, .eh-grid-4, .eh-grid-5 { grid-template-columns: 1fr; }
          .eh-grid-rating { grid-template-columns: repeat(2, 1fr); }
          .eh-showcase, .eh-philosophy { grid-template-columns: 1fr; }
        }
      `}</style>

      {/* ── HERO ── */}
      <div style={{ position: "relative", overflow: "hidden", padding: "72px 28px 64px", textAlign: "center" }}>
        <img src={PHOTOS.campus2} alt="" style={{ position: "absolute", inset: 0, width: "100%", height: "100%", objectFit: "cover", opacity: 0.3 }} />
        <div style={{ position: "absolute", inset: 0, background: `linear-gradient(180deg, ${C.overlay}99 0%, ${C.overlay}CC 60%, ${C.overlay} 100%)` }} />
        <div style={{ position: "relative", maxWidth: 680, margin: "0 auto" }}>
          <span style={{ display: "inline-flex", padding: "6px 14px", borderRadius: 999, border: `1px solid ${C.teal}66`, background: `${C.teal}22`, fontSize: 12.5, fontWeight: 700, color: C.teal, fontFamily: FH, marginBottom: 18 }}>
            {t("nav.about")}
          </span>
          <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(28px,4vw,42px)", color: "#fff", marginBottom: 16, letterSpacing: "-.02em", lineHeight: 1.15 }}>
            {t({ ru: "Национальная платформа образования Таджикистана", tg: "Платформаи миллии маорифи Тоҷикистон" })}
          </h1>
          <p style={{ fontSize: 15, color: "rgba(255,255,255,.75)", lineHeight: 1.7, maxWidth: 560, margin: "0 auto 28px" }}>
            {t({
              ru: "Один надёжный каталог с объективным рейтингом и честными отзывами — вместо чатов, соцсетей и советов знакомых.",
              tg: "Як каталоги боэътимод бо рейтинги объективӣ ва шарҳҳои ростқавлона — ба ҷои чатҳо, шабакаҳо ва маслиҳати шиносҳо.",
            })}
          </p>
          <div style={{ display: "flex", gap: 12, justifyContent: "center", flexWrap: "wrap" }}>
            <Link href="/search" style={{ display: "flex", alignItems: "center", gap: 8, padding: "13px 26px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, textDecoration: "none" }}>
              <Search size={15} /> {t({ ru: "Смотреть каталог", tg: "Каталогро дидан" })}
            </Link>
            <a href="#proof" style={{ display: "flex", alignItems: "center", gap: 8, padding: "13px 22px", borderRadius: 11, background: "rgba(255,255,255,.08)", border: "1px solid rgba(255,255,255,.24)", color: "#fff", fontFamily: FH, fontWeight: 700, fontSize: 14, textDecoration: "none" }}>
              <ShieldCheck size={15} style={{ color: C.teal }} /> {t({ ru: "Как мы проверяем", tg: "Мо чӣ гуна санҷем" })}
            </a>
          </div>
        </div>
      </div>

      {/* ── ZONE 1: PHILOSOPHY — почему мы это делаем ── */}
      <div style={{ maxWidth: 1000, margin: "0 auto", padding: "64px 28px 0" }}>
        <h2 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(20px,2.4vw,26px)", color: C.text, marginBottom: 32, letterSpacing: "-.02em" }}>
          {t({ ru: "Почему мы это делаем", tg: "Чаро мо инро мекунем" })}
        </h2>
        <div className="eh-philosophy">
          <div>
            <p style={{ fontSize: 11.5, fontWeight: 700, color: C.dim, textTransform: "uppercase", letterSpacing: ".06em", fontFamily: FH, marginBottom: 14 }}>
              {t({ ru: "Боль, которую мы видим", tg: "Дарде, ки мо мебинем" })}
            </p>
            <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
              {PROBLEMS.map((p) => (
                <div key={p.title.ru} style={{ display: "flex", gap: 14 }}>
                  <p.icon size={18} style={{ color: C.red, flexShrink: 0, marginTop: 2 }} />
                  <div>
                    <h3 style={{ fontFamily: FH, fontWeight: 700, fontSize: 14.5, color: C.text, marginBottom: 4 }}>{t(p.title)}</h3>
                    <p style={{ fontSize: 13, color: C.sub, lineHeight: 1.6, maxWidth: 380 }}>{t(p.desc)}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
          <div>
            <p style={{ fontSize: 11.5, fontWeight: 700, color: C.dim, textTransform: "uppercase", letterSpacing: ".06em", fontFamily: FH, marginBottom: 14 }}>
              {t({ ru: "Во что мы верим", tg: "Ба чӣ бовар дорем" })}
            </p>
            <div className="eh-grid-5" style={{ gridTemplateColumns: "repeat(2,1fr)" }}>
              {PRINCIPLES.map((p) => (
                <div key={p.title.ru} style={{ borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: 16 }}>
                  <div style={{ width: 32, height: 32, borderRadius: 9, background: `${C.teal}18`, display: "flex", alignItems: "center", justifyContent: "center", marginBottom: 10 }}>
                    <p.icon size={15} style={{ color: C.teal }} />
                  </div>
                  <h3 style={{ fontFamily: FH, fontWeight: 700, fontSize: 12.5, color: C.text, marginBottom: 5 }}>{t(p.title)}</h3>
                  <p style={{ fontSize: 11.5, color: C.sub, lineHeight: 1.5 }}>{t(p.desc)}</p>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* ── ZONE 2: MECHANICS — как это работает (нейтрально, табличный стиль) ── */}
      <div style={{ maxWidth: 1000, margin: "0 auto", padding: "72px 28px 0" }}>
        <h2 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(20px,2.4vw,26px)", color: C.text, marginBottom: 32, letterSpacing: "-.02em" }}>
          {t({ ru: "Как это работает", tg: "Ин чӣ гуна кор мекунад" })}
        </h2>

        <div className="eh-showcase" style={{ marginBottom: 48 }}>
          <div>
            <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 17, color: C.text, marginBottom: 10 }}>
              {t({ ru: "Одно место вместо десятка вкладок", tg: "Як ҷой ба ҷои даҳ варақа" })}
            </h3>
            <p style={{ fontSize: 13.5, color: C.sub, lineHeight: 1.7, maxWidth: 400 }}>
              {t({
                ru: "Фильтры по району, цене и программе. Рейтинг по 8 метрикам. Отзывы только от родителей с подтверждённой связью. Прямой чат — без звонков.",
                tg: "Филтрҳо аз рӯи ноҳия, нарх ва барнома. Рейтинг аз рӯи 8 меъёр. Шарҳҳо танҳо аз волидайни тасдиқшуда. Чати бевосита.",
              })}
            </p>
          </div>
          <div style={{ borderRadius: 16, overflow: "hidden", border: `1px solid ${C.border}`, boxShadow: "0 20px 50px rgba(0,0,0,.35)" }}>
            <img src="/guide/parents-1-search.png" alt="" style={{ width: "100%", display: "block" }} />
          </div>
        </div>

        <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 28, marginBottom: 24 }}>
          <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 15.5, color: C.text, marginBottom: 4, display: "flex", alignItems: "center", gap: 8 }}>
            <Target size={17} style={{ color: C.teal }} /> {t({ ru: "8 метрик рейтинга, не общий балл", tg: "8 меъёри рейтинг, на балли умумӣ" })}
          </h3>
          <p style={{ fontSize: 13, color: C.sub, marginBottom: 20, maxWidth: 560 }}>
            {t({ ru: "Не «нравится/не нравится» — разбивка по конкретным категориям, которые родитель выбирает сам.", tg: "На «маъқул/номаъқул» — тақсимот аз рӯи категорияҳои мушаххас." })}
          </p>
          <div className="eh-grid-rating">
            {RATING_CATEGORIES.map((c, i) => (
              <div key={c.ru} style={{ display: "flex", alignItems: "center", gap: 10, borderRadius: 10, background: C.s2, padding: "10px 12px" }}>
                <span style={{ fontFamily: FH, fontWeight: 800, fontSize: 11, color: C.teal, fontVariantNumeric: "tabular-nums", flexShrink: 0 }}>{String(i + 1).padStart(2, "0")}</span>
                <span style={{ fontSize: 12.5, fontWeight: 600, color: C.text, fontFamily: FH }}>{t(c)}</span>
              </div>
            ))}
          </div>
        </div>

        <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: "20px 28px", display: "flex", alignItems: "center", justifyContent: "space-between", gap: 20, flexWrap: "wrap" }}>
          <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
            <MapPin size={18} style={{ color: C.teal }} />
            <div>
              <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: C.text }}>{t({ ru: "География", tg: "Ҷуғрофия" })}</p>
              <p style={{ fontSize: 12, color: C.sub }}>{t({ ru: "Начали с Душанбе, расширяемся по регионам", tg: "Аз Душанбе оғоз, дар минтақаҳо паҳн мешавем" })}</p>
            </div>
          </div>
          <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
            {REGION_ORDER.map((r) => (
              <span key={r} style={{ padding: "6px 13px", borderRadius: 999, border: `1px solid ${C.border}`, background: C.s2, fontSize: 12.5, fontWeight: 600, color: C.text, fontFamily: FH }}>
                {t(REGION_LABEL[r])}
              </span>
            ))}
          </div>
        </div>
      </div>

      {/* ── ZONE 3: PROOF — почему нам можно верить (доминирующая, фото-фон) ── */}
      <div id="proof" style={{ position: "relative", overflow: "hidden", marginTop: 72 }}>
        <img src={PHOTOS.classroom} alt="" style={{ position: "absolute", inset: 0, width: "100%", height: "100%", objectFit: "cover" }} />
        <div style={{ position: "absolute", inset: 0, background: `linear-gradient(180deg, ${C.overlay}55 0%, ${C.overlay}D8 25%, ${C.overlay} 100%)` }} />
        <div style={{ position: "relative", maxWidth: 1000, margin: "0 auto", padding: "64px 28px 72px" }}>
          <span style={{ display: "inline-flex", padding: "5px 12px", borderRadius: 999, border: `1px solid ${C.teal}66`, background: `${C.teal}22`, fontSize: 12, fontWeight: 700, color: C.teal, fontFamily: FH, marginBottom: 16 }}>
            {t({ ru: "Рейтинг нельзя купить", tg: "Рейтингро харидан мумкин нест" })}
          </span>
          <h2 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(22px,3vw,30px)", color: "#fff", marginBottom: 10, letterSpacing: "-.02em", maxWidth: 620 }}>
            {t({ ru: "Почему нам можно верить", tg: "Чаро ба мо бовар кардан мумкин аст" })}
          </h2>
          <p style={{ fontSize: 14.5, color: "rgba(255,255,255,.8)", marginBottom: 36, maxWidth: 580, lineHeight: 1.65 }}>
            {t({ ru: "Платный тариф учреждения влияет только на видимость в поиске, но никогда — на значение оценки. Это касается всех трёх сторон платформы.", tg: "Тарифи пулакии муассиса танҳо ба намоён будан таъсир мерасонад, аммо ҳеҷ гоҳ — ба арзиши баҳо. Ин ба ҳар се тарафи платформа дахл дорад." })}
          </p>
          <div className="eh-grid-5" style={{ marginBottom: 28 }}>
            {TRUST_PILLARS.map((p) => (
              <div key={p.title.ru} style={{ borderRadius: 16, border: "1px solid rgba(255,255,255,.16)", background: "rgba(255,255,255,.07)", backdropFilter: "blur(10px)", padding: 20 }}>
                <div style={{ width: 38, height: 38, borderRadius: 11, background: `${C.teal}30`, display: "flex", alignItems: "center", justifyContent: "center", marginBottom: 12 }}>
                  <p.icon size={17} style={{ color: C.teal }} />
                </div>
                <h3 style={{ fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: "#fff", marginBottom: 6 }}>{t(p.title)}</h3>
                <p style={{ fontSize: 12, color: "rgba(255,255,255,.72)", lineHeight: 1.55 }}>{t(p.desc)}</p>
              </div>
            ))}
          </div>

          <div style={{ borderRadius: 16, border: "1px solid rgba(255,255,255,.16)", background: "rgba(255,255,255,.07)", backdropFilter: "blur(10px)", padding: 24, display: "flex", gap: 16, alignItems: "flex-start", marginBottom: 32 }}>
            <ShieldCheck size={20} style={{ color: C.teal, flexShrink: 0, marginTop: 2 }} />
            <div>
              <h3 style={{ fontFamily: FH, fontWeight: 700, fontSize: 14.5, color: "#fff", marginBottom: 6 }}>
                {t({ ru: "Как мы модерируем отзывы", tg: "Мо шарҳҳоро чӣ гуна модератсия мекунем" })}
              </h3>
              <p style={{ fontSize: 13, color: "rgba(255,255,255,.75)", lineHeight: 1.6 }}>
                {t({
                  ru: "Отзывы проверяются на спам и оскорбления, но не редактируются по содержанию и не удаляются по просьбе учреждения. Учреждение отвечает публично в своём кабинете — это тоже видно всем.",
                  tg: "Шарҳҳо аз назари спам ва таҳқир санҷида мешаванд, аммо аз рӯи мазмун таҳрир намешаванд ва бо дархости муассиса нест намешаванд. Муассиса дар кабинети худ оммавӣ ҷавоб медиҳад.",
                })}
              </p>
            </div>
          </div>

          <div style={{ display: "flex", gap: 12, flexWrap: "wrap" }}>
            <Link href="/search" style={{ display: "flex", alignItems: "center", gap: 8, padding: "13px 26px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, textDecoration: "none" }}>
              <Search size={15} /> {t({ ru: "Смотреть каталог", tg: "Каталогро дидан" })}
            </Link>
          </div>
        </div>
      </div>

      {/* ── AUDIENCE LINKS ── */}
      <div style={{ maxWidth: 1000, margin: "0 auto", padding: "48px 28px 0" }}>
        <div className="eh-grid-5">
          {AUDIENCE_LINKS.map((a) => (
            <Link key={a.href} href={a.href} style={{ display: "flex", alignItems: "center", gap: 10, borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "16px 18px", textDecoration: "none" }}>
              <a.icon size={18} style={{ color: C.teal, flexShrink: 0 }} />
              <span style={{ flex: 1, fontSize: 13.5, fontWeight: 700, color: C.text, fontFamily: FH }}>{t(a.label)}</span>
              <ArrowRight size={14} style={{ color: C.dim }} />
            </Link>
          ))}
        </div>
      </div>

      {/* ── CONTACT / FEEDBACK FORM (фото-фон под заголовком, форма — светлая карточка) ── */}
      <div style={{ position: "relative", overflow: "hidden", marginTop: 56 }}>
        <img src={PHOTOS.library} alt="" style={{ position: "absolute", inset: 0, width: "100%", height: "100%", objectFit: "cover" }} />
        <div style={{ position: "absolute", inset: 0, background: `linear-gradient(180deg, ${C.overlay}66 0%, ${C.overlay}D8 55%, ${C.overlay} 100%)` }} />
        <div style={{ position: "relative", textAlign: "center", padding: "56px 28px 8px" }}>
          <Mail size={24} style={{ color: C.teal, marginBottom: 12 }} />
          <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: "#fff", marginBottom: 8 }}>
            {t({ ru: "Есть вопрос, которого нет в гайдах?", tg: "Саволе доред, ки дар дастурҳо нест?" })}
          </h2>
          <p style={{ fontSize: 13.5, color: "rgba(255,255,255,.78)" }}>
            {t({ ru: "Напишите нам напрямую — отвечаем в течение 1-2 рабочих дней.", tg: "Мустақим ба мо нависед — дар давоми 1-2 рӯзи корӣ ҷавоб медиҳем." })}
          </p>
        </div>
      </div>
      <div style={{ maxWidth: 560, margin: "0 auto", padding: "24px 28px 80px" }}>
        <ContactForm />
        <div style={{ textAlign: "center", marginTop: 20 }}>
          <a href="mailto:info@eduhub.tj" style={{ fontSize: 13, color: C.sub, fontFamily: FH, textDecoration: "none" }}>info@eduhub.tj</a>
        </div>
      </div>
    </div>
  );
}
