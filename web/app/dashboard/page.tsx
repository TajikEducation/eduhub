"use client";

import { useState } from "react";
import Link from "next/link";
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from "recharts";
import { useRouter } from "next/navigation";
import {
  LayoutDashboard, Newspaper, Users, Star, Settings, Plus, Pencil, Trash2, X, LogOut, Briefcase, Award, UserSquare2, Camera, ArrowLeft, Clock, type LucideIcon,
} from "lucide-react";
import {
  C, FH, FB, INSTITUTIONS, ALL_STAFF, REVIEWS, NEWS_ITEMS, VACANCIES, VACANCY_CANDIDATES, SEED_APPLICANTS, PHOTOS, CATEGORY_META, REGION_LABEL, REGION_ORDER,
  type Person, type NewsArticle, type Vacancy, type Achievement, type Alumnus, type Bi, type Region, type Institution,
} from "@/lib/data";
import { Modal } from "@/components/ui/Modal";
import { Toast } from "@/components/Toast";
import { useAppState } from "@/lib/app-state";
import { useT } from "@/lib/i18n";

const DASH_INST_ID = 1;

type Tab = "overview" | "news" | "staff" | "vacancies" | "achievements" | "alumni" | "gallery" | "reviews" | "settings";
type SettingsTab = "info" | "contacts" | "hours";

const TAB_META: { k: Tab; labelKey: string; icon: LucideIcon }[] = [
  { k: "overview", labelKey: "dash.overview", icon: LayoutDashboard },
  { k: "news", labelKey: "tab.news", icon: Newspaper },
  { k: "staff", labelKey: "tab.staff", icon: Users },
  { k: "vacancies", labelKey: "tab.vacancies", icon: Briefcase },
  { k: "achievements", labelKey: "tab.achievements", icon: Award },
  { k: "alumni", labelKey: "tab.alumni", icon: UserSquare2 },
  { k: "gallery", labelKey: "tab.gallery", icon: Camera },
  { k: "reviews", labelKey: "tab.reviews", icon: Star },
  { k: "settings", labelKey: "dash.settings", icon: Settings },
];

const PHOTO_KEYS = Object.keys(PHOTOS) as (keyof typeof PHOTOS)[];
const bi = (ru: string, tg: string): Bi => ({ ru, tg });

const WEEK_TREND = [
  { day: "Пн", views: 42, responses: 3 },
  { day: "Вт", views: 58, responses: 5 },
  { day: "Ср", views: 51, responses: 4 },
  { day: "Чт", views: 67, responses: 7 },
  { day: "Пт", views: 74, responses: 6 },
  { day: "Сб", views: 39, responses: 2 },
  { day: "Вс", views: 33, responses: 3 },
];

const DEMO_FALLBACK = INSTITUTIONS.find((i) => i.id === DASH_INST_ID)!;

function PendingScreen({ inst, router, t }: { inst: Institution; router: ReturnType<typeof useRouter>; t: ReturnType<typeof useT> }) {
  return (
    <div style={{ display: "flex", flex: 1, alignItems: "center", justifyContent: "center", minHeight: "70vh", padding: 32 }}>
      <div style={{ maxWidth: 480, textAlign: "center" }}>
        <button onClick={() => (window.history.length > 1 ? router.back() : router.push("/"))} style={{ display: "inline-flex", alignItems: "center", gap: 7, fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: C.teal, marginBottom: 24, padding: "8px 14px", borderRadius: 9, border: `1px solid ${C.teal}40`, background: `${C.teal}10`, cursor: "pointer" }}>
          <ArrowLeft size={15} /> {t("common.back")}
        </button>
        <div style={{ width: 56, height: 56, borderRadius: 16, background: `${C.gold}18`, border: `1px solid ${C.gold}44`, display: "flex", alignItems: "center", justifyContent: "center", margin: "0 auto 18px" }}>
          <Clock size={26} style={{ color: C.gold }} />
        </div>
        <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 22, color: C.text, marginBottom: 10 }}>
          {t({ ru: "Заявка на рассмотрении", tg: "Ариза дар баррасӣ аст" })}
        </h1>
        <p style={{ fontSize: 14, color: C.sub, lineHeight: 1.6, marginBottom: 20 }}>
          {t({ ru: "Модератор проверяет данные учреждения «", tg: "Модератор маълумоти муассисаи «" })}{inst.name.ru}{t({ ru: "» — обычно это занимает до 48 часов. Как только заявка одобрена, профиль появится в общем каталоге и станет доступен весь функционал кабинета.", tg: "»-ро месанҷад — то 48 соат. Пас аз тасдиқ профил дар каталоги умумӣ пайдо мешавад." })}
        </p>
        <a href={`/institutions/${inst.id}`} style={{ display: "inline-flex", alignItems: "center", gap: 6, fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: C.teal, textDecoration: "none" }}>
          {t({ ru: "Предпросмотр профиля", tg: "Пешнамоиши профил" })} →
        </a>
      </div>
    </div>
  );
}

function RejectedScreen({ router, t }: { router: ReturnType<typeof useRouter>; t: ReturnType<typeof useT> }) {
  return (
    <div style={{ display: "flex", flex: 1, alignItems: "center", justifyContent: "center", minHeight: "70vh", padding: 32 }}>
      <div style={{ maxWidth: 440, textAlign: "center" }}>
        <button onClick={() => (window.history.length > 1 ? router.back() : router.push("/"))} style={{ display: "inline-flex", alignItems: "center", gap: 7, fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: C.teal, marginBottom: 24, padding: "8px 14px", borderRadius: 9, border: `1px solid ${C.teal}40`, background: `${C.teal}10`, cursor: "pointer" }}>
          <ArrowLeft size={15} /> {t("common.back")}
        </button>
        <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 22, color: C.text, marginBottom: 10 }}>
          {t({ ru: "Заявка отклонена", tg: "Ариза рад карда шуд" })}
        </h1>
        <p style={{ fontSize: 14, color: C.sub, lineHeight: 1.6, marginBottom: 20 }}>
          {t({ ru: "Модератор не смог подтвердить данные учреждения. Свяжитесь с нами, чтобы уточнить причину.", tg: "Модератор маълумоти муассисаро тасдиқ карда натавонист." })}
        </p>
        <a href="/company" style={{ display: "inline-flex", alignItems: "center", gap: 6, fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: C.teal, textDecoration: "none" }}>
          {t({ ru: "Написать нам", tg: "Ба мо нависед" })} →
        </a>
      </div>
    </div>
  );
}

