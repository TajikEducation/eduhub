"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { Award, Baby, Bookmark, Briefcase, ChevronRight, FileText, GraduationCap, Medal, MessageSquare, Pencil, Plus, Sparkles, Star, Trash2, Trophy, Upload, X, type LucideIcon } from "lucide-react";
import { C, FH, FB, INSTITUTIONS, REVIEWS, VACANCIES, type Achievement, type Bi, type ApplicantVisibility } from "@/lib/data";
import { useAppState, type ChildStatus } from "@/lib/app-state";
import { chatHref } from "@/lib/chat-window";
import { useT, useLocale } from "@/lib/i18n";
import { Toast } from "./Toast";
import { Modal } from "./ui/Modal";

type Tab = "applicant" | "saved" | "reviews" | "messages" | "children";
type ApplicantSubTab = "profile" | "achievements" | "applications";

const MY_REVIEW_IDS = ["r1", "r6", "r9", "r12", "r14"];
const CHILD_STATUS_KEY: Record<ChildStatus, string> = {
  current: "children.statusCurrent",
  alumnus: "children.statusAlumnus",
  transferred: "children.statusTransferred",
};
const ACH_TIER: Record<string, { icon: LucideIcon; color: string }> = {
  gold: { icon: Medal, color: C.gold }, silver: { icon: Medal, color: C.sub }, bronze: { icon: Medal, color: C.goldD }, special: { icon: Trophy, color: C.teal },
};
const bi = (ru: string, tg: string): Bi => ({ ru, tg });
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

function ListEditor({ items, onChange, placeholder }: { items: Bi[]; onChange: (next: Bi[]) => void; placeholder: string }) {
  const { locale } = useLocale();
  const t = useT();
  const [draft, setDraft] = useState("");
  function add() {
    if (!draft.trim()) return;
    onChange([...items, bi(draft.trim(), draft.trim())]);
    setDraft("");
  }
  return (
    <div>
      <div style={{ display: "flex", flexDirection: "column", gap: 8, marginBottom: 10 }}>
        {items.map((it, i) => (
          <div key={i} style={{ display: "flex", alignItems: "center", gap: 10, borderRadius: 10, border: `1px solid ${C.border}`, background: C.s2, padding: "9px 12px" }}>
            <span style={{ flex: 1, fontSize: 13.5, color: C.text }}>{it[locale]}</span>
            <button onClick={() => onChange(items.filter((_, j) => j !== i))} style={{ background: "none", border: "none", color: C.muted, cursor: "pointer" }}><Trash2 size={14} /></button>
          </div>
        ))}
      </div>
      <div style={{ display: "flex", gap: 8 }}>
        <input value={draft} onChange={(e) => setDraft(e.target.value)} placeholder={placeholder} style={{ ...inputStyle, flex: 1 }} onKeyDown={(e) => e.key === "Enter" && add()} />
        <button onClick={add} style={{ display: "flex", alignItems: "center", gap: 6, padding: "9px 16px", borderRadius: 10, background: C.s3, color: C.text, fontFamily: FH, fontWeight: 700, fontSize: 13, border: "none", cursor: "pointer", whiteSpace: "nowrap" }}>
          <Plus size={14} /> {t("common.add")}
        </button>
      </div>
    </div>
  );
}

