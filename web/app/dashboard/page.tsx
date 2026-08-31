"use client";

import { useState } from "react";
import Link from "next/link";
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from "recharts";
import { useRouter } from "next/navigation";
import {
  LayoutDashboard, Newspaper, Users, Star, Settings, Plus, Pencil, Trash2, X, LogOut, Briefcase, Award, UserSquare2, Camera, ArrowLeft, Building2, type LucideIcon,
} from "lucide-react";
import { C, FH, FB } from "@/lib/data";
import { Modal } from "@/components/ui/Modal";
import { Button } from "@/shared/components/button";
import { Input } from "@/shared/components/input";
import { Textarea } from "@/shared/components/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/components/select";
import { toast } from "sonner";
import { useAppState } from "@/lib/app-state";
import { useT } from "@/lib/i18n";
import {
  useGetMineQuery, useGetMineFullQuery, useUpdateInstitutionMutation,
  useListNewsQuery, useCreateNewsMutation, useUpdateNewsMutation, useDeleteNewsMutation,
  useListMyReviewsQuery, useReplyToReviewMutation,
  useCreateStaffMutation, useUpdateStaffMutation, useDeleteStaffMutation,
  useCreateAchievementMutation, useDeleteAchievementMutation,
  useCreateAlumnusMutation, useDeleteAlumnusMutation,
  useCreateGalleryItemMutation, useDeleteGalleryItemMutation,
  useListMyVacanciesQuery, useCreateVacancyMutation, useUpdateVacancyMutation, useDeleteVacancyMutation,
  type StaffMemberDTO, type AchievementDTO, type AlumnusDTO, type GalleryItemDTO, type NewsArticleDTO, type VacancyDTO, type VacancyRequest,
} from "./api/dashboardApi";

type Tab = "overview" | "news" | "staff" | "vacancies" | "achievements" | "alumni" | "gallery" | "reviews" | "settings";

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

const WEEK_TREND = [
  { day: "Пн", views: 42, responses: 3 },
  { day: "Вт", views: 58, responses: 5 },
  { day: "Ср", views: 51, responses: 4 },
  { day: "Чт", views: 67, responses: 7 },
  { day: "Пт", views: 74, responses: 6 },
  { day: "Сб", views: 39, responses: 2 },
  { day: "Вс", views: 33, responses: 3 },
];

function NoInstitutionScreen({ t }: { t: ReturnType<typeof useT> }) {
  return (
    <div style={{ display: "flex", flex: 1, alignItems: "center", justifyContent: "center", minHeight: "70vh", padding: 32 }}>
      <div style={{ maxWidth: 440, textAlign: "center" }}>
        <Building2 size={28} style={{ color: C.teal, margin: "0 auto 14px" }} />
        <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 22, color: C.text, marginBottom: 10 }}>
          {t({ ru: "У вас пока нет учреждения", tg: "Шумо ҳанӯз муассиса надоред" })}
        </h1>
        <p style={{ fontSize: 14, color: C.sub, lineHeight: 1.6, marginBottom: 20 }}>
          {t({ ru: "Зарегистрируйте учреждение, чтобы получить доступ к кабинету.", tg: "Барои дастрасӣ ба кабинет муассисаро сабт кунед." })}
        </p>
        <Link href="/register-institution" style={{ display: "inline-flex", padding: "12px 26px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, textDecoration: "none" }}>
          {t({ ru: "Зарегистрировать учреждение", tg: "Сабти муассиса" })}
        </Link>
      </div>
    </div>
  );
}

export default function DashboardPage() {
  const { locale } = useAppState();
  const t = useT();
  const router = useRouter();
  const [tab, setTab] = useState<Tab>("overview");

  // Институт текущего владельца берётся из backend/internal/catalog (GET /api/v1/institutions/mine
  // → id, затем GET /api/v1/institutions/{id}/mine → полная карточка со staff/achievements/gallery/alumni).
  const { data: mine, isLoading: mineLoading } = useGetMineQuery();
  const instId = mine?.items[0]?.id;
  const { data: inst, isLoading: instLoading } = useGetMineFullQuery(instId ?? "", { skip: !instId });

  if (mineLoading || (instId && instLoading)) {
    return <div style={{ padding: 60, textAlign: "center", color: C.muted }}>{t({ ru: "Загрузка…", tg: "Боркунӣ…" })}</div>;
  }
  if (!instId || !inst) {
    return <NoInstitutionScreen t={t} />;
  }

  return <DashboardInner inst={inst} instId={instId} locale={locale} t={t} router={router} tab={tab} setTab={setTab} />;
}

