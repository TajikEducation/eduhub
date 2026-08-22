"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { Briefcase, Wallet, Clock, ChevronRight, Search, X } from "lucide-react";
import { C, FH, FB, PHOTOS, VACANCIES, INSTITUTIONS, CATEGORY_META, REGION_LABEL, REGION_ORDER, type Vacancy, type Institution, type CategoryKey, type Region } from "@/lib/data";
import { SubjectMotifs } from "@/components/SubjectMotifs";
import { useReveal, revealStyle } from "@/lib/useReveal";
import { useT } from "@/lib/i18n";

const CATEGORY_KEYS: CategoryKey[] = ["cat_kg", "cat_school", "cat_center", "cat_uni"];

function Chip({ l, on, click }: { l: string; on: boolean; click: () => void }) {
  return (
    <button onClick={click} style={{ fontSize: 12.5, fontWeight: 600, padding: "6px 12px", borderRadius: 8, border: `1px solid ${on ? C.teal : C.border}`, background: on ? `${C.teal}22` : "transparent", color: on ? C.teal : C.sub, fontFamily: FH, whiteSpace: "nowrap", cursor: "pointer" }}>
      {l}
    </button>
  );
}

function VacancyRow({ v, inst, style }: { v: Vacancy; inst: Institution | undefined; style: React.CSSProperties }) {
  const t = useT();
  const [hov, setHov] = useState(false);
  const meta = inst ? CATEGORY_META[inst.tk] : null;
  const Icon = meta?.icon;
  const accent = inst?.color ?? C.teal;
  return (
    <Link href={inst ? `/institutions/${inst.id}?tab=vacancies&vacancy=${v.id}` : "#"} onMouseEnter={() => setHov(true)} onMouseLeave={() => setHov(false)}
      style={{ ...style, display: "flex", alignItems: "center", gap: 16, borderRadius: 16, border: `1px solid ${hov ? accent + "55" : C.border}`, background: C.s1, padding: "14px 20px 14px 14px", textDecoration: "none", transition: "all .22s", transform: hov ? "translateY(-3px)" : "none", boxShadow: hov ? "0 16px 40px rgba(0,0,0,.4)" : undefined }}>
      <div style={{ width: 64, height: 64, borderRadius: 12, overflow: "hidden", position: "relative", flexShrink: 0, background: accent }}>
        {inst?.coverPhoto && <img src={inst.coverPhoto} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />}
        {Icon && (
          <div style={{ position: "absolute", bottom: 4, left: 4, width: 22, height: 22, borderRadius: 6, background: "rgba(0,0,0,.5)", backdropFilter: "blur(6px)", display: "flex", alignItems: "center", justifyContent: "center" }}>
            <Icon size={12} color="#fff" />
          </div>
        )}
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 15, color: C.text }}>{t(v.title)}</p>
        <p style={{ fontSize: 13, color: C.sub, marginTop: 2 }}>{inst ? t(inst.name) : ""}{inst ? ` · ${inst.area}` : ""}</p>
        <div style={{ display: "flex", gap: 14, marginTop: 8, flexWrap: "wrap" }}>
          {v.salaryFrom && (
            <span style={{ display: "flex", alignItems: "center", gap: 5, fontSize: 12.5, color: C.sub }}>
              <Wallet size={12}/> {v.salaryFrom}–{v.salaryTo} {t("common.perMonth")}
            </span>
          )}
          <span style={{ display: "flex", alignItems: "center", gap: 5, fontSize: 12.5, color: C.sub }}>
            <Clock size={12}/> {t(v.employment)}
          </span>
        </div>
      </div>
      <ChevronRight size={18} style={{ color: hov ? accent : C.dim, flexShrink: 0, transition: "color .15s" }}/>
    </Link>
  );
}