// Единый кабинет для любого вошедшего пользователя (RBAC-роль "user"). "Родитель"
// и "соискатель" — не отдельные роли/варианты кабинета, а факты о пользователе
// (есть ли привязанный ребёнок / опубликован ли профиль резюме) — модель B.
export function UserCabinet() {
  const router = useRouter();
  const { savedIds, toggleSaved, children_: children, addChild, removeChild, applicant, setApplicant, applications, employerResponses } = useAppState();
  const myResponses = employerResponses.filter((r) => r.applicantId === applicant.id);
  const t = useT();
  const { locale } = useLocale();
  const [tab, setTab] = useState<Tab>("saved");
  const [toast, setToast] = useState<string | null>(null);
  const [childName, setChildName] = useState("");
  const [childAge, setChildAge] = useState("");
  const [childInstId, setChildInstId] = useState<number | "">("");
  const [childInstQuery, setChildInstQuery] = useState("");
  const [childInstOpen, setChildInstOpen] = useState(false);
  const childInstBoxRef = useRef<HTMLDivElement>(null);
  const [childStatus, setChildStatus] = useState<ChildStatus>("current");
  const [name] = useState("Мадина Юсупова");
  const [applicantSubTab, setApplicantSubTab] = useState<ApplicantSubTab>("profile");
  const [achievementForm, setAchievementForm] = useState<Achievement | null>(null);
  const photoInputRef = useRef<HTMLInputElement>(null);
  const cvInputRef = useRef<HTMLInputElement>(null);

  function onPhotoSelected(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => setApplicant((prev) => ({ ...prev, photo: reader.result as string }));
    reader.readAsDataURL(file);
  }

  // Прототип без бэкенда: сохраняем только имя файла (не сам бинарник — модерация/антивирус/S3
  // нужны на реальном сервере).
  function onCvSelected(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setApplicant((prev) => ({ ...prev, cvFileName: file.name }));
    setToast(t("common.saved"));
  }

  function newAchievement() {
    setAchievementForm({ id: `ach-app-${Date.now()}`, title: bi("", ""), year: new Date().getFullYear(), category: "gold", desc: bi("", ""), type: "teacher" });
  }
  function saveAchievement() {
    if (!achievementForm || !achievementForm.title[locale].trim()) return;
    setApplicant((prev) => {
      const exists = prev.achievements.some((a) => a.id === achievementForm.id);
      const achievements = exists
        ? prev.achievements.map((a) => (a.id === achievementForm.id ? achievementForm : a))
        : [achievementForm, ...prev.achievements];
      return { ...prev, achievements };
    });
    setAchievementForm(null);
    setToast(t("common.saved"));
  }
  function deleteAchievement(id: string) {
    setApplicant((prev) => ({ ...prev, achievements: prev.achievements.filter((a) => a.id !== id) }));
  }

  function setVisibility(v: ApplicantVisibility) {
    setApplicant((prev) => ({ ...prev, visibility: v }));
    setToast(t("common.saved"));
  }

  const savedInstitutions = INSTITUTIONS.filter((i) => savedIds.includes(i.id));
  const myReviews = REVIEWS.filter((r) => MY_REVIEW_IDS.includes(r.id));

  const TABS: { k: Tab; label: string; icon: LucideIcon }[] = [
    { k: "applicant", label: t({ru:"Профиль",tg:"Профил"}), icon: Briefcase },
    { k: "saved", label: t({ru:"Избранное",tg:"Интихобшуда"}), icon: Bookmark },
    { k: "reviews", label: t({ru:"Мои отзывы",tg:"Шарҳҳои ман"}), icon: Star },
    { k: "messages", label: t({ru:"Сообщения",tg:"Паёмҳо"}), icon: MessageSquare },
    { k: "children", label: t({ru:"Дети",tg:"Фарзандон"}), icon: Baby },
  ];

  function submitChild() {
    if (!childName.trim() || childInstId === "") return;
    addChild({ name: childName.trim(), age: childAge.trim() || "—", instId: childInstId, status: childStatus });
    setChildName("");
    setChildAge("");
    setChildInstId("");
    setChildInstQuery("");
    setChildStatus("current");
  }

  const childInstMatches = childInstQuery.trim()
    ? INSTITUTIONS.filter((i) => t(i.name).toLowerCase().includes(childInstQuery.trim().toLowerCase())).slice(0, 6)
    : INSTITUTIONS.slice(0, 6);

  function pickChildInst(i: (typeof INSTITUTIONS)[number]) {
    setChildInstId(i.id);
    setChildInstQuery(t(i.name));
    setChildInstOpen(false);
  }

  useEffect(() => {
    function onClick(e: MouseEvent) {
      if (childInstBoxRef.current && !childInstBoxRef.current.contains(e.target as Node)) setChildInstOpen(false);
    }
    document.addEventListener("mousedown", onClick);
    return () => document.removeEventListener("mousedown", onClick);
  }, []);

  return (
    <div style={{ maxWidth: 1080, margin: "0 auto", padding: "28px 28px 90px", fontFamily: FB }}>
      <div style={{ marginBottom: 28 }}>
        <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(22px,3vw,30px)", color: C.text, letterSpacing: "-.02em" }}>
          {t({ru:"Личный кабинет",tg:"Кабинети шахсӣ"})}
        </h1>
        <p style={{ fontSize: 14, color: C.sub, marginTop: 4 }}>{name}</p>
      </div>

      <div style={{ display: "flex", gap: 2, borderBottom: `1px solid ${C.border}`, marginBottom: 28, overflowX: "auto" }}>
        {TABS.map(({ k, label, icon: Icon }) => (
          <button
            key={k}
            onClick={() => setTab(k)}
            style={{ display: "flex", alignItems: "center", gap: 6, padding: "12px 14px", fontFamily: FH, fontWeight: 700, fontSize: 13, color: tab === k ? C.teal : C.muted, borderBottom: `2px solid ${tab === k ? C.teal : "transparent"}`, whiteSpace: "nowrap", background: "none", borderLeft: "none", borderRight: "none", borderTop: "none", cursor: "pointer" }}
          >
            <Icon size={14} /> {label}
          </button>
        ))}
      </div>

      {tab === "saved" && (
        <div>
          {savedInstitutions.length === 0 ? (
            <div style={{ padding: 48, borderRadius: 16, border: `1px dashed ${C.border}`, textAlign: "center", color: C.muted }}>
              <Bookmark size={28} style={{ color: C.dim, margin: "0 auto 12px" }} />
              <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 16, color: C.text, marginBottom: 6 }}>{t({ru:"Пока пусто",tg:"Ҳоло холӣ аст"})}</p>
              <p style={{ fontSize: 13.5, marginBottom: 14 }}>{t({ru:"Нажмите на ♡ на карточке учреждения, чтобы сохранить его сюда",tg:"Дар корти муассиса ♡-ро пахш кунед, то онро дар ин ҷо нигоҳ доред"})}</p>
              <button onClick={() => router.push("/search")} style={{ padding: "9px 20px", borderRadius: 10, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 700, fontSize: 13.5, border: "none", cursor: "pointer" }}>
                {t({ru:"Перейти к поиску",tg:"Ба ҷустуҷӯ гузаред"})}
              </button>
            </div>
          ) : (
            <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill,minmax(260px,1fr))", gap: 14 }}>
              {savedInstitutions.map((inst) => (
                <div key={inst.id} style={{ borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, overflow: "hidden" }}>
                  <div style={{ height: 110, position: "relative" }}>
                    <img src={inst.coverPhoto} alt={t(inst.name)} style={{ width: "100%", height: "100%", objectFit: "cover" }} />
                    <button
                      onClick={() => toggleSaved(inst.id)}
                      aria-label={t({ru:"Убрать из избранного",tg:"Аз интихобшуда бардоштан"})}
                      style={{ position: "absolute", top: 8, right: 8, width: 28, height: 28, borderRadius: "50%", background: "rgba(0,0,0,.5)", border: "none", display: "flex", alignItems: "center", justifyContent: "center", cursor: "pointer", color: "#fff" }}
                    >
                      <X size={13} />
                    </button>
                  </div>
                  <div style={{ padding: "12px 14px" }}>
                    <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text, marginBottom: 8 }}>{t(inst.name)}</p>
                    <button
                      onClick={() => router.push(`/institutions/${inst.id}`)}
                      style={{ width: "100%", padding: "8px", borderRadius: 9, background: C.s3, color: C.text, fontFamily: FH, fontWeight: 700, fontSize: 12.5, border: "none", cursor: "pointer" }}
                    >
                      {t("common.open")}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {tab === "reviews" && (
        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          {myReviews.length === 0 && <p style={{ color: C.muted, fontSize: 14 }}>{t({ru:"Вы ещё не оставляли отзывов",tg:"Шумо ҳанӯз шарҳ нагузоштаед"})}</p>}
          {myReviews.map((r) => {
            const inst = INSTITUTIONS.find((i) => i.id === r.instId);
            return (
              <button key={r.id} onClick={() => inst && router.push(`/institutions/${inst.id}?tab=reviews&review=${r.id}`)}
                style={{ display: "block", width: "100%", textAlign: "left", cursor: inst ? "pointer" : "default", borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: "16px 18px", font: "inherit" }}>
                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 8 }}>
                  <div>
                    <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text }}>{inst ? t(inst.name) : t({ru:"Учреждение",tg:"Муассиса"})}</p>
                    <p style={{ fontSize: 12, color: C.sub }}>{r.date}</p>
                  </div>
                  <span style={{ color: C.gold, fontFamily: FH, fontWeight: 700, fontSize: 13 }}>{r.score} ★</span>
                </div>
                <p style={{ fontSize: 13.5, color: C.sub, lineHeight: 1.6 }}>{t(r.text)}</p>
              </button>
            );
          })}
        </div>
      )}

      {tab === "messages" && (
        <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 32, textAlign: "center" }}>
          <MessageSquare size={30} style={{ color: C.teal, margin: "0 auto 14px" }} />
          <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 17, color: C.text, marginBottom: 8 }}>{t({ru:"Все переписки — в разделе «Сообщения»",tg:"Ҳама мукотиба — дар бахши «Паёмҳо»"})}</p>
          <p style={{ fontSize: 13.5, color: C.sub, marginBottom: 18, maxWidth: 380, margin: "0 auto 18px" }}>
            {t({ru:"Общайтесь с учреждениями напрямую — все треды собраны в одном месте.",tg:"Бо муассисаҳо бевосита муошират кунед — ҳама мукотиба дар як ҷо ҷамъ аст."})}
          </p>
          <button onClick={() => router.push(chatHref())} style={{ padding: "10px 22px", borderRadius: 12, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer" }}>
            {t({ru:"Открыть сообщения",tg:"Паёмҳоро кушодан"})}
          </button>
        </div>
      )}

      {tab === "children" && (
        <div>
          <p style={{ fontSize: 13, color: C.sub, marginBottom: 16, lineHeight: 1.6 }}>
            {t({ ru: "Привязка ребёнка к учреждению нужна, чтобы вы могли оставлять верифицированные отзывы — только у родителей с реальной связью с учреждением.", tg: "Алоқаи фарзанд бо муассиса лозим аст, то шумо шарҳҳои тасдиқшуда гузошта тавонед — танҳо волидайне, ки бо муассиса алоқаи воқеӣ доранд." })}
          </p>
          <div style={{ display: "flex", flexDirection: "column", gap: 10, marginBottom: 20 }}>
            {children.map((c) => {
              const inst = INSTITUTIONS.find((i) => i.id === c.instId);
              return (
                <div key={c.id} style={{ display: "flex", alignItems: "center", justifyContent: "space-between", borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "14px 18px" }}>
                  <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
                    <div style={{ width: 36, height: 36, borderRadius: "50%", background: `${C.teal}22`, display: "flex", alignItems: "center", justifyContent: "center" }}>
                      <Baby size={16} style={{ color: C.teal }} />
                    </div>
                    <div>
                      <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text }}>{c.name}</p>
                      <p style={{ fontSize: 12.5, color: C.sub }}>
                        {c.age}
                        {inst ? ` · ${t(inst.name)}` : ` · ${t("children.noInstitution")}`}
                        {" · "}{t(CHILD_STATUS_KEY[c.status])}
                      </p>
                    </div>
                  </div>
                  <button onClick={() => removeChild(c.id)} style={{ background: "none", border: "none", color: C.muted, cursor: "pointer" }}>
                    <Trash2 size={15} />
                  </button>
                </div>
              );
            })}
          </div>
          <div style={{ borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: 20, display: "flex", gap: 10, flexWrap: "wrap", position: "relative" }}>
            <input value={childName} onChange={(e) => setChildName(e.target.value)} placeholder={t({ru:"Имя ребёнка",tg:"Номи фарзанд"})} style={{ flex: 1, minWidth: 140, padding: "10px 14px", borderRadius: 10, border: `1px solid ${C.border}`, background: C.s2, color: C.text, fontFamily: FB, fontSize: 14, outline: "none" }} />
            <input value={childAge} onChange={(e) => setChildAge(e.target.value)} placeholder={t({ru:"Возраст",tg:"Синну сол"})} style={{ width: 100, padding: "10px 14px", borderRadius: 10, border: `1px solid ${C.border}`, background: C.s2, color: C.text, fontFamily: FB, fontSize: 14, outline: "none" }} />

            <div ref={childInstBoxRef} style={{ position: "relative", minWidth: 200 }}>
              <input
                value={childInstQuery}
                onChange={(e) => { setChildInstQuery(e.target.value); setChildInstId(""); setChildInstOpen(true); }}
                onFocus={() => setChildInstOpen(true)}
                placeholder={t("children.institution")}
                style={{ width: "100%", padding: "10px 14px", borderRadius: 10, border: `1px solid ${C.border}`, background: C.s2, color: C.text, fontFamily: FB, fontSize: 14, outline: "none", boxSizing: "border-box" }}
              />
              {childInstOpen && childInstMatches.length > 0 && (
                <div style={{ position: "absolute", top: "calc(100% + 4px)", left: 0, right: 0, background: C.s2, border: `1px solid ${C.border}`, borderRadius: 10, boxShadow: "0 12px 30px rgba(0,0,0,.35)", zIndex: 30, maxHeight: 220, overflowY: "auto" }}>
                  {childInstMatches.map((i) => (
                    <button key={i.id} type="button" onClick={() => pickChildInst(i)} style={{ display: "block", width: "100%", textAlign: "left", padding: "9px 14px", background: "none", border: "none", color: C.text, fontFamily: FB, fontSize: 13.5, cursor: "pointer" }}>
                      {t(i.name)}
                    </button>
                  ))}
                </div>
              )}
            </div>

            <select value={childStatus} onChange={(e) => setChildStatus(e.target.value as ChildStatus)} style={{ padding: "10px 14px", borderRadius: 10, border: `1px solid ${C.border}`, background: C.s2, color: C.text, fontFamily: FB, fontSize: 14, outline: "none" }}>
              <option value="current">{t("children.statusCurrent")}</option>
              <option value="alumnus">{t("children.statusAlumnus")}</option>
              <option value="transferred">{t("children.statusTransferred")}</option>
            </select>
            <button onClick={submitChild} disabled={!childName.trim() || childInstId === ""} style={{ display: "flex", alignItems: "center", gap: 6, padding: "10px 18px", borderRadius: 10, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 700, fontSize: 13.5, border: "none", cursor: "pointer", opacity: (!childName.trim() || childInstId === "") ? 0.5 : 1 }}>
              <Plus size={14} /> {t("common.add")}
            </button>
          </div>
          <p style={{ fontSize: 12, color: C.dim, marginTop: 10, lineHeight: 1.5 }}>{t("children.multiInstHint")}</p>
        </div>
      )}

      {tab === "applicant" && (
        <div>
          {/* видимость — разовая настройка, не операционный процесс (SRS-совместимо) */}
          <div style={{ borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: "16px 20px", marginBottom: 20 }}>
            <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 14 }}>
              <Sparkles size={18} style={{ color: applicant.visibility === "public" ? C.teal : C.muted }} />
              <p style={{ fontSize: 13.5, color: C.sub }}>
                {t(`applicant.visibility.${applicant.visibility}Desc`)}
              </p>
            </div>
            <div style={{ display: "flex", gap: 8, flexWrap: "wrap", marginBottom: 14 }}>
              {(["draft", "on_response", "public"] as ApplicantVisibility[]).map((v) => (
                <button key={v} onClick={() => setVisibility(v)} style={{
                  padding: "9px 16px", borderRadius: 10, border: `1px solid ${applicant.visibility === v ? C.teal : C.border}`,
                  background: applicant.visibility === v ? C.teal : "transparent", color: applicant.visibility === v ? C.overlay : C.sub,
                  fontFamily: FH, fontWeight: 700, fontSize: 13, cursor: "pointer",
                }}>
                  {t(`applicant.visibility.${v}`)}
                </button>
              ))}
            </div>
            <label style={{ display: "flex", alignItems: "center", gap: 8, fontSize: 12.5, color: C.sub, cursor: "pointer" }}>
              <input type="checkbox" checked={applicant.hideContacts} onChange={(e) => setApplicant({ ...applicant, hideContacts: e.target.checked })} />
              {t("applicant.hideContacts")}
            </label>
          </div>

          {/* внутренние подвкладки профиля соискателя */}
          <div style={{ display: "flex", gap: 2, borderBottom: `1px solid ${C.border}`, marginBottom: 20, overflowX: "auto" }}>
            {([
              { k: "profile" as ApplicantSubTab, label: t({ ru: "Резюме", tg: "Резюме" }), icon: Briefcase },
              { k: "achievements" as ApplicantSubTab, label: t("tab.achievements"), icon: Award },
              { k: "applications" as ApplicantSubTab, label: t({ru:"Отклики",tg:"Ҷавобҳо"}), icon: MessageSquare },
            ]).map(({ k, label, icon: Icon }) => (
              <button key={k} onClick={() => setApplicantSubTab(k)}
                style={{ display: "flex", alignItems: "center", gap: 6, padding: "10px 14px", fontFamily: FH, fontWeight: 700, fontSize: 13, color: applicantSubTab === k ? C.teal : C.muted, borderBottom: `2px solid ${applicantSubTab === k ? C.teal : "transparent"}`, whiteSpace: "nowrap", background: "none", borderLeft: "none", borderRight: "none", borderTop: "none", cursor: "pointer" }}>
                <Icon size={14} /> {label}
              </button>
            ))}
          </div>

          {applicantSubTab === "profile" && (
            <div style={{ display: "grid", gridTemplateColumns: "260px 1fr", gap: 24, alignItems: "start" }}>
              <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 18 }}>
                <div onClick={() => photoInputRef.current?.click()} style={{ position: "relative", width: "100%", aspectRatio: "1", borderRadius: 14, marginBottom: 16, cursor: "pointer", overflow: "hidden" }}>
                  <img src={applicant.photo} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />
                  <div style={{ position: "absolute", inset: 0, background: "rgba(0,0,0,.45)", opacity: 0, transition: "opacity .15s", display: "flex", alignItems: "center", justifyContent: "center" }}
                    onMouseEnter={(e) => (e.currentTarget.style.opacity = "1")} onMouseLeave={(e) => (e.currentTarget.style.opacity = "0")}>
                    <Upload size={20} color="#fff" />
                  </div>
                </div>
                <input ref={photoInputRef} type="file" accept="image/*" onChange={onPhotoSelected} style={{ display: "none" }} />
                <FormField label={t({ ru: "Имя", tg: "Ном" })}>
                  <input value={applicant.name[locale]} onChange={(e) => setApplicant({ ...applicant, name: { ...applicant.name, [locale]: e.target.value } })} style={inputStyle} />
                </FormField>
                <FormField label={t("applicant.position")}>
                  <input value={applicant.position[locale]} onChange={(e) => setApplicant({ ...applicant, position: { ...applicant.position, [locale]: e.target.value } })} style={inputStyle} />
                </FormField>
                <FormField label={t({ ru: "Телефон", tg: "Телефон" })}>
                  <input value={applicant.phone ?? ""} onChange={(e) => setApplicant({ ...applicant, phone: e.target.value })} style={inputStyle} />
                </FormField>
                <FormField label={t({ ru: "Эл. почта", tg: "Почтаи электронӣ" })}>
                  <input value={applicant.email ?? ""} onChange={(e) => setApplicant({ ...applicant, email: e.target.value })} style={inputStyle} />
                </FormField>
              </div>

              <div style={{ display: "flex", flexDirection: "column", gap: 20 }}>
                <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 22 }}>
                  <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 16, color: C.text, marginBottom: 12, display: "flex", alignItems: "center", gap: 8 }}>
                    <FileText size={17} style={{ color: C.teal }} /> {t("applicant.cv")}
                  </h2>
                  {applicant.cvFileName ? (
                    <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 10, padding: "10px 14px", borderRadius: 10, background: C.s2 }}>
                      <span style={{ fontSize: 13.5, color: C.text, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{applicant.cvFileName}</span>
                      <button onClick={() => setApplicant({ ...applicant, cvFileName: undefined })} style={{ background: "none", border: "none", color: C.muted, cursor: "pointer", flexShrink: 0 }}><Trash2 size={14} /></button>
                    </div>
                  ) : (
                    <button onClick={() => cvInputRef.current?.click()} style={{ display: "flex", alignItems: "center", gap: 8, padding: "10px 16px", borderRadius: 10, background: C.s2, color: C.sub, fontFamily: FH, fontWeight: 700, fontSize: 13, border: `1px dashed ${C.border}`, cursor: "pointer", width: "100%", justifyContent: "center" }}>
                      <Upload size={14} /> {t({ ru: "Прикрепить файл (PDF, DOC)", tg: "Файл замима кардан (PDF, DOC)" })}
                    </button>
                  )}
                  <input ref={cvInputRef} type="file" accept=".pdf,.doc,.docx" onChange={onCvSelected} style={{ display: "none" }} />
                  <p style={{ fontSize: 11.5, color: C.dim, marginTop: 8, lineHeight: 1.5 }}>
                    {t({ ru: "Прототип: файл не загружается на сервер, сохраняется только его название.", tg: "Прототип: файл ба сервер бор намешавад, танҳо номи он нигоҳ дошта мешавад." })}
                  </p>
                </div>

                <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 22 }}>
                  <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 16, color: C.text, marginBottom: 12 }}>{t("applicant.about")}</h2>
                  <textarea value={applicant.bio[locale]} onChange={(e) => setApplicant({ ...applicant, bio: { ...applicant.bio, [locale]: e.target.value } })} rows={4} style={{ ...inputStyle, resize: "vertical" }} />
                </div>

                <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 22 }}>
                  <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 16, color: C.text, marginBottom: 12, display: "flex", alignItems: "center", gap: 8 }}>
                    <GraduationCap size={17} style={{ color: C.teal }} /> {t("applicant.education")}
                  </h2>
                  <ListEditor items={applicant.education} onChange={(v) => setApplicant({ ...applicant, education: v })} placeholder={t({ ru: "Учебное заведение, специальность, год", tg: "Муассисаи таълимӣ, ихтисос, сол" })} />
                </div>

                <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 22 }}>
                  <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 16, color: C.text, marginBottom: 12, display: "flex", alignItems: "center", gap: 8 }}>
                    <Briefcase size={17} style={{ color: C.teal }} /> {t("applicant.experience")}
                  </h2>
                  <ListEditor items={applicant.experience} onChange={(v) => setApplicant({ ...applicant, experience: v })} placeholder={t({ ru: "Годы, должность, место работы", tg: "Солҳо, мансаб, ҷои кор" })} />
                </div>

                <div style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 22 }}>
                  <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 16, color: C.text, marginBottom: 12 }}>{t("applicant.skills")}</h2>
                  <ListEditor items={applicant.skills} onChange={(v) => setApplicant({ ...applicant, skills: v })} placeholder={t({ ru: "Навык", tg: "Малака" })} />
                </div>
              </div>
            </div>
          )}

          {applicantSubTab === "achievements" && (
            <div>
              <div style={{ display: "flex", justifyContent: "flex-end", marginBottom: 16 }}>
                <button onClick={newAchievement} style={{ display: "flex", alignItems: "center", gap: 6, padding: "9px 16px", borderRadius: 10, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 700, fontSize: 13, border: "none", cursor: "pointer" }}>
                  <Plus size={14} /> {t("common.add")}
                </button>
              </div>
              {applicant.achievements.length === 0 ? (
                <div style={{ padding: 48, borderRadius: 16, border: `1px dashed ${C.border}`, textAlign: "center", color: C.muted }}>
                  <Award size={28} style={{ color: C.dim, margin: "0 auto 12px" }} />
                  <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 16, color: C.text }}>{t("empty.achievements")}</p>
                </div>
              ) : (
                <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill,minmax(280px,1fr))", gap: 12 }}>
                  {applicant.achievements.map((a) => {
                    const tier = ACH_TIER[a.category] ?? ACH_TIER.special;
                    const TierIcon = tier.icon;
                    return (
                      <div key={a.id} style={{ display: "flex", alignItems: "center", gap: 12, padding: "14px 16px", borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1 }}>
                        <TierIcon size={18} style={{ color: tier.color, flexShrink: 0 }} />
                        <div style={{ flex: 1, minWidth: 0 }}>
                          <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 13.5, color: C.text }}>{a.title[locale]}</p>
                          <p style={{ fontSize: 12, color: C.sub }}>{a.year} · {a.desc[locale]}</p>
                        </div>
                        <button onClick={() => setAchievementForm(a)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 7, cursor: "pointer", color: C.sub }}><Pencil size={13} /></button>
                        <button onClick={() => deleteAchievement(a.id)} style={{ background: C.s3, border: "none", borderRadius: 8, padding: 7, cursor: "pointer", color: C.red }}><Trash2 size={13} /></button>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          )}

          {applicantSubTab === "applications" && (
            <div style={{ display: "flex", flexDirection: "column", gap: 28 }}>
              <div>
                <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 15, color: C.text, marginBottom: 12 }}>{t("applicant.myApplications")}</p>
                {applications.length === 0 ? (
                  <p style={{ color: C.muted, fontSize: 13.5 }}>{t("empty.applications")}</p>
                ) : (
                  <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                    {applications.map((a) => {
                      const vacancy = VACANCIES.find((v) => v.id === a.vacancyId);
                      const inst = vacancy ? INSTITUTIONS.find((i) => i.id === vacancy.instId) : null;
                      return (
                        <button key={a.id} onClick={() => vacancy && inst && router.push(`/institutions/${inst.id}?tab=vacancies&vacancy=${vacancy.id}`)} style={{ display: "flex", alignItems: "center", justifyContent: "space-between", width: "100%", textAlign: "left", borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, padding: "14px 18px", cursor: vacancy && inst ? "pointer" : "default", font: "inherit" }}>
                          <div>
                            <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text }}>{vacancy ? vacancy.title[locale] : a.vacancyId}</p>
                            <p style={{ fontSize: 12.5, color: C.sub, marginTop: 2 }}>{inst ? inst.name[locale] : ""} · {a.createdAt}</p>
                          </div>
                          <span style={{ fontSize: 12, fontWeight: 700, color: a.status === "closed" ? C.muted : C.teal, fontFamily: FH, flexShrink: 0 }}>
                            {t(`application.status.${a.status}`)}
                          </span>
                        </button>
                      );
                    })}
                  </div>
                )}
              </div>

              <div>
                <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 15, color: C.text, marginBottom: 12 }}>{t("applicant.employerResponses")}</p>
                {myResponses.length === 0 ? (
                  <p style={{ color: C.muted, fontSize: 13.5 }}>{t("empty.employerResponses")}</p>
                ) : (
                  <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                    {myResponses.map((r) => {
                      const inst = INSTITUTIONS.find((i) => i.id === r.instId);
                      if (!inst) return null;
                      return (
                        <button key={r.id} onClick={() => router.push(`/institutions/${inst.id}`)} style={{ display: "flex", alignItems: "center", gap: 14, width: "100%", textAlign: "left", borderRadius: 14, border: `1px solid ${C.teal}33`, background: `${C.teal}0a`, padding: "14px 18px", cursor: "pointer", font: "inherit" }}>
                          <img src={inst.coverPhoto} alt="" style={{ width: 44, height: 44, borderRadius: 10, objectFit: "cover", flexShrink: 0 }} />
                          <div style={{ flex: 1, minWidth: 0 }}>
                            <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text }}>{t(inst.name)}</p>
                            <p style={{ fontSize: 12.5, color: C.sub, marginTop: 2, overflow: "hidden", textOverflow: "ellipsis", display: "-webkit-box", WebkitLineClamp: 2, WebkitBoxOrient: "vertical" as const }}>{t(r.message)}</p>
                            <p style={{ fontSize: 11.5, color: C.dim, marginTop: 4 }}>{r.date}</p>
                          </div>
                          <ChevronRight size={16} style={{ color: C.teal, flexShrink: 0 }} />
                        </button>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      )}

      <Modal open={!!achievementForm} onClose={() => setAchievementForm(null)} maxWidth={480}>
        {achievementForm && (
          <div style={{ padding: 24 }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 18 }}>
              <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 18, color: C.text }}>
                {applicant.achievements.some((a) => a.id === achievementForm.id) ? t("common.edit") : t({ ru: "Новое достижение", tg: "Дастоварди нав" })}
              </h3>
              <button onClick={() => setAchievementForm(null)} style={{ background: "none", border: "none", color: C.sub, cursor: "pointer" }}><X size={18} /></button>
            </div>
            <FormField label={`${t({ ru: "Название", tg: "Ном" })} (${locale.toUpperCase()})`}>
              <input value={achievementForm.title[locale]} onChange={(e) => setAchievementForm({ ...achievementForm, title: { ...achievementForm.title, [locale]: e.target.value } })} style={inputStyle} />
            </FormField>
            <FormField label={`${t({ ru: "Описание", tg: "Тавсиф" })} (${locale.toUpperCase()})`}>
              <textarea value={achievementForm.desc[locale]} onChange={(e) => setAchievementForm({ ...achievementForm, desc: { ...achievementForm.desc, [locale]: e.target.value } })} style={{ ...inputStyle, height: 80, resize: "none" as const }} />
            </FormField>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12 }}>
              <FormField label={t({ ru: "Год", tg: "Сол" })}>
                <input type="number" value={achievementForm.year} onChange={(e) => setAchievementForm({ ...achievementForm, year: Number(e.target.value) })} style={inputStyle} />
              </FormField>
              <FormField label={t({ ru: "Уровень", tg: "Дараҷа" })}>
                <select value={achievementForm.category} onChange={(e) => setAchievementForm({ ...achievementForm, category: e.target.value as Achievement["category"] })} style={inputStyle}>
                  <option value="gold">{t({ ru: "Золото", tg: "Тилло" })}</option>
                  <option value="silver">{t({ ru: "Серебро", tg: "Нуқра" })}</option>
                  <option value="bronze">{t({ ru: "Бронза", tg: "Биринҷӣ" })}</option>
                  <option value="special">{t({ ru: "Особое", tg: "Махсус" })}</option>
                </select>
              </FormField>
            </div>
            <button onClick={saveAchievement} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: "pointer", marginTop: 6 }}>{t("common.save")}</button>
          </div>
        )}
      </Modal>

      <Toast message={toast} onDone={() => setToast(null)} />
    </div>
  );
}