function DashboardInner({
  inst, instId, locale, t, router, tab, setTab,
}: {
  inst: import("./api/dashboardApi").FullInstitutionDTO;
  instId: string;
  locale: "ru" | "tg";
  t: ReturnType<typeof useT>;
  router: ReturnType<typeof useRouter>;
  tab: Tab;
  setTab: (t: Tab) => void;
}) {
  // ── news (реальный backend) ──
  const { data: newsData } = useListNewsQuery(instId);
  const articles = newsData?.items ?? [];
  const [newsForm, setNewsForm] = useState<Partial<NewsArticleDTO> | null>(null);
  const [createNews] = useCreateNewsMutation();
  const [updateNews] = useUpdateNewsMutation();
  const [deleteNewsMut] = useDeleteNewsMutation();

  function newArticle() {
    setNewsForm({ title: { ru: "", tg: "" }, content: { ru: "", tg: "" }, status: "draft" });
  }
  async function saveArticle() {
    if (!newsForm || !newsForm.title?.[locale]?.trim()) return;
    const body = {
      title: newsForm.title ?? { ru: "", tg: "" },
      content: newsForm.content ?? { ru: "", tg: "" },
      status: (newsForm.status ?? "draft") as "draft" | "published",
    };
    try {
      if (newsForm.id) {
        await updateNews({ institutionId: instId, newsId: newsForm.id, body }).unwrap();
      } else {
        await createNews({ institutionId: instId, body }).unwrap();
      }
      setNewsForm(null);
      toast.success(t({ ru: "Новость сохранена", tg: "Хабар нигоҳ дошта шуд" }));
    } catch {
      toast.error(t({ ru: "Не удалось сохранить", tg: "Нигоҳ дошта нашуд" }));
    }
  }
  async function deleteArticle(id: string) {
    try {
      await deleteNewsMut({ institutionId: instId, newsId: id }).unwrap();
    } catch {
      toast.error(t({ ru: "Не удалось удалить", tg: "Нест карда нашуд" }));
    }
  }

  // ── staff (реальный backend) ──
  const staffList = inst.staff;
  const [staffForm, setStaffForm] = useState<Partial<StaffMemberDTO> | null>(null);
  const [createStaff] = useCreateStaffMutation();
  const [updateStaff] = useUpdateStaffMutation();
  const [deleteStaffMut] = useDeleteStaffMutation();

  function newStaff() {
    setStaffForm({ name: { ru: "", tg: "" }, role_type: "teacher", role_label: { ru: "", tg: "" } });
  }
  async function saveStaff() {
    if (!staffForm || !staffForm.name?.[locale]?.trim()) return;
    const body = {
      name: staffForm.name ?? { ru: "", tg: "" },
      role_type: staffForm.role_type ?? "teacher",
      role_label: staffForm.role_label ?? { ru: "", tg: "" },
      subject: staffForm.subject,
      exp: staffForm.exp,
      bio: staffForm.bio,
      email: staffForm.email,
      phone: staffForm.phone,
      photo_url: staffForm.photo_url,
    };
    try {
      if (staffForm.id) {
        await updateStaff({ institutionId: instId, staffId: staffForm.id, body }).unwrap();
      } else {
        await createStaff({ institutionId: instId, body }).unwrap();
      }
      setStaffForm(null);
      toast.success(t({ ru: "Сотрудник сохранён", tg: "Корманд нигоҳ дошта шуд" }));
    } catch {
      toast.error(t({ ru: "Не удалось сохранить", tg: "Нигоҳ дошта нашуд" }));
    }
  }
  async function deleteStaff(id: string) {
    try {
      await deleteStaffMut({ institutionId: instId, staffId: id }).unwrap();
    } catch {
      toast.error(t({ ru: "Не удалось удалить", tg: "Нест карда нашуд" }));
    }
  }

  // ── vacancies (реальный backend: create/update/delete, FR-36) ──
  const { data: vacanciesData } = useListMyVacanciesQuery(instId);
  const vacancyList = vacanciesData?.items ?? [];
  const [vacancyForm, setVacancyForm] = useState<Partial<VacancyDTO> | null>(null);
  const [createVacancy] = useCreateVacancyMutation();
  const [updateVacancyMut] = useUpdateVacancyMutation();
  const [deleteVacancyMut] = useDeleteVacancyMutation();
  const employmentFullTime = t({ ru: "Полная занятость", tg: "Шуғли пурра" });

  function newVacancy() {
    setVacancyForm({ title: { ru: "", tg: "" }, description: { ru: "", tg: "" }, employment: { ru: employmentFullTime, tg: employmentFullTime }, status: "draft" });
  }
  async function saveVacancy() {
    if (!vacancyForm || !vacancyForm.title?.[locale]?.trim()) return;
    const body: VacancyRequest = {
      title: vacancyForm.title ?? { ru: "", tg: "" },
      description: vacancyForm.description ?? { ru: "", tg: "" },
      requirements: vacancyForm.requirements,
      salary_from: vacancyForm.salary_from,
      salary_to: vacancyForm.salary_to,
      employment: vacancyForm.employment ?? { ru: employmentFullTime, tg: employmentFullTime },
      status: vacancyForm.status ?? "draft",
    };
    try {
      if (vacancyForm.id) {
        await updateVacancyMut({ institutionId: instId, vacancyId: vacancyForm.id, body }).unwrap();
      } else {
        await createVacancy({ institutionId: instId, body }).unwrap();
      }
      setVacancyForm(null);
      toast.success(t({ ru: "Вакансия сохранена", tg: "Ҷои холӣ нигоҳ дошта шуд" }));
    } catch {
      toast.error(t({ ru: "Не удалось сохранить", tg: "Нигоҳ дошта нашуд" }));
    }
  }
  async function deleteVacancy(id: string) {
    try {
      await deleteVacancyMut({ institutionId: instId, vacancyId: id }).unwrap();
    } catch {
      toast.error(t({ ru: "Не удалось удалить", tg: "Нест карда нашуд" }));
    }
  }

  // ── achievements (реальный backend: create/delete, без update) ──
  const achievementsList = inst.achievements;
  const [achievementForm, setAchievementForm] = useState<Partial<AchievementDTO> | null>(null);
  const [createAchievement] = useCreateAchievementMutation();
  const [deleteAchievementMut] = useDeleteAchievementMutation();

  function newAchievement() {
    setAchievementForm({ title: { ru: "", tg: "" }, year: new Date().getFullYear(), category: "gold", description: { ru: "", tg: "" } });
  }
  async function saveAchievement() {
    if (!achievementForm || !achievementForm.title?.[locale]?.trim()) return;
    try {
      await createAchievement({
        institutionId: instId,
        body: {
          title: achievementForm.title ?? { ru: "", tg: "" }, year: achievementForm.year ?? new Date().getFullYear(),
          category: achievementForm.category ?? "gold", description: achievementForm.description ?? { ru: "", tg: "" },
        },
      }).unwrap();
      setAchievementForm(null);
      toast.success(t({ ru: "Достижение сохранено", tg: "Дастовард нигоҳ дошта шуд" }));
    } catch {
      toast.error(t({ ru: "Не удалось сохранить", tg: "Нигоҳ дошта нашуд" }));
    }
  }
  async function deleteAchievement(id: string) {
    try {
      await deleteAchievementMut({ institutionId: instId, achId: id }).unwrap();
    } catch {
      toast.error(t({ ru: "Не удалось удалить", tg: "Нест карда нашуд" }));
    }
  }

  // ── alumni (реальный backend: create/delete, без update) ──
  const alumniList = inst.alumni;
  const [alumnusForm, setAlumnusForm] = useState<Partial<AlumnusDTO> | null>(null);
  const [createAlumnus] = useCreateAlumnusMutation();
  const [deleteAlumnusMut] = useDeleteAlumnusMutation();

  function newAlumnus() {
    setAlumnusForm({ name: { ru: "", tg: "" }, grad_year: new Date().getFullYear() });
  }
  async function saveAlumnus() {
    if (!alumnusForm || !alumnusForm.name?.[locale]?.trim()) return;
    try {
      await createAlumnus({
        institutionId: instId,
        body: { name: alumnusForm.name ?? { ru: "", tg: "" }, grad_year: alumnusForm.grad_year ?? new Date().getFullYear(), now_label: alumnusForm.now_label },
      }).unwrap();
      setAlumnusForm(null);
      toast.success(t({ ru: "Выпускник сохранён", tg: "Хатмкунанда нигоҳ дошта шуд" }));
    } catch {
      toast.error(t({ ru: "Не удалось сохранить", tg: "Нигоҳ дошта нашуд" }));
    }
  }
  async function deleteAlumnus(id: string) {
    try {
      await deleteAlumnusMut({ institutionId: instId, alumnusId: id }).unwrap();
    } catch {
      toast.error(t({ ru: "Не удалось удалить", tg: "Нест карда нашуд" }));
    }
  }

  // ── gallery (реальный backend: create/delete, без update; s3_key — прямой URL,
  // настоящей загрузки файлов ещё нет, см. SRS E3.6) ──
  const galleryList = inst.gallery;
  const [galleryForm, setGalleryForm] = useState<Partial<GalleryItemDTO> | null>(null);
  const [createGalleryItem] = useCreateGalleryItemMutation();
  const [deleteGalleryItemMut] = useDeleteGalleryItemMutation();

  function newGalleryItem() {
    setGalleryForm({ s3_key: "", sort_order: galleryList.length });
  }
  async function saveGalleryItem() {
    if (!galleryForm || !galleryForm.s3_key?.trim()) return;
    try {
      await createGalleryItem({
        institutionId: instId,
        body: { s3_key: galleryForm.s3_key, label: galleryForm.label, sort_order: galleryForm.sort_order ?? 0 },
      }).unwrap();
      setGalleryForm(null);
      toast.success(t({ ru: "Фото добавлено в галерею", tg: "Акс ба галерея илова шуд" }));
    } catch {
      toast.error(t({ ru: "Не удалось сохранить", tg: "Нигоҳ дошта нашуд" }));
    }
  }
  async function deleteGalleryItem(id: string) {
    try {
      await deleteGalleryItemMut({ institutionId: instId, itemId: id }).unwrap();
    } catch {
      toast.error(t({ ru: "Не удалось удалить", tg: "Нест карда нашуд" }));
    }
  }

  // ── reviews (реальный backend: GET .../reviews/mine + POST /reviews/{id}/reply) ──
  const { data: reviewsData } = useListMyReviewsQuery(instId);
  const reviewsList = reviewsData?.items ?? [];
  const [replyDraftId, setReplyDraftId] = useState<string | null>(null);
  const [replyText, setReplyText] = useState("");
  const [replyToReview] = useReplyToReviewMutation();
  async function submitReply(id: string) {
    if (!replyText.trim()) return;
    try {
      await replyToReview({ institutionId: instId, reviewId: id, reply: replyText.trim() }).unwrap();
      setReplyDraftId(null);
      setReplyText("");
      toast.success(t({ ru: "Ответ опубликован", tg: "Ҷавоб нашр шуд" }));
    } catch {
      toast.error(t({ ru: "Не удалось отправить ответ", tg: "Ҷавоб фиристода нашуд" }));
    }
  }

  // ── settings (реальный backend: только скалярные поля, которые поддерживает
  // PATCH /api/v1/institutions/{id} — name/region/адрес backend не редактирует) ──
  const [settingsForm, setSettingsForm] = useState({
    description: inst.description ?? { ru: "", tg: "" },
    phone: inst.phone ?? "", email: inst.email ?? "", website: inst.website ?? "",
    price: inst.price ?? 0, ageRange: inst.age_range ?? "",
  });
  const [updateInstitution, { isLoading: savingSettings }] = useUpdateInstitutionMutation();

  async function saveSettings(e: React.FormEvent) {
    e.preventDefault();
    try {
      await updateInstitution({
        id: instId,
        body: {
          description: settingsForm.description, phone: settingsForm.phone || undefined, email: settingsForm.email || undefined,
          website: settingsForm.website || undefined, price: settingsForm.price || undefined, age_range: settingsForm.ageRange || undefined,
        },
      }).unwrap();
      toast.success(t({ ru: "Настройки сохранены", tg: "Танзимот нигоҳ дошта шуд" }));
    } catch {
      toast.error(t({ ru: "Не удалось сохранить", tg: "Нигоҳ дошта нашуд" }));
    }
  }

  return (
    <div className="eh-sidebar-shell" style={{ display: "flex", fontFamily: FB, background: C.bg, color: C.text }}>
      {/* ── SIDEBAR ── */}
      <aside style={{ width: 240, flexShrink: 0, borderRight: `1px solid ${C.border}`, padding: "20px 14px", display: "flex", flexDirection: "column" }}>
        <p style={{ padding: "0 10px", fontSize: 11, fontWeight: 700, color: C.dim, textTransform: "uppercase", letterSpacing: ".06em", marginBottom: 8 }}>{t({ ru: "Кабинет учреждения", tg: "Кабинети муассиса" })}</p>
        <p style={{ padding: "0 10px", fontSize: 13.5, fontWeight: 700, color: C.text, marginBottom: 20, lineHeight: 1.3 }}>{inst.name[locale]}</p>

        <nav style={{ display: "flex", flexDirection: "column", gap: 2, flex: 1 }}>
          {TAB_META.map(({ k, labelKey, icon: Icon }) => (
            <Button className="h-auto w-auto" key={k} onClick={() => setTab(k)} style={{ display: "flex", alignItems: "center", gap: 10, padding: "10px 12px", borderRadius: 10, fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: tab === k ? C.teal : C.sub, background: tab === k ? `${C.teal}18` : "transparent", border: "none", cursor: "pointer", textAlign: "left" }}>
              <Icon size={16} /> {t(labelKey)}
            </Button>
          ))}
        </nav>

        <Link href="/" style={{ display: "flex", alignItems: "center", gap: 10, padding: "10px 12px", borderRadius: 10, fontFamily: FH, fontWeight: 700, fontSize: 13, color: C.muted, textDecoration: "none" }}>
          <LogOut size={15} /> {t("nav.home")}
        </Link>
      </aside>

      {/* ── CONTENT ── */}
      <div style={{ flex: 1, padding: "28px 32px 80px", overflowX: "hidden" }}>
        <Button className="h-auto w-auto" onClick={() => (window.history.length > 1 ? router.back() : router.push("/"))} style={{ display: "inline-flex", alignItems: "center", gap: 7, fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: C.teal, marginBottom: 20, padding: "8px 14px", borderRadius: 9, border: `1px solid ${C.teal}40`, background: `${C.teal}10`, cursor: "pointer" }}>
          <ArrowLeft size={15} /> {t("common.back")}
        </Button>

        {tab === "overview" && (
          <div>
            <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24, marginBottom: 22 }}>{t("dash.overview")}</h1>
            <div className="eh-mobile-1col" style={{ display: "grid", gridTemplateColumns: "repeat(4,1fr)", gap: 14, marginBottom: 28 }}>
              {[
                { l: t({ ru: "Рейтинг", tg: "Рейтинг" }), v: inst.rating_avg?.toFixed(1) ?? "—", sub: `${inst.review_count} ${t("common.reviews")}` },
                { l: t({ ru: "Учеников", tg: "Хонандагон" }), v: inst.students_count ?? "—", sub: inst.founded ? `${t({ ru: "с", tg: "аз" })} ${inst.founded}` : "" },
                { l: t("tab.news"), v: articles.length, sub: `${articles.filter((a) => a.status === "published").length} ${t({ ru: "опубликовано", tg: "нашр шуд" })}` },
                { l: t("tab.staff"), v: staffList.length, sub: t({ ru: "в профиле", tg: "дар профил" }) },
              ].map((s) => (
                <div key={s.l} style={{ borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: 18 }}>
                  <p style={{ fontFamily: FH, fontWeight: 900, fontSize: 26, color: C.teal }}>{s.v}</p>
                  <p style={{ fontSize: 13, color: C.text, fontWeight: 600, marginTop: 2 }}>{s.l}</p>
                  <p style={{ fontSize: 11.5, color: C.muted, marginTop: 2 }}>{s.sub}</p>
                </div>
              ))}
            </div>
            <div style={{ borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: 20 }}>
              <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 15, marginBottom: 4 }}>{t({ ru: "Просмотры и отклики за неделю (демо-график)", tg: "Дидан ва ҷавобҳо дар як ҳафта (нақшаи демо)" })}</h3>
              <p style={{ fontSize: 12, color: C.muted, marginBottom: 14 }}>{t({ ru: "Аналитика ещё не подключена к backend (веха 6)", tg: "Таҳлилот ҳанӯз ба backend пайваст нашудааст" })}</p>
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
                    <Area type="monotone" dataKey="views" name={t({ ru: "Просмотры", tg: "Дидан" })} stroke={C.teal} strokeWidth={2} fill="url(#viewsFill)" />
                    <Area type="monotone" dataKey="responses" name={t({ ru: "Отклики", tg: "Ҷавобҳо" })} stroke={C.gold} strokeWidth={2} fill="none" />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>
        )}

        {tab === "news" && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
              <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24 }}>{t("tab.news")}</h1>
              <Button className="h-auto w-auto" onClick={newArticle} style={{ display: "flex", alignItems: "center", gap: 7, padding: "10px 18px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                <Plus size={15} /> {t("common.add")}
              </Button>
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {articles.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t("empty.news")}</p>}
              {articles.map((a) => (
                <div key={a.id} style={{ display: "flex", alignItems: "center", gap: 14, borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "12px 16px" }}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{a.title[locale]}</p>
                    <p style={{ fontSize: 12, color: C.muted, marginTop: 2 }}>
                      <span style={{ color: a.status === "published" ? C.ok : C.gold }}>{a.status === "published" ? t({ ru: "Опубликовано", tg: "Нашр шудааст" }) : t({ ru: "Черновик", tg: "Пешнавис" })}</span>
                    </p>
                  </div>
                  <Button className="h-auto w-auto" onClick={() => setNewsForm(a)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.sub }}><Pencil size={14} /></Button>
                  <Button className="h-auto w-auto" onClick={() => deleteArticle(a.id)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.red }}><Trash2 size={14} /></Button>
                </div>
              ))}
            </div>

            <Modal open={!!newsForm} onClose={() => setNewsForm(null)} maxWidth={520}>
              {newsForm && (
                <div style={{ padding: 26 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                    <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18 }}>{newsForm.id ? t({ ru: "Редактировать новость", tg: "Хабарро таҳрир кардан" }) : t({ ru: "Новая новость", tg: "Хабари нав" })}</h3>
                    <Button className="h-auto w-auto" onClick={() => setNewsForm(null)} style={{ background: "none", border: "none", color: C.sub, cursor: "pointer" }}><X size={18} /></Button>
                  </div>
                  <FormField label={`${t({ ru: "Заголовок", tg: "Сарлавҳа" })} (${locale.toUpperCase()})`}><Input value={newsForm.title?.[locale] ?? ""} onChange={(e) => setNewsForm({ ...newsForm, title: { ru: newsForm.title?.ru ?? "", tg: newsForm.title?.tg ?? "", [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={`${t({ ru: "Текст", tg: "Матн" })} (${locale.toUpperCase()})`}><Textarea value={newsForm.content?.[locale] ?? ""} onChange={(e) => setNewsForm({ ...newsForm, content: { ru: newsForm.content?.ru ?? "", tg: newsForm.content?.tg ?? "", [locale]: e.target.value } })} style={{ ...inputStyle, height: 100, resize: "none" as const }} /></FormField>
                  <div style={{ display: "flex", gap: 10, alignItems: "center", marginTop: 6, marginBottom: 20 }}>
                    <label style={{ fontSize: 13, color: C.sub, fontFamily: FB }}>{t({ ru: "Статус:", tg: "Ҳолат:" })}</label>
                    <Select value={newsForm.status ?? "draft"} onValueChange={(v) => setNewsForm({ ...newsForm, status: v as "draft" | "published" })}>
                      <SelectTrigger style={{ ...inputStyle, width: "auto", padding: "8px 12px" }} className="h-auto w-auto"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="draft">{t({ ru: "Черновик", tg: "Пешнавис" })}</SelectItem>
                        <SelectItem value="published">{t({ ru: "Опубликовано", tg: "Нашр шудааст" })}</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <Button className="h-auto w-auto" onClick={saveArticle} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer" }}>{t("common.save")}</Button>
                </div>
              )}
            </Modal>
          </div>
        )}

        {tab === "staff" && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
              <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24 }}>{t("tab.staff")}</h1>
              <Button className="h-auto w-auto" onClick={newStaff} style={{ display: "flex", alignItems: "center", gap: 7, padding: "10px 18px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                <Plus size={15} /> {t("common.add")}
              </Button>
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {staffList.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t("empty.staff")}</p>}
              {staffList.map((p) => (
                <div key={p.id} style={{ display: "flex", alignItems: "center", gap: 14, borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "12px 16px" }}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text }}>{p.name[locale]}</p>
                    <p style={{ fontSize: 12, color: C.muted, marginTop: 2 }}>{p.role_label[locale]}{p.subject ? ` · ${p.subject[locale]}` : ""}</p>
                  </div>
                  <Button className="h-auto w-auto" onClick={() => setStaffForm(p)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.sub }}><Pencil size={14} /></Button>
                  <Button className="h-auto w-auto" onClick={() => deleteStaff(p.id)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.red }}><Trash2 size={14} /></Button>
                </div>
              ))}
            </div>

            <Modal open={!!staffForm} onClose={() => setStaffForm(null)} maxWidth={520}>
              {staffForm && (
                <div style={{ padding: 26 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                    <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18 }}>{staffForm.id ? t({ ru: "Редактировать сотрудника", tg: "Корманд таҳрир кардан" }) : t({ ru: "Новый сотрудник", tg: "Корманди нав" })}</h3>
                    <Button className="h-auto w-auto" onClick={() => setStaffForm(null)} style={{ background: "none", border: "none", color: C.sub, cursor: "pointer" }}><X size={18} /></Button>
                  </div>
                  <FormField label={`${t({ ru: "Имя", tg: "Ном" })} (${locale.toUpperCase()})`}><Input value={staffForm.name?.[locale] ?? ""} onChange={(e) => setStaffForm({ ...staffForm, name: { ru: staffForm.name?.ru ?? "", tg: staffForm.name?.tg ?? "", [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={`${t({ ru: "Должность", tg: "Вазифа" })} (${locale.toUpperCase()})`}><Input value={staffForm.role_label?.[locale] ?? ""} onChange={(e) => setStaffForm({ ...staffForm, role_label: { ru: staffForm.role_label?.ru ?? "", tg: staffForm.role_label?.tg ?? "", [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={t({ ru: "Опыт", tg: "Таҷриба" })}><Input value={staffForm.exp ?? ""} onChange={(e) => setStaffForm({ ...staffForm, exp: e.target.value })} style={inputStyle} /></FormField>
                  <FormField label={`${t({ ru: "О сотруднике", tg: "Дар бораи корманд" })} (${locale.toUpperCase()})`}><Textarea value={staffForm.bio?.[locale] ?? ""} onChange={(e) => setStaffForm({ ...staffForm, bio: { ru: staffForm.bio?.ru ?? "", tg: staffForm.bio?.tg ?? "", [locale]: e.target.value } })} style={{ ...inputStyle, height: 90, resize: "none" as const }} /></FormField>
                  <FormField label={t({ ru: "Email", tg: "Почта" })}><Input value={staffForm.email ?? ""} onChange={(e) => setStaffForm({ ...staffForm, email: e.target.value })} style={inputStyle} /></FormField>
                  <Button className="h-auto w-auto" onClick={saveStaff} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer", marginTop: 6 }}>{t("common.save")}</Button>
                </div>
              )}
            </Modal>
          </div>
        )}

        {tab === "vacancies" && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
              <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24 }}>{t("tab.vacancies")}</h1>
              <Button className="h-auto w-auto" onClick={newVacancy} style={{ display: "flex", alignItems: "center", gap: 7, padding: "10px 18px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                <Plus size={15} /> {t("common.add")}
              </Button>
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {vacancyList.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t("empty.vacancies")}</p>}
              {vacancyList.map((v) => (
                <div key={v.id} style={{ borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "12px 16px" }}>
                  <div style={{ display: "flex", alignItems: "center", gap: 14 }}>
                    <div style={{ width: 44, height: 44, borderRadius: 10, background: `${C.teal}18`, display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
                      <Briefcase size={19} style={{ color: C.teal }} />
                    </div>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{v.title[locale]}</p>
                      <p style={{ fontSize: 12, color: C.muted, marginTop: 2 }}>
                        {new Date(v.created_at).toLocaleDateString("ru-RU")} · <span style={{ color: v.status === "published" ? C.ok : C.gold }}>{v.status === "published" ? t({ ru: "Опубликовано", tg: "Нашр шудааст" }) : t({ ru: "Черновик", tg: "Пешнавис" })}</span>
                      </p>
                    </div>
                    <Button className="h-auto w-auto" onClick={() => setVacancyForm(v)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.sub }}><Pencil size={14} /></Button>
                    <Button className="h-auto w-auto" onClick={() => deleteVacancy(v.id)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.red }}><Trash2 size={14} /></Button>
                  </div>
                </div>
              ))}
            </div>

            <Modal open={!!vacancyForm} onClose={() => setVacancyForm(null)} maxWidth={520}>
              {vacancyForm && (
                <div style={{ padding: 26 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                    <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18 }}>{vacancyForm.id ? t({ ru: "Редактировать вакансию", tg: "Ҷои холиро таҳрир кардан" }) : t({ ru: "Новая вакансия", tg: "Ҷои холии нав" })}</h3>
                    <Button className="h-auto w-auto" onClick={() => setVacancyForm(null)} style={{ background: "none", border: "none", color: C.sub, cursor: "pointer" }}><X size={18} /></Button>
                  </div>
                  <FormField label={`${t({ ru: "Должность", tg: "Вазифа" })} (${locale.toUpperCase()})`}><Input value={vacancyForm.title?.[locale] ?? ""} onChange={(e) => setVacancyForm({ ...vacancyForm, title: { ...(vacancyForm.title ?? { ru: "", tg: "" }), [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={`${t({ ru: "Описание", tg: "Тавсиф" })} (${locale.toUpperCase()})`}><Textarea value={vacancyForm.description?.[locale] ?? ""} onChange={(e) => setVacancyForm({ ...vacancyForm, description: { ...(vacancyForm.description ?? { ru: "", tg: "" }), [locale]: e.target.value } })} style={{ ...inputStyle, height: 90, resize: "none" as const }} /></FormField>
                  <div style={{ display: "flex", gap: 10, alignItems: "center", marginTop: 6, marginBottom: 20 }}>
                    <label style={{ fontSize: 13, color: C.sub, fontFamily: FB }}>{t({ ru: "Статус:", tg: "Ҳолат:" })}</label>
                    <Select value={vacancyForm.status} onValueChange={(v) => setVacancyForm({ ...vacancyForm, status: v as VacancyDTO["status"] })}>
                      <SelectTrigger style={{ ...inputStyle, width: "auto", padding: "8px 12px" }} className="h-auto w-auto"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="draft">{t({ ru: "Черновик", tg: "Пешнавис" })}</SelectItem>
                        <SelectItem value="published">{t({ ru: "Опубликовано", tg: "Нашр шудааст" })}</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <Button className="h-auto w-auto" onClick={saveVacancy} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer" }}>{t("common.save")}</Button>
                </div>
              )}
            </Modal>
          </div>
        )}

        {tab === "achievements" && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
              <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24 }}>{t("tab.achievements")}</h1>
              <Button className="h-auto w-auto" onClick={newAchievement} style={{ display: "flex", alignItems: "center", gap: 7, padding: "10px 18px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                <Plus size={15} /> {t("common.add")}
              </Button>
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {achievementsList.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t("empty.achievements")}</p>}
              {achievementsList.map((a) => (
                <div key={a.id} style={{ display: "flex", alignItems: "center", gap: 14, borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "12px 16px" }}>
                  <div style={{ width: 44, height: 44, borderRadius: 12, background: `${C.gold}18`, display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
                    <Award size={19} style={{ color: C.gold }} />
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{a.title[locale]}</p>
                    <p style={{ fontSize: 12, color: C.muted, marginTop: 2 }}>{a.year} · {a.category}</p>
                  </div>
                  <Button className="h-auto w-auto" onClick={() => deleteAchievement(a.id)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.red }}><Trash2 size={14} /></Button>
                </div>
              ))}
            </div>

            <Modal open={!!achievementForm} onClose={() => setAchievementForm(null)} maxWidth={480}>
              {achievementForm && (
                <div style={{ padding: 26 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                    <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18 }}>{t({ ru: "Новое достижение", tg: "Дастоварди нав" })}</h3>
                    <Button className="h-auto w-auto" onClick={() => setAchievementForm(null)} style={{ background: "none", border: "none", color: C.sub, cursor: "pointer" }}><X size={18} /></Button>
                  </div>
                  <FormField label={`${t({ ru: "Название", tg: "Ном" })} (${locale.toUpperCase()})`}><Input value={achievementForm.title?.[locale] ?? ""} onChange={(e) => setAchievementForm({ ...achievementForm, title: { ru: achievementForm.title?.ru ?? "", tg: achievementForm.title?.tg ?? "", [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={`${t({ ru: "Описание", tg: "Тавсиф" })} (${locale.toUpperCase()})`}><Textarea value={achievementForm.description?.[locale] ?? ""} onChange={(e) => setAchievementForm({ ...achievementForm, description: { ru: achievementForm.description?.ru ?? "", tg: achievementForm.description?.tg ?? "", [locale]: e.target.value } })} style={{ ...inputStyle, height: 80, resize: "none" as const }} /></FormField>
                  <div style={{ display: "flex", gap: 10 }}>
                    <div style={{ flex: 1 }}>
                      <FormField label={t({ ru: "Год", tg: "Сол" })}><Input type="number" value={achievementForm.year ?? new Date().getFullYear()} onChange={(e) => setAchievementForm({ ...achievementForm, year: Number(e.target.value) })} style={inputStyle} /></FormField>
                    </div>
                    <div style={{ flex: 1 }}>
                      <FormField label={t({ ru: "Уровень", tg: "Дараҷа" })}>
                        <Select value={achievementForm.category ?? "gold"} onValueChange={(v) => setAchievementForm({ ...achievementForm, category: v })}>
                          <SelectTrigger style={inputStyle} className="h-auto w-auto"><SelectValue /></SelectTrigger>
                          <SelectContent>
                            <SelectItem value="gold">{t({ ru: "Золото", tg: "Тилло" })}</SelectItem>
                            <SelectItem value="silver">{t({ ru: "Серебро", tg: "Нуқра" })}</SelectItem>
                            <SelectItem value="bronze">{t({ ru: "Бронза", tg: "Биринҷӣ" })}</SelectItem>
                            <SelectItem value="special">{t({ ru: "Особое", tg: "Махсус" })}</SelectItem>
                          </SelectContent>
                        </Select>
                      </FormField>
                    </div>
                  </div>
                  <Button className="h-auto w-auto" onClick={saveAchievement} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer", marginTop: 6 }}>{t("common.save")}</Button>
                </div>
              )}
            </Modal>
          </div>
        )}

        {tab === "alumni" && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
              <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24 }}>{t("tab.alumni")}</h1>
              <Button className="h-auto w-auto" onClick={newAlumnus} style={{ display: "flex", alignItems: "center", gap: 7, padding: "10px 18px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                <Plus size={15} /> {t("common.add")}
              </Button>
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {alumniList.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t({ ru: "Выпускники не добавлены", tg: "Хатмкунандагон илова нашудаанд" })}</p>}
              {alumniList.map((a) => (
                <div key={a.id} style={{ display: "flex", alignItems: "center", gap: 14, borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "12px 16px" }}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text }}>{a.name[locale]}</p>
                    <p style={{ fontSize: 12, color: C.muted, marginTop: 2 }}>{a.grad_year}{a.now_label ? ` · ${a.now_label[locale]}` : ""}</p>
                  </div>
                  <Button className="h-auto w-auto" onClick={() => deleteAlumnus(a.id)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 8, cursor: "pointer", color: C.red }}><Trash2 size={14} /></Button>
                </div>
              ))}
            </div>

            <Modal open={!!alumnusForm} onClose={() => setAlumnusForm(null)} maxWidth={480}>
              {alumnusForm && (
                <div style={{ padding: 26 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                    <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18 }}>{t({ ru: "Новый выпускник", tg: "Хатмкунандаи нав" })}</h3>
                    <Button className="h-auto w-auto" onClick={() => setAlumnusForm(null)} style={{ background: "none", border: "none", color: C.sub, cursor: "pointer" }}><X size={18} /></Button>
                  </div>
                  <FormField label={`${t({ ru: "Имя", tg: "Ном" })} (${locale.toUpperCase()})`}><Input value={alumnusForm.name?.[locale] ?? ""} onChange={(e) => setAlumnusForm({ ...alumnusForm, name: { ru: alumnusForm.name?.ru ?? "", tg: alumnusForm.name?.tg ?? "", [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <FormField label={t({ ru: "Год выпуска", tg: "Соли хатм" })}><Input type="number" value={alumnusForm.grad_year ?? new Date().getFullYear()} onChange={(e) => setAlumnusForm({ ...alumnusForm, grad_year: Number(e.target.value) })} style={inputStyle} /></FormField>
                  <FormField label={`${t({ ru: "Где сейчас", tg: "Ҳозир дар куҷо" })} (${locale.toUpperCase()})`}><Input value={alumnusForm.now_label?.[locale] ?? ""} onChange={(e) => setAlumnusForm({ ...alumnusForm, now_label: { ru: alumnusForm.now_label?.ru ?? "", tg: alumnusForm.now_label?.tg ?? "", [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <Button className="h-auto w-auto" onClick={saveAlumnus} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer", marginTop: 6 }}>{t("common.save")}</Button>
                </div>
              )}
            </Modal>
          </div>
        )}

        {tab === "gallery" && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
              <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24 }}>{t("tab.gallery")}</h1>
              <Button className="h-auto w-auto" onClick={newGalleryItem} style={{ display: "flex", alignItems: "center", gap: 7, padding: "10px 18px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                <Plus size={15} /> {t("common.add")}
              </Button>
            </div>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill,minmax(160px,1fr))", gap: 12 }}>
              {galleryList.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t({ ru: "Галерея пуста", tg: "Галерея холист" })}</p>}
              {galleryList.map((g) => (
                <div key={g.id} style={{ borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, overflow: "hidden" }}>
                  {/* eslint-disable-next-line @next/next/no-img-element -- s3_key is an arbitrary external URL until multipart upload (E3.6) ships */}
                  <img src={g.s3_key} alt="" style={{ width: "100%", height: 100, objectFit: "cover" }} />
                  <div style={{ padding: "8px 10px", display: "flex", alignItems: "center", gap: 6 }}>
                    <p style={{ flex: 1, minWidth: 0, fontSize: 12, color: C.text, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{g.label?.[locale] ?? ""}</p>
                    <Button className="h-auto w-auto" onClick={() => deleteGalleryItem(g.id)} style={{ background: C.s3, border: "none", borderRadius: 7, padding: 6, cursor: "pointer", color: C.red }}><Trash2 size={11} /></Button>
                  </div>
                </div>
              ))}
            </div>

            <Modal open={!!galleryForm} onClose={() => setGalleryForm(null)} maxWidth={440}>
              {galleryForm && (
                <div style={{ padding: 26 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                    <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18 }}>{t({ ru: "Фото галереи", tg: "Акси галерея" })}</h3>
                    <Button className="h-auto w-auto" onClick={() => setGalleryForm(null)} style={{ background: "none", border: "none", color: C.sub, cursor: "pointer" }}><X size={18} /></Button>
                  </div>
                  <FormField label={t({ ru: "Ссылка на фото (URL)", tg: "Пайванд ба акс (URL)" })}><Input value={galleryForm.s3_key ?? ""} onChange={(e) => setGalleryForm({ ...galleryForm, s3_key: e.target.value })} placeholder="https://…" style={inputStyle} /></FormField>
                  <FormField label={`${t({ ru: "Подпись", tg: "Изоҳ" })} (${locale.toUpperCase()})`}><Input value={galleryForm.label?.[locale] ?? ""} onChange={(e) => setGalleryForm({ ...galleryForm, label: { ru: galleryForm.label?.ru ?? "", tg: galleryForm.label?.tg ?? "", [locale]: e.target.value } })} style={inputStyle} /></FormField>
                  <Button className="h-auto w-auto" onClick={saveGalleryItem} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer", marginTop: 6 }}>{t("common.save")}</Button>
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
                    <p style={{ fontSize: 12, color: C.sub }}>
                      {new Date(r.created_at).toLocaleDateString("ru-RU")} · <span style={{ color: r.status === "approved" ? C.ok : r.status === "rejected" ? C.red : C.gold }}>{r.status}</span>
                    </p>
                    <span style={{ color: C.gold, fontFamily: FH, fontWeight: 700 }}>{r.rating} ★</span>
                  </div>
                  <p style={{ fontSize: 13.5, color: C.sub, lineHeight: 1.6, marginBottom: 10 }}>{r.text}</p>
                  {r.reply && (
                    <div style={{ padding: "10px 14px", borderRadius: 10, background: `${C.teal}12`, marginBottom: 10 }}>
                      <p style={{ fontSize: 12, fontWeight: 700, color: C.teal, fontFamily: FH, marginBottom: 4 }}>{t({ ru: "Ваш ответ:", tg: "Ҷавоби шумо:" })}</p>
                      <p style={{ fontSize: 13, color: C.sub }}>{r.reply}</p>
                    </div>
                  )}
                  {replyDraftId === r.id ? (
                    <div style={{ display: "flex", gap: 8 }}>
                      <Input value={replyText} onChange={(e) => setReplyText(e.target.value)} placeholder={t({ ru: "Ваш ответ…", tg: "Ҷавоби шумо…" })} style={{ ...inputStyle, flex: 1 }} />
                      <Button className="h-auto w-auto" onClick={() => submitReply(r.id)} style={{ padding: "0 16px", borderRadius: 10, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 700, fontSize: 13, border: "none", cursor: "pointer" }}>{t("common.send")}</Button>
                    </div>
                  ) : (
                    <Button className="h-auto w-auto" onClick={() => { setReplyDraftId(r.id); setReplyText(r.reply ?? ""); }} style={{ fontSize: 12.5, fontWeight: 700, color: C.teal, fontFamily: FH, background: "none", border: "none", cursor: "pointer" }}>
                      {r.reply ? t({ ru: "Изменить ответ", tg: "Ҷавобро тағйир додан" }) : t("common.reply")}
                    </Button>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {tab === "settings" && (
          <div>
            <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: 24, marginBottom: 20 }}>{t("dash.settings")}</h1>
            <p style={{ fontSize: 12, color: C.muted, marginBottom: 20 }}>{t({ ru: "Название, регион и адрес пока не редактируются — обратитесь в поддержку", tg: "Ном, минтақа ва суроға ҳанӯз таҳрир намешаванд" })}</p>
            <form onSubmit={saveSettings} style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 26, maxWidth: 520 }}>
              <FormField label={`${t({ ru: "Описание", tg: "Тавсиф" })} (${locale.toUpperCase()})`}>
                <Textarea value={settingsForm.description[locale]} onChange={(e) => setSettingsForm({ ...settingsForm, description: { ...settingsForm.description, [locale]: e.target.value } })} style={{ ...inputStyle, height: 100, resize: "none" as const }} />
              </FormField>
              <FormField label={t({ ru: "Телефон", tg: "Телефон" })}><Input value={settingsForm.phone} onChange={(e) => setSettingsForm({ ...settingsForm, phone: e.target.value })} style={inputStyle} /></FormField>
              <FormField label={t({ ru: "Эл. почта", tg: "Почтаи электронӣ" })}><Input value={settingsForm.email} onChange={(e) => setSettingsForm({ ...settingsForm, email: e.target.value })} style={inputStyle} /></FormField>
              <FormField label={t({ ru: "Официальный сайт", tg: "Сомонаи расмӣ" })}><Input value={settingsForm.website} onChange={(e) => setSettingsForm({ ...settingsForm, website: e.target.value })} style={inputStyle} /></FormField>
              <FormField label={t({ ru: "Цена, сомони/мес", tg: "Нарх, сомонӣ/моҳ" })}><Input type="number" value={settingsForm.price} onChange={(e) => setSettingsForm({ ...settingsForm, price: Number(e.target.value) })} style={inputStyle} /></FormField>
              <FormField label={t({ ru: "Возрастная группа", tg: "Гурӯҳи синнусолӣ" })}><Input value={settingsForm.ageRange} onChange={(e) => setSettingsForm({ ...settingsForm, ageRange: e.target.value })} style={inputStyle} /></FormField>
              <Button className="h-auto w-auto" type="submit" disabled={savingSettings} style={{ marginTop: 20, padding: "12px 24px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer", opacity: savingSettings ? 0.6 : 1 }}>
                {savingSettings ? t({ ru: "Сохранение…", tg: "Нигоҳ дошта истодааст…" }) : t("common.save")}
              </Button>
            </form>
          </div>
        )}
      </div>
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