export default function DashboardPage() {
  const { locale, myInstitution, applicant } = useAppState();
  const inst0 = myInstitution ?? DEMO_FALLBACK;
  const t = useT();
  const router = useRouter();
  const [tab, setTab] = useState<Tab>("overview");
  const [settingsTab, setSettingsTab] = useState<SettingsTab>("info");
  const [toast, setToast] = useState<string | null>(null);

  // ── news ──
  const [articles, setArticles] = useState<NewsArticle[]>(NEWS_ITEMS.filter((n) => n.instId === inst0.id));
  const [newsForm, setNewsForm] = useState<NewsArticle | null>(null);

  function newArticle() {
    setNewsForm({ id: `n${Date.now()}`, instId: DASH_INST_ID, title: bi("", ""), category: bi(t({ru:"Новость",tg:"Хабар"}), t({ru:"Новость",tg:"Хабар"})), coverUrl: PHOTOS.classroom2, content: bi("", ""), tags: [], date: new Date().toLocaleDateString("ru-RU"), status: "draft", views: 0 });
  }
  function saveArticle() {
    if (!newsForm || !newsForm.title[locale].trim()) return;
    setArticles((prev) => {
      const exists = prev.some((a) => a.id === newsForm.id);
      return exists ? prev.map((a) => (a.id === newsForm.id ? newsForm : a)) : [newsForm, ...prev];
    });
    setNewsForm(null);
    setToast(t({ru:"Новость сохранена",tg:"Хабар нигоҳ дошта шуд"}));
  }
  function deleteArticle(id: string) {
    setArticles((prev) => prev.filter((a) => a.id !== id));
  }

  // ── staff ──
  const [staffList, setStaffList] = useState<Person[]>(ALL_STAFF.filter((p) => p.instId === DASH_INST_ID));
  const [staffForm, setStaffForm] = useState<Person | null>(null);

  function newStaff() {
    setStaffForm({ id: `p${Date.now()}`, instId: DASH_INST_ID, name: bi("",""), roleLabel: bi("",""), roleType: "teacher", subject: bi("",""), photo: PHOTOS.pw1, exp: "", bio: bi("",""), education: [], achievements: [], email: "", phone: "" });
  }
  function saveStaff() {
    if (!staffForm || !staffForm.name[locale].trim()) return;
    setStaffList((prev) => {
      const exists = prev.some((p) => p.id === staffForm.id);
      return exists ? prev.map((p) => (p.id === staffForm.id ? staffForm : p)) : [...prev, staffForm];
    });
    setStaffForm(null);
    setToast(t({ru:"Сотрудник сохранён",tg:"Корманд нигоҳ дошта шуд"}));
  }
  function deleteStaff(id: string) {
    setStaffList((prev) => prev.filter((p) => p.id !== id));
  }

  // ── vacancies ──
  const [vacancyList, setVacancyList] = useState<Vacancy[]>(VACANCIES.filter((v) => v.instId === DASH_INST_ID));
  const [vacancyForm, setVacancyForm] = useState<Vacancy | null>(null);

  function newVacancy() {
    setVacancyForm({ id: `v${Date.now()}`, instId: DASH_INST_ID, title: bi("", ""), description: bi("", ""), requirements: [], employment: bi(t({ru:"Полная занятость",tg:"Шуғли пурра"}), t({ru:"Полная занятость",tg:"Шуғли пурра"})), date: new Date().toLocaleDateString("ru-RU"), status: "draft" });
  }
  function saveVacancy() {
    if (!vacancyForm || !vacancyForm.title[locale].trim()) return;
    setVacancyList((prev) => {
      const exists = prev.some((v) => v.id === vacancyForm.id);
      return exists ? prev.map((v) => (v.id === vacancyForm.id ? vacancyForm : v)) : [vacancyForm, ...prev];
    });
    setVacancyForm(null);
    setToast(t({ru:"Вакансия сохранена",tg:"Ҷои холӣ нигоҳ дошта шуд"}));
  }
  function deleteVacancy(id: string) {
    setVacancyList((prev) => prev.filter((v) => v.id !== id));
  }

  // ── achievements ──
  const [achievementsList, setAchievementsList] = useState<Achievement[]>(inst0.achievements);
  const [achievementForm, setAchievementForm] = useState<Achievement | null>(null);

  function newAchievement() {
    setAchievementForm({ id: `ach${Date.now()}`, title: bi("", ""), year: new Date().getFullYear(), category: "gold", desc: bi("", ""), type: "institution" });
  }
  function saveAchievement() {
    if (!achievementForm || !achievementForm.title[locale].trim()) return;
    setAchievementsList((prev) => {
      const exists = prev.some((a) => a.id === achievementForm.id);
      return exists ? prev.map((a) => (a.id === achievementForm.id ? achievementForm : a)) : [achievementForm, ...prev];
    });
    setAchievementForm(null);
    setToast(t({ru:"Достижение сохранено",tg:"Дастовард нигоҳ дошта шуд"}));
  }
  function deleteAchievement(id: string) {
    setAchievementsList((prev) => prev.filter((a) => a.id !== id));
  }

  // ── alumni ──
  const [alumniList, setAlumniList] = useState<Alumnus[]>(inst0.alumni ?? []);
  const [alumnusForm, setAlumnusForm] = useState<Alumnus | null>(null);

  function newAlumnus() {
    setAlumnusForm({ id: `al${Date.now()}`, name: bi("", ""), photo: PHOTOS.pw1, gradYear: new Date().getFullYear(), now: bi("", "") });
  }
  function saveAlumnus() {
    if (!alumnusForm || !alumnusForm.name[locale].trim()) return;
    setAlumniList((prev) => {
      const exists = prev.some((a) => a.id === alumnusForm.id);
      return exists ? prev.map((a) => (a.id === alumnusForm.id ? alumnusForm : a)) : [alumnusForm, ...prev];
    });
    setAlumnusForm(null);
    setToast(t({ru:"Выпускник сохранён",tg:"Хатмкунанда нигоҳ дошта шуд"}));
  }
  function deleteAlumnus(id: string) {
    setAlumniList((prev) => prev.filter((a) => a.id !== id));
  }

  // ── gallery ──
  const [galleryList, setGalleryList] = useState<{ url: string; label: Bi }[]>(inst0.gallery);
  const [galleryForm, setGalleryForm] = useState<{ url: string; label: Bi } | null>(null);

  function newGalleryItem() {
    setGalleryForm({ url: PHOTOS.classroom2, label: bi("", "") });
  }
  function saveGalleryItem() {
    if (!galleryForm || !galleryForm.label[locale].trim()) return;
    setGalleryList((prev) => (prev.some((g) => g.url === galleryForm.url) ? prev.map((g) => (g.url === galleryForm.url ? galleryForm : g)) : [galleryForm, ...prev]));
    setGalleryForm(null);
    setToast(t({ru:"Фото добавлено в галерею",tg:"Акс ба галерея илова шуд"}));
  }
  function deleteGalleryItem(url: string) {
    setGalleryList((prev) => prev.filter((g) => g.url !== url));
  }

  // ── reviews ──
  const [reviewsList, setReviewsList] = useState(REVIEWS.filter((r) => r.instId === DASH_INST_ID));
  const [replyDraftId, setReplyDraftId] = useState<string | null>(null);
  const [replyText, setReplyText] = useState("");

  function submitReply(id: string) {
    if (!replyText.trim()) return;
    setReviewsList((prev) => prev.map((r) => (r.id === id ? { ...r, reply: bi(replyText.trim(), replyText.trim()) } : r)));
    setReplyDraftId(null);
    setReplyText("");
    setToast(t({ru:"Ответ опубликован",tg:"Ҷавоб нашр шуд"}));
  }

  // ── settings ──
  const [info, setInfo] = useState({ name: inst0.name, description: inst0.description, area: inst0.area });
  const [region, setRegionField] = useState<Region>(inst0.region);
  const [contacts, setContacts] = useState({
    phone: inst0.phone, email: inst0.email, website: inst0.website, city: inst0.city, street: inst0.street,
    instagram: inst0.socials?.instagram ?? "", telegram: inst0.socials?.telegram ?? "", facebook: inst0.socials?.facebook ?? "",
  });
  const [hours, setHours] = useState([
    { labelKey: "hours.weekdays", time: "07:30 – 18:00" },
    { labelKey: "hours.saturday", time: "08:00 – 14:00" },
    { labelKey: "hours.sunday", time: t("hours.dayoff") },
  ]);

  function saveSettings(e: React.FormEvent) {
    e.preventDefault();
    setToast(t({ru:"Настройки сохранены",tg:"Танзимот нигоҳ дошта шуд"}));
  }

  if (myInstitution?.status === "pending") return <PendingScreen inst={myInstitution} router={router} t={t} />;
  if (myInstitution?.status === "rejected") return <RejectedScreen router={router} t={t} />;

  return (
    <div style={{ display: "flex", fontFamily: FB, background: C.bg, color: C.text }}>
      {/* ── SIDEBAR ── */}
      <aside style={{ width: 240, flexShrink: 0, borderRight: `1px solid ${C.border}`, padding: "20px 14px", display: "flex", flexDirection: "column" }}>
        <p style={{ padding: "0 10px", fontSize: 11, fontWeight: 700, color: C.dim, textTransform: "uppercase", letterSpacing: ".06em", marginBottom: 8 }}>{t({ru:"Кабинет учреждения",tg:"Кабинети муассиса"})}</p>
        <p style={{ padding: "0 10px", fontSize: 13.5, fontWeight: 700, color: C.text, marginBottom: 20, lineHeight: 1.3 }}>{t(inst0.name)}</p>

        <nav style={{ display: "flex", flexDirection: "column", gap: 2, flex: 1 }}>
          {TAB_META.map(({ k, labelKey, icon: Icon }) => (
            <button
              key={k}
              onClick={() => setTab(k)}
              style={{ display: "flex", alignItems: "center", gap: 10, padding: "10px 12px", borderRadius: 10, fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: tab === k ? C.teal : C.sub, background: tab === k ? `${C.teal}18` : "transparent", border: "none", cursor: "pointer", textAlign: "left" }}
            >
              <Icon size={16} /> {t(labelKey)}
            </button>
          ))}
        </nav>

        <Link href="/" style={{ display: "flex", alignItems: "center", gap: 10, padding: "10px 12px", borderRadius: 10, fontFamily: FH, fontWeight: 700, fontSize: 13, color: C.muted, textDecoration: "none" }}>
          <LogOut size={15} /> {t("nav.home")}
        </Link>
      </aside>

      {/* ── CONTENT ── */}
      <div style={{ flex: 1, padding: "28px 32px 80px", overflowX: "hidden" }}>
        <button onClick={() => (window.history.length > 1 ? router.back() : router.push("/"))} style={{ display: "inline-flex", alignItems: "center", gap: 7, fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: C.teal, marginBottom: 20, padding: "8px 14px", borderRadius: 9, border: `1px solid ${C.teal}40`, background: `${C.teal}10`, cursor: "pointer" }}>
          <ArrowLeft size={15} /> {t("common.back")}
        </button>
        {tab === "overview" && (
          <div>
            <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24, marginBottom: 22 }}>{t("dash.overview")}</h1>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(4,1fr)", gap: 14, marginBottom: 28 }}>
              {[
                { l: t({ru:"Рейтинг",tg:"Рейтинг"}), v: inst0.score, sub: `${inst0.rev} ${t("common.reviews")}` },
                { l: t({ru:"Учеников",tg:"Хонандагон"}), v: inst0.students, sub: `${t({ru:"с",tg:"аз"})} ${inst0.founded}` },
                { l: t("tab.news"), v: articles.length, sub: `${articles.filter((a) => a.status === "published").length} ${t({ru:"опубликовано",tg:"нашр шуд"})}` },
                { l: t("tab.staff"), v: staffList.length, sub: t({ru:"в профиле",tg:"дар профил"}) },
              ].map((s) => (
                <div key={s.l} style={{ borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: 18 }}>
                  <p style={{ fontFamily: FH, fontWeight: 900, fontSize: 26, color: C.teal }}>{s.v}</p>
                  <p style={{ fontSize: 13, color: C.text, fontWeight: 600, marginTop: 2 }}>{s.l}</p>
                  <p style={{ fontSize: 11.5, color: C.muted, marginTop: 2 }}>{s.sub}</p>
                </div>
              ))}
            </div>
            <div style={{ borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: 20, marginBottom: 20 }}>
              <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 15, marginBottom: 4 }}>{t({ru:"Просмотры и отклики за неделю",tg:"Дидан ва ҷавобҳо дар як ҳафта"})}</h3>
              <p style={{ fontSize: 12, color: C.muted, marginBottom: 14 }}>{t({ru:"Просмотры профиля и новые отклики/сообщения от посетителей",tg:"Дидани профил ва ҷавобу паёмҳои нави меҳмонон"})}</p>
              <div style={{ height: 200 }}>
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={WEEK_TREND} margin={{ top: 4, right: 8, left: -20, bottom: 0 }}>
                    <defs>
                      <linearGradient id="viewsFill" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="0%" stopColor={C.teal} stopOpacity={0.35} />
                        <stop offset="100%" stopColor={C.teal} stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <CartesianGrid stroke={C.border} vertical={false} />
                    <XAxis dataKey="day" tick={{ fill: C.muted, fontSize: 12 }} axisLine={{ stroke: C.border }} tickLine={false} />
                    <YAxis tick={{ fill: C.muted, fontSize: 12 }} axisLine={false} tickLine={false} width={28} />
                    <Tooltip contentStyle={{ background: C.s2, border: `1px solid ${C.border}`, borderRadius: 10, fontSize: 12.5 }} labelStyle={{ color: C.text, fontFamily: FH, fontWeight: 700 }} />
                    <Area type="monotone" dataKey="views" name={t({ru:"Просмотры",tg:"Дидан"})} stroke={C.teal} strokeWidth={2} fill="url(#viewsFill)" />
                    <Area type="monotone" dataKey="responses" name={t({ru:"Отклики",tg:"Ҷавобҳо"})} stroke={C.gold} strokeWidth={2} fill="none" />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>
            <div style={{ borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: 20 }}>
              <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 15, marginBottom: 14 }}>{t({ru:"Рейтинг по параметрам",tg:"Рейтинг аз рӯи меъёрҳо"})}</h3>
              {inst0.metrics.map((m) => (
                <div key={m.label} style={{ marginBottom: 9 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", fontSize: 12, color: C.sub, marginBottom: 4 }}>
                    <span>{t(m.label)}</span>
                    <span style={{ fontWeight: 700, color: C.text }}>{m.v.toFixed(1)}</span>
                  </div>
                  <div style={{ height: 5, borderRadius: 999, background: C.s3 }}>
                    <div style={{ height: "100%", borderRadius: 999, width: `${m.pct}%`, background: C.teal }} />
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {tab === "news" && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
              <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24 }}>{t("tab.news")}</h1>
              <button onClick={newArticle} style={{ display: "flex", alignItems: "center", gap: 7, padding: "10px 18px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                <Plus size={15} /> {t("common.add")}
              </button>
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {articles.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t("empty.news")}</p>}
              {articles.map((a) => (
                <div key={a.id} style={{ display: "flex", alignItems: "center", gap: 14, borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "12px 16px" }}>
                  <img src={a.coverUrl} alt="" style={{ width: 56, height: 56, borderRadius: 10, objectFit: "cover", flexShrink: 0 }} />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{t(a.title)}</p>
                    <p style={{ fontSize: 12, color: C.muted, marginTop: 2 }}>
                      {t(a.category)} · {a.date} · <span style={{ color: a.status === "published" ? C.ok : C.gold }}>{a.status === "published" ? t({ru:"Опубликовано",tg:"Нашр шудааст"}) : t({ru:"Черновик",tg:"Пешнавис"})}</span>
                    </p>
                  </div>
                  <button onClick={() => setNewsForm(a)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.sub }}><Pencil size={14} /></button>
                  <button onClick={() => deleteArticle(a.id)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.red }}><Trash2 size={14} /></button>
                </div>
              ))}
            </div>

            <Modal open={!!newsForm} onClose={() => setNewsForm(null)} maxWidth={520}>
              {newsForm && (
                <div style={{ padding: 26 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                    <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18 }}>{articles.some((a) => a.id === newsForm.id) ? t({ru:"Редактировать новость",tg:"Хабарро таҳрир кардан"}) : t({ru:"Новая новость",tg:"Хабари нав"})}</h3>
                    <button onClick={() => setNewsForm(null)} style={{ background: "none", border: "none", color: C.sub, cursor: "pointer" }}><X size={18} /></button>
                  </div>
                  <FormField label={`${t({ru:"Заголовок",tg:"Сарлавҳа"})} (${locale.toUpperCase()})`}><input value={newsForm.title[locale]} onChange={(e) => setNewsForm({ ...newsForm, title: { ...newsForm.title, [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={`${t({ru:"Категория",tg:"Категория"})} (${locale.toUpperCase()})`}><input value={newsForm.category[locale]} onChange={(e) => setNewsForm({ ...newsForm, category: { ...newsForm.category, [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={`${t({ru:"Текст",tg:"Матн"})} (${locale.toUpperCase()})`}><textarea value={newsForm.content[locale]} onChange={(e) => setNewsForm({ ...newsForm, content: { ...newsForm.content, [locale]: e.target.value } })} style={{ ...inputStyle, height: 100, resize: "none" as const }} /></FormField>
                  <div style={{ display: "flex", gap: 10, alignItems: "center", marginTop: 6, marginBottom: 20 }}>
                    <label style={{ fontSize: 13, color: C.sub, fontFamily: FB }}>{t({ru:"Статус:",tg:"Ҳолат:"})}</label>
                    <select value={newsForm.status} onChange={(e) => setNewsForm({ ...newsForm, status: e.target.value as NewsArticle["status"] })} style={{ ...inputStyle, width: "auto", padding: "8px 12px" }}>
                      <option value="draft">{t({ru:"Черновик",tg:"Пешнавис"})}</option>
                      <option value="published">{t({ru:"Опубликовано",tg:"Нашр шудааст"})}</option>
                    </select>
                  </div>
                  <button onClick={saveArticle} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer" }}>{t("common.save")}</button>
                </div>
              )}
            </Modal>
          </div>
        )}

        {tab === "staff" && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
              <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24 }}>{t("tab.staff")}</h1>
              <button onClick={newStaff} style={{ display: "flex", alignItems: "center", gap: 7, padding: "10px 18px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                <Plus size={15} /> {t("common.add")}
              </button>
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {staffList.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t("empty.staff")}</p>}
              {staffList.map((p) => (
                <div key={p.id} style={{ display: "flex", alignItems: "center", gap: 14, borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "12px 16px" }}>
                  <img src={p.photo} alt="" style={{ width: 48, height: 48, borderRadius: "50%", objectFit: "cover", flexShrink: 0 }} />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text }}>{t(p.name)}</p>
                    <p style={{ fontSize: 12, color: C.muted, marginTop: 2 }}>{t(p.roleLabel)}{p.subject ? ` · ${t(p.subject)}` : ""}</p>
                  </div>
                  <button onClick={() => setStaffForm(p)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.sub }}><Pencil size={14} /></button>
                  <button onClick={() => deleteStaff(p.id)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.red }}><Trash2 size={14} /></button>
                </div>
              ))}
            </div>

            <Modal open={!!staffForm} onClose={() => setStaffForm(null)} maxWidth={520}>
              {staffForm && (
                <div style={{ padding: 26 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                    <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18 }}>{staffList.some((p) => p.id === staffForm.id) ? t({ru:"Редактировать сотрудника",tg:"Корманд таҳрир кардан"}) : t({ru:"Новый сотрудник",tg:"Корманди нав"})}</h3>
                    <button onClick={() => setStaffForm(null)} style={{ background: "none", border: "none", color: C.sub, cursor: "pointer" }}><X size={18} /></button>
                  </div>
                  <FormField label={`${t({ru:"Имя",tg:"Ном"})} (${locale.toUpperCase()})`}><input value={staffForm.name[locale]} onChange={(e) => setStaffForm({ ...staffForm, name: { ...staffForm.name, [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={`${t({ru:"Должность",tg:"Вазифа"})} (${locale.toUpperCase()})`}><input value={staffForm.roleLabel[locale]} onChange={(e) => setStaffForm({ ...staffForm, roleLabel: { ...staffForm.roleLabel, [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={t({ru:"Предмет (опционально)",tg:"Фан (ихтиёрӣ)"})}><input value={staffForm.subject?.[locale] ?? ""} onChange={(e) => setStaffForm({ ...staffForm, subject: { ru: staffForm.subject?.ru ?? "", tg: staffForm.subject?.tg ?? "", [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={t({ru:"Опыт",tg:"Таҷриба"})}><input value={staffForm.exp} onChange={(e) => setStaffForm({ ...staffForm, exp: e.target.value })} style={inputStyle} /></FormField>
                  <FormField label={t({ru:"Фото",tg:"Акс"})}>
                    <select value={PHOTO_KEYS.find((k) => PHOTOS[k] === staffForm.photo) ?? PHOTO_KEYS[0]} onChange={(e) => setStaffForm({ ...staffForm, photo: PHOTOS[e.target.value as keyof typeof PHOTOS] })} style={inputStyle}>
                      {PHOTO_KEYS.filter((k) => /^p[wm]\d/.test(k)).map((k) => <option key={k} value={k}>{k}</option>)}
                    </select>
                  </FormField>
                  <FormField label={`${t({ru:"О сотруднике",tg:"Дар бораи корманд"})} (${locale.toUpperCase()})`}><textarea value={staffForm.bio[locale]} onChange={(e) => setStaffForm({ ...staffForm, bio: { ...staffForm.bio, [locale]: e.target.value } })} style={{ ...inputStyle, height: 90, resize: "none" as const }} /></FormField>
                  <FormField label={t({ru:"Email",tg:"Почта"})}><input value={staffForm.email ?? ""} onChange={(e) => setStaffForm({ ...staffForm, email: e.target.value })} style={inputStyle} /></FormField>
                  <button onClick={saveStaff} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer", marginTop: 6 }}>{t("common.save")}</button>
                </div>
              )}
            </Modal>
          </div>
        )}

        {tab === "vacancies" && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
              <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24 }}>{t("tab.vacancies")}</h1>
              <button onClick={newVacancy} style={{ display: "flex", alignItems: "center", gap: 7, padding: "10px 18px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                <Plus size={15} /> {t("common.add")}
              </button>
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {vacancyList.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t("empty.vacancies")}</p>}
              {vacancyList.map((v) => {
                const candidates = VACANCY_CANDIDATES.filter((c) => c.vacancyId === v.id);
                return (
                <div key={v.id} style={{ borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "12px 16px" }}>
                  <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
                    <div style={{ width: 44, height: 44, borderRadius: 10, background: `${C.teal}18`, display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
                      <Briefcase size={19} style={{ color: C.teal }} />
                    </div>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{t(v.title)}</p>
                      <p style={{ fontSize: 12, color: C.muted, marginTop: 2 }}>
                        {v.date} · <span style={{ color: v.status === "published" ? C.ok : C.gold }}>{v.status === "published" ? t({ru:"Опубликовано",tg:"Нашр шудааст"}) : t({ru:"Черновик",tg:"Пешнавис"})}</span>
                      </p>
                    </div>
                    <button onClick={() => setVacancyForm(v)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.sub }}><Pencil size={14} /></button>
                    <button onClick={() => deleteVacancy(v.id)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.red }}><Trash2 size={14} /></button>
                  </div>
                  <div style={{ marginTop: 12, paddingTop: 12, borderTop: `1px solid ${C.border}`, display: "flex", alignItems: "center", gap: 10, flexWrap: "wrap" }}>
                    <span style={{ display: "flex", alignItems: "center", gap: 6, fontSize: 12.5, fontFamily: FH, fontWeight: 700, color: candidates.length ? C.teal : C.muted }}>
                      <Users size={13} /> {candidates.length} {t({ru:"откликнулись",tg:"ҷавоб доданд"})}
                    </span>
                    {candidates.length > 0 && (
                      <div style={{ display: "flex", alignItems: "center", flexWrap: "wrap", gap: "6px 10px" }}>
                        {candidates.map((c) => {
                          const prof = c.applicantId === applicant.id ? applicant : SEED_APPLICANTS.find((a) => a.id === c.applicantId);
                          if (!prof) return null;
                          return (
                            <Link key={c.id} href={`/applicants/${c.applicantId}`} style={{ display: "flex", alignItems: "center", gap: 6, padding: "4px 10px 4px 4px", borderRadius: 999, background: C.s2, border: `1px solid ${C.border}`, textDecoration: "none" }}>
                              <img src={prof.photo} alt="" style={{ width: 22, height: 22, borderRadius: "50%", objectFit: "cover" }} />
                              <span style={{ fontSize: 12, color: C.text }}>{t(prof.name)}</span>
                              <span style={{ fontSize: 11, color: C.dim }}>· {c.appliedAt}</span>
                            </Link>
                          );
                        })}
                      </div>
                    )}
                  </div>
                </div>
                );
              })}
            </div>

            <Modal open={!!vacancyForm} onClose={() => setVacancyForm(null)} maxWidth={520}>
              {vacancyForm && (
                <div style={{ padding: 26 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                    <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18 }}>{vacancyList.some((v) => v.id === vacancyForm.id) ? t({ru:"Редактировать вакансию",tg:"Ҷои холиро таҳрир кардан"}) : t({ru:"Новая вакансия",tg:"Ҷои холии нав"})}</h3>
                    <button onClick={() => setVacancyForm(null)} style={{ background: "none", border: "none", color: C.sub, cursor: "pointer" }}><X size={18} /></button>
                  </div>
                  <FormField label={`${t({ru:"Должность",tg:"Вазифа"})} (${locale.toUpperCase()})`}><input value={vacancyForm.title[locale]} onChange={(e) => setVacancyForm({ ...vacancyForm, title: { ...vacancyForm.title, [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={`${t({ru:"Описание",tg:"Тавсиф"})} (${locale.toUpperCase()})`}><textarea value={vacancyForm.description[locale]} onChange={(e) => setVacancyForm({ ...vacancyForm, description: { ...vacancyForm.description, [locale]: e.target.value } })} style={{ ...inputStyle, height: 90, resize: "none" as const }} /></FormField>
                  <FormField label={t({ru:"Требования (по одному на строку)",tg:"Талабот (ҳар як дар як сатр)"})}>
                    <textarea
                      value={vacancyForm.requirements.map((r) => r[locale]).join("\n")}
                      onChange={(e) => setVacancyForm({ ...vacancyForm, requirements: e.target.value.split("\n").filter((l) => l.trim()).map((l) => bi(l, l)) })}
                      style={{ ...inputStyle, height: 80, resize: "none" as const }}
                    />
                  </FormField>
                  <div style={{ display: "flex", gap: 10 }}>
                    <div style={{ flex: 1 }}>
                      <FormField label={t({ru:"Зарплата от",tg:"Маош аз"})}><input type="number" value={vacancyForm.salaryFrom ?? ""} onChange={(e) => setVacancyForm({ ...vacancyForm, salaryFrom: e.target.value ? Number(e.target.value) : undefined })} style={inputStyle} /></FormField>
                    </div>
                    <div style={{ flex: 1 }}>
                      <FormField label={t({ru:"Зарплата до",tg:"Маош то"})}><input type="number" value={vacancyForm.salaryTo ?? ""} onChange={(e) => setVacancyForm({ ...vacancyForm, salaryTo: e.target.value ? Number(e.target.value) : undefined })} style={inputStyle} /></FormField>
                    </div>
                  </div>
                  <FormField label={`${t({ru:"Занятость",tg:"Шуғл"})} (${locale.toUpperCase()})`}><input value={vacancyForm.employment[locale]} onChange={(e) => setVacancyForm({ ...vacancyForm, employment: { ...vacancyForm.employment, [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <div style={{ display: "flex", gap: 10, alignItems: "center", marginTop: 6, marginBottom: 20 }}>
                    <label style={{ fontSize: 13, color: C.sub, fontFamily: FB }}>{t({ru:"Статус:",tg:"Ҳолат:"})}</label>
                    <select value={vacancyForm.status} onChange={(e) => setVacancyForm({ ...vacancyForm, status: e.target.value as Vacancy["status"] })} style={{ ...inputStyle, width: "auto", padding: "8px 12px" }}>
                      <option value="draft">{t({ru:"Черновик",tg:"Пешнавис"})}</option>
                      <option value="published">{t({ru:"Опубликовано",tg:"Нашр шудааст"})}</option>
                    </select>
                  </div>
                  <button onClick={saveVacancy} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer" }}>{t("common.save")}</button>
                </div>
              )}
            </Modal>
          </div>
        )}

        {tab === "achievements" && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
              <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24 }}>{t("tab.achievements")}</h1>
              <button onClick={newAchievement} style={{ display: "flex", alignItems: "center", gap: 7, padding: "10px 18px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                <Plus size={15} /> {t("common.add")}
              </button>
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {achievementsList.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t("empty.achievements")}</p>}
              {achievementsList.map((a) => (
                <div key={a.id} style={{ display: "flex", alignItems: "center", gap: 14, borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "12px 16px" }}>
                  <div style={{ width: 44, height: 44, borderRadius: 12, background: `${C.gold}18`, display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
                    <Award size={19} style={{ color: C.gold }} />
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{t(a.title)}</p>
                    <p style={{ fontSize: 12, color: C.muted, marginTop: 2 }}>{a.year} · {a.category}</p>
                  </div>
                  <button onClick={() => setAchievementForm(a)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.sub }}><Pencil size={14} /></button>
                  <button onClick={() => deleteAchievement(a.id)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.red }}><Trash2 size={14} /></button>
                </div>
              ))}
            </div>

            <Modal open={!!achievementForm} onClose={() => setAchievementForm(null)} maxWidth={480}>
              {achievementForm && (
                <div style={{ padding: 26 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                    <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18 }}>{achievementsList.some((a) => a.id === achievementForm.id) ? t({ru:"Редактировать достижение",tg:"Дастовардро таҳрир кардан"}) : t({ru:"Новое достижение",tg:"Дастоварди нав"})}</h3>
                    <button onClick={() => setAchievementForm(null)} style={{ background: "none", border: "none", color: C.sub, cursor: "pointer" }}><X size={18} /></button>
                  </div>
                  <FormField label={`${t({ru:"Название",tg:"Ном"})} (${locale.toUpperCase()})`}><input value={achievementForm.title[locale]} onChange={(e) => setAchievementForm({ ...achievementForm, title: { ...achievementForm.title, [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={`${t({ru:"Описание",tg:"Тавсиф"})} (${locale.toUpperCase()})`}><textarea value={achievementForm.desc[locale]} onChange={(e) => setAchievementForm({ ...achievementForm, desc: { ...achievementForm.desc, [locale]: e.target.value } })} style={{ ...inputStyle, height: 80, resize: "none" as const }} /></FormField>
                  <div style={{ display: "flex", gap: 10 }}>
                    <div style={{ flex: 1 }}>
                      <FormField label={t({ru:"Год",tg:"Сол"})}><input type="number" value={achievementForm.year} onChange={(e) => setAchievementForm({ ...achievementForm, year: Number(e.target.value) })} style={inputStyle} /></FormField>
                    </div>
                    <div style={{ flex: 1 }}>
                      <FormField label={t({ru:"Уровень",tg:"Дараҷа"})}>
                        <select value={achievementForm.category} onChange={(e) => setAchievementForm({ ...achievementForm, category: e.target.value as Achievement["category"] })} style={inputStyle}>
                          <option value="gold">{t({ru:"Золото",tg:"Тилло"})}</option>
                          <option value="silver">{t({ru:"Серебро",tg:"Нуқра"})}</option>
                          <option value="bronze">{t({ru:"Бронза",tg:"Биринҷӣ"})}</option>
                          <option value="special">{t({ru:"Особое",tg:"Махсус"})}</option>
                        </select>
                      </FormField>
                    </div>
                  </div>
                  <FormField label={t({ru:"Тип",tg:"Навъ"})}>
                    <select value={achievementForm.type} onChange={(e) => setAchievementForm({ ...achievementForm, type: e.target.value as Achievement["type"] })} style={inputStyle}>
                      <option value="institution">{t({ru:"Учреждение",tg:"Муассиса"})}</option>
                      <option value="student">{t({ru:"Ученик",tg:"Хонанда"})}</option>
                      <option value="teacher">{t({ru:"Педагог",tg:"Омӯзгор"})}</option>
                    </select>
                  </FormField>
                  <button onClick={saveAchievement} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer", marginTop: 6 }}>{t("common.save")}</button>
                </div>
              )}
            </Modal>
          </div>
        )}

        {tab === "alumni" && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
              <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24 }}>{t("tab.alumni")}</h1>
              <button onClick={newAlumnus} style={{ display: "flex", alignItems: "center", gap: 7, padding: "10px 18px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                <Plus size={15} /> {t("common.add")}
              </button>
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {alumniList.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t({ru:"Выпускники не добавлены",tg:"Хатмкунандагон илова нашудаанд"})}</p>}
              {alumniList.map((a) => (
                <div key={a.id} style={{ display: "flex", alignItems: "center", gap: 14, borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "12px 16px" }}>
                  <img src={a.photo} alt="" style={{ width: 44, height: 44, borderRadius: "50%", objectFit: "cover", flexShrink: 0 }} />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text }}>{t(a.name)}</p>
                    <p style={{ fontSize: 12, color: C.muted, marginTop: 2 }}>{a.gradYear} · {t(a.now)}</p>
                  </div>
                  <button onClick={() => setAlumnusForm(a)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.sub }}><Pencil size={14} /></button>
                  <button onClick={() => deleteAlumnus(a.id)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.red }}><Trash2 size={14} /></button>
                </div>
              ))}
            </div>

            <Modal open={!!alumnusForm} onClose={() => setAlumnusForm(null)} maxWidth={480}>
              {alumnusForm && (
                <div style={{ padding: 26 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                    <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18 }}>{alumniList.some((a) => a.id === alumnusForm.id) ? t({ru:"Редактировать выпускника",tg:"Хатмкунандаро таҳрир кардан"}) : t({ru:"Новый выпускник",tg:"Хатмкунандаи нав"})}</h3>
                    <button onClick={() => setAlumnusForm(null)} style={{ background: "none", border: "none", color: C.sub, cursor: "pointer" }}><X size={18} /></button>
                  </div>
                  <FormField label={`${t({ru:"Имя",tg:"Ном"})} (${locale.toUpperCase()})`}><input value={alumnusForm.name[locale]} onChange={(e) => setAlumnusForm({ ...alumnusForm, name: { ...alumnusForm.name, [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={t({ru:"Фото",tg:"Акс"})}>
                    <select value={PHOTO_KEYS.find((k) => PHOTOS[k] === alumnusForm.photo) ?? PHOTO_KEYS[0]} onChange={(e) => setAlumnusForm({ ...alumnusForm, photo: PHOTOS[e.target.value as keyof typeof PHOTOS] })} style={inputStyle}>
                      {PHOTO_KEYS.filter((k) => /^p[wm]\d/.test(k)).map((k) => <option key={k} value={k}>{k}</option>)}
                    </select>
                  </FormField>
                  <FormField label={t({ru:"Год выпуска",tg:"Соли хатм"})}><input type="number" value={alumnusForm.gradYear} onChange={(e) => setAlumnusForm({ ...alumnusForm, gradYear: Number(e.target.value) })} style={inputStyle} /></FormField>
                  <FormField label={`${t({ru:"Где сейчас",tg:"Ҳозир дар куҷо"})} (${locale.toUpperCase()})`}><input value={alumnusForm.now[locale]} onChange={(e) => setAlumnusForm({ ...alumnusForm, now: { ...alumnusForm.now, [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <button onClick={saveAlumnus} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer", marginTop: 6 }}>{t("common.save")}</button>
                </div>
              )}
            </Modal>
          </div>
        )}

        {tab === "gallery" && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
              <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24 }}>{t("tab.gallery")}</h1>
              <button onClick={newGalleryItem} style={{ display: "flex", alignItems: "center", gap: 7, padding: "10px 18px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                <Plus size={15} /> {t("common.add")}
              </button>
            </div>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill,minmax(160px,1fr))", gap: 12 }}>
              {galleryList.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t({ru:"Галерея пуста",tg:"Галерея холист"})}</p>}
              {galleryList.map((g) => (
                <div key={g.url} style={{ borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, overflow: "hidden" }}>
                  <img src={g.url} alt="" style={{ width: "100%", height: 100, objectFit: "cover" }} />
                  <div style={{ padding: "8px 10px", display: "flex", alignItems: "center", gap: 6 }}>
                    <p style={{ flex: 1, minWidth: 0, fontSize: 12, color: C.text, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{t(g.label)}</p>
                    <button onClick={() => setGalleryForm(g)} style={{ background: C.s3, border: "none", borderRadius: 7, padding: 6, cursor: "pointer", color: C.sub }}><Pencil size={11} /></button>
                    <button onClick={() => deleteGalleryItem(g.url)} style={{ background: C.s3, border: "none", borderRadius: 7, padding: 6, cursor: "pointer", color: C.red }}><Trash2 size={11} /></button>
                  </div>
                </div>
              ))}
            </div>

            <Modal open={!!galleryForm} onClose={() => setGalleryForm(null)} maxWidth={440}>
              {galleryForm && (
                <div style={{ padding: 26 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                    <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18 }}>{t({ru:"Фото галереи",tg:"Акси галерея"})}</h3>
                    <button onClick={() => setGalleryForm(null)} style={{ background: "none", border: "none", color: C.sub, cursor: "pointer" }}><X size={18} /></button>
                  </div>
                  <FormField label={t({ru:"Фото",tg:"Акс"})}>
                    <select value={PHOTO_KEYS.find((k) => PHOTOS[k] === galleryForm.url) ?? PHOTO_KEYS[0]} onChange={(e) => setGalleryForm({ ...galleryForm, url: PHOTOS[e.target.value as keyof typeof PHOTOS] })} style={inputStyle}>
                      {PHOTO_KEYS.filter((k) => !/^p[wm]\d/.test(k)).map((k) => <option key={k} value={k}>{k}</option>)}
                    </select>
                  </FormField>
                  <FormField label={`${t({ru:"Подпись",tg:"Изоҳ"})} (${locale.toUpperCase()})`}><input value={galleryForm.label[locale]} onChange={(e) => setGalleryForm({ ...galleryForm, label: { ...galleryForm.label, [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <button onClick={saveGalleryItem} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer", marginTop: 6 }}>{t("common.save")}</button>
                </div>
              )}
            </Modal>
          </div>
        )}

        {tab === "reviews" && (
          <div>
            <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24, marginBottom: 20 }}>{t("tab.reviews")}</h1>
            <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
              {reviewsList.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t("empty.reviews")}</p>}
              {reviewsList.map((r) => (
                <div key={r.id} style={{ borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: "16px 18px" }}>
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 8 }}>
                    <div>
                      <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text }}>{t(r.name)}</p>
                      <p style={{ fontSize: 12, color: C.sub }}>{r.date}</p>
                    </div>
                    <span style={{ color: C.gold, fontFamily: FH, fontWeight: 700 }}>{r.score} ★</span>
                  </div>
                  <p style={{ fontSize: 13.5, color: C.sub, lineHeight: 1.6, marginBottom: 10 }}>{t(r.text)}</p>
                  {r.reply && (
                    <div style={{ padding: "10px 14px", borderRadius: 10, background: `${C.teal}12`, marginBottom: 10 }}>
                      <p style={{ fontSize: 12, fontWeight: 700, color: C.teal, fontFamily: FH, marginBottom: 4 }}>{t({ru:"Ваш ответ:",tg:"Ҷавоби шумо:"})}</p>
                      <p style={{ fontSize: 13, color: C.sub }}>{t(r.reply)}</p>
                    </div>
                  )}
                  {replyDraftId === r.id ? (
                    <div style={{ display: "flex", gap: 8 }}>
                      <input value={replyText} onChange={(e) => setReplyText(e.target.value)} placeholder={t({ru:"Ваш ответ…",tg:"Ҷавоби шумо…"})} style={{ ...inputStyle, flex: 1 }} />
                      <button onClick={() => submitReply(r.id)} style={{ padding: "0 16px", borderRadius: 10, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 700, fontSize: 13, border: "none", cursor: "pointer" }}>{t("common.send")}</button>
                    </div>
                  ) : (
                    <button onClick={() => { setReplyDraftId(r.id); setReplyText(r.reply?.[locale] ?? ""); }} style={{ fontSize: 12.5, fontWeight: 700, color: C.teal, fontFamily: FH, background: "none", border: "none", cursor: "pointer" }}>
                      {r.reply ? t({ru:"Изменить ответ",tg:"Ҷавобро тағйир додан"}) : t("common.reply")}
                    </button>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {tab === "settings" && (
          <div>
            <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24, marginBottom: 20 }}>{t("dash.settings")}</h1>
            <div style={{ display: "flex", gap: 6, marginBottom: 20 }}>
              {([["info", t({ru:"Информация",tg:"Маълумот"})], ["contacts", t({ru:"Контакты и соцсети",tg:"Тамос ва шабакаҳо"})], ["hours", t({ru:"Часы работы",tg:"Тартиби корӣ"})]] as [SettingsTab, string][]).map(([k, l]) => (
                <button key={k} onClick={() => setSettingsTab(k)} style={{ padding: "8px 16px", borderRadius: 9, fontFamily: FH, fontWeight: 700, fontSize: 13, border: `1px solid ${settingsTab === k ? C.teal : C.border}`, background: settingsTab === k ? `${C.teal}18` : "transparent", color: settingsTab === k ? C.teal : C.sub, cursor: "pointer" }}>
                  {l}
                </button>
              ))}
            </div>

            <form onSubmit={saveSettings} style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 26, maxWidth: 520 }}>
              {settingsTab === "info" && (
                <>
                  <FormField label={`${t({ru:"Название",tg:"Ном"})} (${locale.toUpperCase()})`}><input value={info.name[locale]} onChange={(e) => setInfo({ ...info, name: { ...info.name, [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={t({ru:"Тип",tg:"Навъ"})}><input value={t(CATEGORY_META[inst0.tk].label)} disabled style={{ ...inputStyle, opacity: 0.6, cursor: "not-allowed" }} /></FormField>
                  <FormField label={t({ru:"Район",tg:"Ноҳия"})}><input value={info.area} onChange={(e) => setInfo({ ...info, area: e.target.value })} style={inputStyle} /></FormField>
                  <FormField label={`${t({ru:"Описание",tg:"Тавсиф"})} (${locale.toUpperCase()})`}><textarea value={info.description[locale]} onChange={(e) => setInfo({ ...info, description: { ...info.description, [locale]: e.target.value } })} style={{ ...inputStyle, height: 110, resize: "none" as const }} /></FormField>
                </>
              )}
              {settingsTab === "contacts" && (
                <>
                  <FormField label={t({ru:"Телефон",tg:"Телефон"})}><input value={contacts.phone} onChange={(e) => setContacts({ ...contacts, phone: e.target.value })} style={inputStyle} /></FormField>
                  <FormField label={t({ru:"Эл. почта",tg:"Почтаи электронӣ"})}><input value={contacts.email} onChange={(e) => setContacts({ ...contacts, email: e.target.value })} style={inputStyle} /></FormField>
                  <FormField label={t("geo.region")}>
                    <select value={region} onChange={(e) => setRegionField(e.target.value as Region)} style={inputStyle}>
                      {REGION_ORDER.map((r) => <option key={r} value={r}>{t(REGION_LABEL[r])}</option>)}
                    </select>
                  </FormField>
                  <FormField label={`${t("geo.city")} (${locale.toUpperCase()})`}><input value={contacts.city[locale]} onChange={(e) => setContacts({ ...contacts, city: { ...contacts.city, [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={`${t("geo.street")} (${locale.toUpperCase()})`}><input value={contacts.street[locale]} onChange={(e) => setContacts({ ...contacts, street: { ...contacts.street, [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={t({ru:"Официальный сайт",tg:"Сомонаи расмӣ"})}><input value={contacts.website} onChange={(e) => setContacts({ ...contacts, website: e.target.value })} style={inputStyle} /></FormField>
                  <div style={{ display: "flex", gap: 10, marginTop: 4 }}>
                    <div style={{ flex: 1 }}>
                      <FormField label="Instagram">
                        <input value={contacts.instagram} onChange={(e) => setContacts({ ...contacts, instagram: e.target.value })} placeholder="username" style={inputStyle} />
                      </FormField>
                    </div>
                    <div style={{ flex: 1 }}>
                      <FormField label="Telegram">
                        <input value={contacts.telegram} onChange={(e) => setContacts({ ...contacts, telegram: e.target.value })} placeholder="username" style={inputStyle} />
                      </FormField>
                    </div>
                    <div style={{ flex: 1 }}>
                      <FormField label="Facebook">
                        <input value={contacts.facebook} onChange={(e) => setContacts({ ...contacts, facebook: e.target.value })} placeholder="username" style={inputStyle} />
                      </FormField>
                    </div>
                  </div>
                </>
              )}
              {settingsTab === "hours" && (
                <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
                  {hours.map((h, i) => (
                    <div key={h.labelKey} style={{ display: "flex", gap: 10, alignItems: "center" }}>
                      <span style={{ width: 110, fontSize: 13, color: C.sub, fontFamily: FB, flexShrink: 0 }}>{t(h.labelKey)}</span>
                      <input value={h.time} onChange={(e) => setHours((prev) => prev.map((x, xi) => (xi === i ? { ...x, time: e.target.value } : x)))} style={inputStyle} />
                    </div>
                  ))}
                </div>
              )}
              <button type="submit" style={{ marginTop: 20, padding: "12px 24px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer" }}>
                {t("common.save")}
              </button>
            </form>
          </div>
        )}
      </div>

      <Toast message={toast} onDone={() => setToast(null)} />
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  width: "100%", padding: "10px 13px", borderRadius: 10, border: `1px solid ${C.border}`,
  background: C.s2, color: C.text, fontFamily: FB, fontSize: 13.5, outline: "none", boxSizing: "border-box",
};

function FormField({ label, children }: { label: React.ReactNode; children: React.ReactNode }) {
  return (
    <div style={{ marginBottom: 14 }}>
      <label style={{ display: "block", fontSize: 11.5, fontWeight: 700, color: C.dim, textTransform: "uppercase", letterSpacing: ".05em", fontFamily: FH, marginBottom: 6 }}>{label}</label>
      {children}
    </div>
  );
}
