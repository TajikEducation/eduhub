"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { MapPin } from "lucide-react";
import { C, FH, FB, CATEGORY_META, REGION_LABEL, REGION_ORDER, type Region } from "@/lib/data";
import { RegionMap, type Bbox } from "@/components/InstMap";
import { useVisibleInstitutions } from "@/lib/app-state";
import { useT } from "@/lib/i18n";

const REGION_BBOX: Record<Region, Bbox> = {
  dushanbe: [68.70, 38.50, 68.87, 38.62],
  sughd: [69.45, 40.15, 69.80, 40.42],
  khatlon: [68.55, 37.70, 69.05, 38.00],
  gbao: [70.80, 37.10, 74.85, 38.95],
  rrp: [68.60, 38.30, 70.20, 39.30],
};
const ALL_BBOX: Bbox = [67.3, 36.6, 75.2, 41.05];

export default function MapPage() {
  const router = useRouter();
  const t = useT();
  const [regionFilter, setRegionFilter] = useState<Region | null>(null);
  const institutions = useVisibleInstitutions();

  const list = useMemo(
    () => institutions.filter((i) => i.geo && (!regionFilter || i.region === regionFilter)),
    [institutions, regionFilter]
  );
  const bbox = regionFilter ? REGION_BBOX[regionFilter] : ALL_BBOX;
  const points = list.map((i) => ({
    id: i.id, lat: i.geo!.lat, lng: i.geo!.lng,
    label: t(i.name), color: CATEGORY_META[i.tk].color, href: `/institutions/${i.id}`,
  }));

  return (
    <div style={{ maxWidth: 1260, margin: "0 auto", padding: "28px 28px 80px" }}>
      <div style={{ marginBottom: 20 }}>
        <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(22px,3vw,30px)", color: C.text, letterSpacing: "-.02em" }}>
          {t({ ru: "Карта учреждений", tg: "Харитаи муассисаҳо" })}
        </h1>
        <p style={{ fontSize: 14, color: C.sub, marginTop: 4 }}>
          {t({ ru: "Все зарегистрированные учреждения по регионам Таджикистана", tg: "Ҳама муассисаҳои сабтшуда аз рӯи минтақаҳои Тоҷикистон" })}
        </p>
      </div>

      <div style={{ display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap", marginBottom: 20 }}>
        <span style={{ display: "flex", alignItems: "center", gap: 5, fontSize: 12.5, color: C.dim, fontFamily: FH, marginRight: 4 }}>
          <MapPin size={13} /> {t("geo.region")}:
        </span>
        <button
          onClick={() => setRegionFilter(null)}
          style={{ fontSize: 12.5, fontWeight: 600, padding: "6px 12px", borderRadius: 8, border: `1px solid ${!regionFilter ? C.teal : C.border}`, background: !regionFilter ? `${C.teal}22` : "transparent", color: !regionFilter ? C.teal : C.sub, fontFamily: FH, cursor: "pointer" }}
        >
          {t("nav.allRegions")}
        </button>
        {REGION_ORDER.map((r) => (
          <button
            key={r}
            onClick={() => setRegionFilter(r)}
            style={{ fontSize: 12.5, fontWeight: 600, padding: "6px 12px", borderRadius: 8, border: `1px solid ${regionFilter === r ? C.teal : C.border}`, background: regionFilter === r ? `${C.teal}22` : "transparent", color: regionFilter === r ? C.teal : C.sub, fontFamily: FH, cursor: "pointer" }}
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
          <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 17, color: C.text }}>{t("empty.notFound")}</p>
        </div>
      ) : (
        <RegionMap points={points} bbox={bbox} height={520} />
      )}

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill,minmax(240px,1fr))", gap: 10, marginTop: 20 }}>
        {list.map((i) => {
          const meta = CATEGORY_META[i.tk];
          return (
            <button key={i.id} onClick={() => router.push(`/institutions/${i.id}`)}
              style={{ display: "flex", alignItems: "center", gap: 10, padding: "10px 14px", borderRadius: 12, border: `1px solid ${C.border}`, background: C.s1, cursor: "pointer", textAlign: "left", font: "inherit" }}>
              <MapPin size={14} style={{ color: meta.color, flexShrink: 0 }} fill={meta.color} />
              <span style={{ fontSize: 13, color: C.text, fontFamily: FH, fontWeight: 600 }}>{t(i.name)}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
}