export default function VacanciesPage() {
  const t = useT();
  const { ref, visible } = useReveal<HTMLDivElement>({ threshold: 0 });
  const [q, setQ] = useState("");
  const [regionF, setRegionF] = useState<Region | null>(null);
  const [typeF, setTypeF] = useState<CategoryKey | null>(null);
  const [fullTimeOnly, setFullTimeOnly] = useState(false);

  const withInst = useMemo(
    () => VACANCIES.filter(v => v.status === "published").map(v => ({ v, inst: INSTITUTIONS.find(x => x.id === v.instId) })),
    []
  );

  const vacancies = useMemo(() => withInst.filter(({ v, inst }) => {
    if (q) {
      const needle = q.toLowerCase();
      const hay = [t(v.title), inst ? t(inst.name) : ""].join(" ").toLowerCase();
      if (!hay.includes(needle)) return false;
    }
    if (regionF && inst?.region !== regionF) return false;
    if (typeF && inst?.tk !== typeF) return false;
    if (fullTimeOnly && v.employment.ru !== "Полная занятость") return false;
    return true;
  }).map(({ v }) => v), [withInst, q, regionF, typeF, fullTimeOnly, t]);

  const activeCount = [regionF !== null, typeF !== null, fullTimeOnly].filter(Boolean).length;
  const clearAll = () => { setRegionF(null); setTypeF(null); setFullTimeOnly(false); };

  return (
    <div>
      {/* ── HERO ── */}
      <div style={{ position: "relative", height: 260, overflow: "hidden" }}>
        <img src={PHOTOS.heroGuideA} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />
        <div style={{ position: "absolute", inset: 0, background: `linear-gradient(180deg, ${C.overlay}55 0%, ${C.overlay}D8 75%, ${C.overlay} 100%)` }} />
        <SubjectMotifs/>
        <div style={{ position: "absolute", inset: 0, display: "flex", alignItems: "flex-end" }}>
          <div style={{ maxWidth: 1260, margin: "0 auto", padding: "0 28px 28px", width: "100%" }}>
            <div style={{ width: 44, height: 44, borderRadius: 12, background: `${C.teal}22`, border: `1px solid ${C.teal}44`, display: "flex", alignItems: "center", justifyContent: "center" }}>
              <Briefcase size={20} style={{ color: C.teal }}/>
            </div>
            <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(24px,3.4vw,38px)", color: "#fff", margin: "10px 0 6px", letterSpacing: "-.02em" }}>
              {t({ ru: "Вакансии в образовательных учреждениях", tg: "Ҷойҳои холӣ дар муассисаҳои таълимӣ" })}
            </h1>
            <p style={{ fontSize: 14.5, color: "rgba(255,255,255,.72)", maxWidth: 560 }}>
              {t({ ru: "Открытые вакансии от школ, детских садов, учебных центров и вузов Таджикистана", tg: "Ҷойҳои холии кушода аз мактабҳо, боғчаҳо, марказҳои таълимӣ ва донишгоҳҳои Тоҷикистон" })}
            </p>
          </div>
        </div>
      </div>

      {/* ── LIST ── */}
      <div style={{ maxWidth: 1260, margin: "0 auto", padding: "28px 28px 80px" }}>
        <div style={{ display: "flex", gap: 8, marginBottom: 20 }}>
          <span style={{ padding: "8px 16px", borderRadius: 10, background: `${C.teal}18`, border: `1px solid ${C.teal}44`, color: C.teal, fontFamily: FH, fontWeight: 700, fontSize: 13 }}>
            {t("nav.vacancies")}
          </span>
          <Link href="/applicants" style={{ padding: "8px 16px", borderRadius: 10, background: C.s3, border: `1px solid ${C.border}`, color: C.sub, fontFamily: FH, fontWeight: 700, fontSize: 13, textDecoration: "none" }}>
            {t("tab.candidates")}
          </Link>
        </div>

        {/* ── FILTERS ── */}
        <div style={{ borderRadius: 16, border: `1px solid ${C.border}`, background: C.s1, padding: 16, marginBottom: 20, display: "flex", flexDirection: "column", gap: 12 }}>
          <div style={{ display: "flex", alignItems: "center", gap: 10, borderRadius: 10, border: `1px solid ${C.border}`, background: C.s2, padding: "9px 14px" }}>
            <Search size={15} style={{ color: C.dim, flexShrink: 0 }} />
            <input value={q} onChange={e => setQ(e.target.value)} placeholder={t({ ru: "Должность или учреждение…", tg: "Вазифа ё муассиса…" })}
              style={{ flex: 1, background: "transparent", border: "none", outline: "none", fontSize: 13.5, color: C.text, fontFamily: FB }} />
          </div>
          <div style={{ display: "flex", gap: 8, flexWrap: "wrap", alignItems: "center" }}>
            {CATEGORY_KEYS.map(k => (
              <Chip key={k} l={t(CATEGORY_META[k].label)} on={typeF === k} click={() => setTypeF(typeF === k ? null : k)} />
            ))}
            <span style={{ width: 1, height: 18, background: C.border, margin: "0 2px" }} />
            <select value={regionF ?? ""} onChange={e => setRegionF((e.target.value || null) as Region | null)}
              style={{ fontSize: 12.5, fontWeight: 600, padding: "6px 10px", borderRadius: 8, border: `1px solid ${C.border}`, background: "transparent", color: regionF ? C.teal : C.sub, fontFamily: FH, cursor: "pointer" }}>
              <option value="">{t("nav.allRegions")}</option>
              {REGION_ORDER.map(r => <option key={r} value={r}>{t(REGION_LABEL[r])}</option>)}
            </select>
            <Chip l={t({ ru: "Полная занятость", tg: "Шуғли пурра" })} on={fullTimeOnly} click={() => setFullTimeOnly(v => !v)} />
            {activeCount > 0 && (
              <button onClick={clearAll} style={{ display: "flex", alignItems: "center", gap: 4, fontSize: 12, color: C.dim, background: "none", border: "none", cursor: "pointer", fontFamily: FH, fontWeight: 600 }}>
                <X size={13} /> {t({ ru: "Сбросить", tg: "Тоза кардан" })}
              </button>
            )}
          </div>
        </div>

        {vacancies.length === 0 ? (
          <div style={{ padding: 56, borderRadius: 16, border: `1px dashed ${C.border}`, textAlign: "center", color: C.muted }}>
            <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 17, color: C.text }}>{t("empty.vacancies")}</p>
          </div>
        ) : (
          <div ref={ref} style={{ display: "flex", flexDirection: "column", gap: 12 }}>
            {vacancies.map((v, i) => {
              const inst = INSTITUTIONS.find(x => x.id === v.instId);
              return <VacancyRow key={v.id} v={v} inst={inst} style={revealStyle(visible, Math.min(i, 8) * 45)} />;
            })}
          </div>
        )}
      </div>
    </div>
  );
}
