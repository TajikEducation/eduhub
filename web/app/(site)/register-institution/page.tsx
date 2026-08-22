"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Building2 } from "lucide-react";
import { C, FH, FB, CATEGORY_META, REGION_ORDER, REGION_LABEL, type CategoryKey, type Region } from "@/lib/data";
import { useAppState, type RegisterInstitutionInput } from "@/lib/app-state";
import { useT } from "@/lib/i18n";

const CATEGORY_KEYS: CategoryKey[] = ["cat_kg", "cat_school", "cat_center", "cat_uni"];

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

const EMPTY: RegisterInstitutionInput = {
  name: "", tk: "cat_school", region: "dushanbe", city: "", street: "", area: "",
  price: 0, founded: new Date().getFullYear(), students: 0, age: "",
  phone: "", email: "", website: "", description: "",
};

export default function RegisterInstitutionPage() {
  const t = useT();
  const router = useRouter();
  const { registerInstitution } = useAppState();
  const [form, setForm] = useState<RegisterInstitutionInput>(EMPTY);

  function set<K extends keyof RegisterInstitutionInput>(key: K, value: RegisterInstitutionInput[K]) {
    setForm((prev) => ({ ...prev, [key]: value }));
  }

  const canSubmit = form.name.trim() && form.city.trim() && form.street.trim() && form.area.trim() && form.phone.trim() && form.email.trim();

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    registerInstitution(form);
    router.push("/dashboard");
  }

  return (
    <div style={{ maxWidth: 640, margin: "0 auto", padding: "40px 28px 80px", fontFamily: FB }}>
      <div style={{ textAlign: "center", marginBottom: 28 }}>
        <Building2 size={26} style={{ color: C.teal, marginBottom: 10 }} />
        <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(22px,3vw,28px)", color: C.text, marginBottom: 8 }}>
          {t({ ru: "Регистрация учреждения", tg: "Бақайдгирии муассиса" })}
        </h1>
        <p style={{ fontSize: 13.5, color: C.sub }}>
          {t({ ru: "После заполнения заявка уйдёт на модерацию — в среднем это занимает до 48 часов.", tg: "Пас аз пуркунӣ ариза ба модератсия меравад — то 48 соат." })}
        </p>
      </div>

      <form onSubmit={submit} style={{ borderRadius: 18, border: `1px solid ${C.border}`, background: C.s1, padding: 24 }}>
        <FormField label={t({ ru: "Название учреждения", tg: "Номи муассиса" })}>
          <input required value={form.name} onChange={(e) => set("name", e.target.value)} style={inputStyle} />
        </FormField>

        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
          <FormField label={t({ ru: "Тип", tg: "Навъ" })}>
            <select value={form.tk} onChange={(e) => set("tk", e.target.value as CategoryKey)} style={inputStyle}>
              {CATEGORY_KEYS.map((k) => (<option key={k} value={k}>{t(CATEGORY_META[k].label)}</option>))}
            </select>
          </FormField>
          <FormField label={t({ ru: "Регион", tg: "Минтақа" })}>
            <select value={form.region} onChange={(e) => set("region", e.target.value as Region)} style={inputStyle}>
              {REGION_ORDER.map((r) => (<option key={r} value={r}>{t(REGION_LABEL[r])}</option>))}
            </select>
          </FormField>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
          <FormField label={t({ ru: "Город", tg: "Шаҳр" })}>
            <input required value={form.city} onChange={(e) => set("city", e.target.value)} style={inputStyle} />
          </FormField>
          <FormField label={t({ ru: "Район", tg: "Ноҳия" })}>
            <input required value={form.area} onChange={(e) => set("area", e.target.value)} style={inputStyle} />
          </FormField>
        </div>

        <FormField label={t({ ru: "Улица, дом", tg: "Кӯча, хона" })}>
          <input required value={form.street} onChange={(e) => set("street", e.target.value)} style={inputStyle} />
        </FormField>

        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: 14 }}>
          <FormField label={t({ ru: "Цена, сомони/мес", tg: "Нарх, сомонӣ/моҳ" })}>
            <input type="number" min={0} value={form.price} onChange={(e) => set("price", Number(e.target.value))} style={inputStyle} />
          </FormField>
          <FormField label={t({ ru: "Год основания", tg: "Соли таъсис" })}>
            <input type="number" value={form.founded} onChange={(e) => set("founded", Number(e.target.value))} style={inputStyle} />
          </FormField>
          <FormField label={t({ ru: "Учеников", tg: "Хонандагон" })}>
            <input type="number" min={0} value={form.students} onChange={(e) => set("students", Number(e.target.value))} style={inputStyle} />
          </FormField>
        </div>

        <FormField label={t({ ru: "Возрастная группа (например «7-17 лет»)", tg: "Гурӯҳи синнусолӣ (масалан «7-17 сол»)" })}>
          <input value={form.age} onChange={(e) => set("age", e.target.value)} style={inputStyle} />
        </FormField>

        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
          <FormField label={t({ ru: "Телефон", tg: "Телефон" })}>
            <input required value={form.phone} onChange={(e) => set("phone", e.target.value)} style={inputStyle} />
          </FormField>
          <FormField label="Email">
            <input required type="email" value={form.email} onChange={(e) => set("email", e.target.value)} style={inputStyle} />
          </FormField>
        </div>

        <FormField label={t({ ru: "Сайт (необязательно)", tg: "Сомона (ихтиёрӣ)" })}>
          <input value={form.website} onChange={(e) => set("website", e.target.value)} style={inputStyle} />
        </FormField>

        <FormField label={t({ ru: "Описание", tg: "Тавсиф" })}>
          <textarea rows={4} value={form.description} onChange={(e) => set("description", e.target.value)} style={{ ...inputStyle, resize: "vertical" as const }} />
        </FormField>

        <button type="submit" disabled={!canSubmit} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: canSubmit ? "pointer" : "default", opacity: canSubmit ? 1 : 0.5, marginTop: 4 }}>
          {t({ ru: "Отправить на модерацию", tg: "Ба модератсия фиристодан" })}
        </button>
      </form>
    </div>
  );
}
