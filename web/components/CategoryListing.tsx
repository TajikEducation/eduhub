"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { MapPin } from "lucide-react";
import { C, FH, FB, CATEGORY_META, REGION_LABEL, REGION_ORDER, type CategoryKey } from "@/lib/data";
import { InstitCard } from "./InstitCard";
import { SubjectMotifs } from "./SubjectMotifs";
import { useAppState, useVisibleInstitutions } from "@/lib/app-state";
import { useT } from "@/lib/i18n";
import { useReveal, revealStyle } from "@/lib/useReveal";

export function CategoryListing({ tk }: { tk: CategoryKey }) {
  const router = useRouter();
  const t = useT();
  const { region: globalRegion } = useAppState();
  const [regionFilter, setRegionFilter] = useState(globalRegion);
  const meta = CATEGORY_META[tk];
  const Icon = meta.icon;
  const { ref, visible } = useReveal<HTMLDivElement>();
  const institutions = useVisibleInstitutions();

  const list = useMemo(
    () => institutions.filter((i) => i.tk === tk && (!regionFilter || i.region === regionFilter)),
    [institutions, tk, regionFilter]
  );

  return (
    <div>
      {/* ── HERO ── */}
      <div style={{ position: "relative", height: 260, overflow: "hidden" }}>
        <img src={meta.heroPhoto} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />
        <div style={{ position: "absolute", inset: 0, background: `linear-gradient(180deg, ${C.overlay}55 0%, ${C.overlay}D8 75%, ${C.overlay} 100%)` }} />
        <SubjectMotifs/>
        <div style={{ position: "absolute", inset: 0, display: "flex", alignItems: "flex-end" }}>
          <div style={{ maxWidth: 1260, margin: "0 auto", padding: "0 28px 28px", width: "100%" }}>
            <div style={{ width: 44, height: 44, borderRadius: 12, background: `${meta.color}22`, border: `1px solid ${meta.color}44`, display: "flex", alignItems: "center", justifyContent: "center" }}>
              <Icon size={20} style={{ color: meta.color }}/>
            </div>
            <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(24px,3.4vw,38px)", color: "#fff", margin: "10px 0 6px", letterSpacing: "-.02em" }}>
              {t(meta.plural)}
            </h1>
            <p style={{ fontSize: 14.5, color: "rgba(255,255,255,.72)", maxWidth: 560 }}>{t(meta.desc)}</p>
          </div>
        </div>
      </div>

      {/* ── FILTER + GRID ── */}
      <div style={{ maxWidth: 1260, margin: "0 auto", padding: "28px 28px 80px" }}>
        <div style={{ display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap", marginBottom: 24 }}>
          <span style={{ display: "flex", alignItems: "center", gap: 5, fontSize: 12.5, color: C.dim, fontFamily: FH, marginRight: 4 }}>
            <MapPin size={13} /> {t("geo.region")}:
          </span>
          <button
            onClick={() => setRegionFilter(null)}
            style={{ fontSize:12.5,fontWeight:600,padding:"6px 12px",borderRadius:8,border:`1px solid ${!regionFilter?C.teal:C.border}`,background:!regionFilter?`${C.teal}22`:"transparent",color:!regionFilter?C.teal:C.sub,fontFamily:FH,cursor:"pointer" }}
          >
            {t("nav.allRegions")}
          </button>
          {REGION_ORDER.map((r) => (
            <button
              key={r}
              onClick={() => setRegionFilter(r)}
              style={{ fontSize:12.5,fontWeight:600,padding:"6px 12px",borderRadius:8,border:`1px solid ${regionFilter===r?C.teal:C.border}`,background:regionFilter===r?`${C.teal}22`:"transparent",color:regionFilter===r?C.teal:C.sub,fontFamily:FH,cursor:"pointer" }}
            >
              {t(REGION_LABEL[r])}
            </button>
          ))}
          <span style={{ marginLeft: "auto", fontFamily: FB, fontSize: 14, color: C.muted }}>
            <b style={{ color: C.text }}>{list.length}</b> {t({ ru: "учреждений", tg: "муассиса" })}
          </span>
        </div>

        {list.length === 0 ? (
          <div style={{ padding: 56, borderRadius: 16, border: `1px dashed ${C.border}`, textAlign: "center", color: C.muted }}>
            <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 17, color: C.text, marginBottom: 6 }}>{t("empty.notFound")}</p>
            <p style={{ fontSize: 13.5 }}>{t({ ru: "В этом регионе пока нет учреждений этой категории", tg: "Дар ин минтақа ҳанӯз муассисаи ин категория нест" })}</p>
          </div>
        ) : (
          <div ref={ref} style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill,minmax(260px,1fr))", gap: 18 }}>
            {list.map((inst, i) => (
              <div key={inst.id} style={revealStyle(visible, Math.min(i, 8) * 45)}>
                <InstitCard inst={inst} onClick={() => router.push(`/institutions/${inst.id}`)} />
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
