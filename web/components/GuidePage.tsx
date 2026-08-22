"use client";

import { useState } from "react";
import Link from "next/link";
import { ChevronDown, ArrowRight, type LucideIcon } from "lucide-react";
import { C, FH, FB, type Bi } from "@/lib/data";
import { useT, useLocale } from "@/lib/i18n";
import { SubjectMotifs } from "./SubjectMotifs";

export interface GuideStep { icon: LucideIcon; title: Bi; desc: Bi; screenshot?: string; }
export interface GuideFaq { q: Bi; a: Bi; group?: Bi; }

export function GuidePage({
  heroPhoto, eyebrow, title, subtitle, steps, faq, ctaLabel, ctaHref, secondaryLabel, secondaryHref, accent, extra,
}: {
  heroPhoto: string;
  eyebrow: Bi;
  title: Bi;
  subtitle: Bi;
  steps: GuideStep[];
  faq: GuideFaq[];
  ctaLabel: Bi;
  ctaHref: string;
  secondaryLabel: Bi;
  secondaryHref: string;
  accent: string;
  extra?: React.ReactNode;
}) {
  const t = useT();
  const { locale } = useLocale();
  const [openFaq, setOpenFaq] = useState<Set<number>>(new Set([0]));
  const toggleFaq = (i: number) => setOpenFaq((prev) => {
    const next = new Set(prev);
    next.has(i) ? next.delete(i) : next.add(i);
    return next;
  });

  const faqJsonLd = {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: faq.map((f) => ({
      "@type": "Question",
      name: f.q[locale],
      acceptedAnswer: { "@type": "Answer", text: f.a[locale] },
    })),
  };


  return (
    <div>
      {/* ── HERO ── */}
      <div style={{ position: "relative", overflow: "hidden", padding: "72px 28px 64px", textAlign: "center" }}>
        <img src={heroPhoto} alt="" style={{ position: "absolute", inset: 0, width: "100%", height: "100%", objectFit: "cover", opacity: 0.24 }} />
        <div style={{ position: "absolute", inset: 0, background: `linear-gradient(180deg, ${C.overlay}CC 0%, ${C.overlay}EE 55%, ${C.overlay} 100%)` }} />
        <SubjectMotifs/>
        <div style={{ position: "relative", maxWidth: 680, margin: "0 auto" }}>
          <span style={{ display: "inline-flex", padding: "6px 14px", borderRadius: 999, border: `1px solid ${accent}66`, background: `${accent}22`, fontSize: 12.5, fontWeight: 700, color: accent, fontFamily: FH, marginBottom: 18 }}>
            {t(eyebrow)}
          </span>
          <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(28px,4vw,44px)", color: "#fff", marginBottom: 14, letterSpacing: "-.02em", lineHeight: 1.15 }}>
            {t(title)}
          </h1>
          <p style={{ fontSize: 15, color: "rgba(255,255,255,.72)", lineHeight: 1.65, marginBottom: 30 }}>{t(subtitle)}</p>
          <div style={{ display: "flex", gap: 10, justifyContent: "center", flexWrap: "wrap" }}>
            <Link href={ctaHref} style={{ padding: "12px 24px", borderRadius: 12, background: accent, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, textDecoration: "none" }}>
              {t(ctaLabel)}
            </Link>
            <Link href={secondaryHref} style={{ display: "inline-flex", alignItems: "center", gap: 6, padding: "12px 6px", color: "rgba(255,255,255,.75)", fontFamily: FH, fontWeight: 700, fontSize: 14, textDecoration: "none" }}>
              {t(secondaryLabel)} <ArrowRight size={15}/>
            </Link>
          </div>
        </div>
      </div>

      {extra}

      {/* ── STEPS ── */}
      <div style={{ maxWidth: 1080, margin: "0 auto", padding: "0 28px 64px" }}>
        <div style={{ display: "flex", flexDirection: "column", gap: 56 }}>
          {steps.map((s, i) => {
            const Icon = s.icon;
            const reverse = i % 2 === 1;
            return (
              <div key={i} style={{ display: "grid", gridTemplateColumns: s.screenshot ? "1fr 1fr" : "1fr", gap: 40, alignItems: "center" }}>
                <div style={{ order: s.screenshot && reverse ? 2 : 1 }}>
                  <div style={{ display: "flex", alignItems: "center", gap: 10, marginBottom: 12 }}>
                    <div style={{ width: 36, height: 36, borderRadius: 10, background: `${accent}18`, display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
                      <Icon size={17} style={{ color: accent }} />
                    </div>
                    <span style={{ fontFamily: FH, fontWeight: 800, fontSize: 13, color: accent, letterSpacing: ".04em" }}>
                      {t({ ru: "ШАГ", tg: "ҚАДАМ" })} {i + 1}
                    </span>
                  </div>
                  <h3 style={{ fontFamily: FH, fontWeight: 800, fontSize: 21, color: C.text, marginBottom: 10, lineHeight: 1.25 }}>
                    {t(s.title)}
                  </h3>
                  <p style={{ fontSize: 14.5, color: C.sub, lineHeight: 1.7 }}>{t(s.desc)}</p>
                </div>
                {s.screenshot && (
                  <div style={{ order: reverse ? 1 : 2, borderRadius: 16, overflow: "hidden", border: `1px solid ${C.border}`, boxShadow: "0 20px 50px rgba(0,0,0,.35)" }}>
                    <img src={s.screenshot} alt={t(s.title)} style={{ width: "100%", display: "block" }} />
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* ── FAQ ── */}
      <div style={{ maxWidth: 780, margin: "0 auto", padding: "0 28px 90px" }}>
        <h2 style={{ fontFamily: FH, fontWeight: 800, fontSize: 22, color: C.text, marginBottom: 20, textAlign: "center" }}>
          {t({ ru: "Частые вопросы", tg: "Саволҳои маъмул" })}
        </h2>
        <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
          {faq.map((f, i) => {
            const isOpen = openFaq.has(i);
            const groupText = f.group ? t(f.group) : undefined;
            const prevGroupText = i > 0 && faq[i - 1].group ? t(faq[i - 1].group as Bi) : undefined;
            const showGroupHeader = groupText && groupText !== prevGroupText;
            return (
              <div key={i}>
                {showGroupHeader && (
                  <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 12, color: accent, textTransform: "uppercase", letterSpacing: ".05em", margin: i === 0 ? "0 0 10px" : "24px 0 10px" }}>
                    {groupText}
                  </p>
                )}
                <div style={{ borderRadius: 14, border: `1px solid ${C.border}`, background: C.s1, overflow: "hidden" }}>
                  <button
                    id={`faq-btn-${i}`}
                    onClick={() => toggleFaq(i)}
                    aria-expanded={isOpen}
                    aria-controls={`faq-panel-${i}`}
                    style={{ display: "flex", justifyContent: "space-between", alignItems: "center", width: "100%", padding: "16px 18px", background: "none", border: "none", cursor: "pointer", textAlign: "left" }}
                  >
                    <span style={{ fontFamily: FH, fontWeight: 700, fontSize: 14, color: C.text }}>{t(f.q)}</span>
                    <ChevronDown size={16} style={{ color: C.muted, transform: isOpen ? "rotate(180deg)" : "none", transition: "transform .15s", flexShrink: 0 }} />
                  </button>
                  {/* ответ остаётся в DOM всегда (виден краулерам/JSON-LD дублирует для FAQ rich-results), скрывается визуально через grid-строку (не layout-свойство) вместо max-height */}
                  <div
                    id={`faq-panel-${i}`}
                    role="region"
                    aria-labelledby={`faq-btn-${i}`}
                    style={{ display: "grid", gridTemplateRows: isOpen ? "1fr" : "0fr", opacity: isOpen ? 1 : 0, transition: "grid-template-rows .25s ease, opacity .2s ease" }}
                  >
                    <p style={{ minHeight: 0, overflow: "hidden", padding: "0 18px 16px", fontSize: 13.5, color: C.sub, lineHeight: 1.65, fontFamily: FB }}>
                      {t(f.a)}
                    </p>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(faqJsonLd) }} />
    </div>
  );
}
