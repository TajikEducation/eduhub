"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Building2 } from "lucide-react";
import { C, FH, FB, CATEGORY_META, REGION_ORDER, REGION_LABEL, type CategoryKey, type Region } from "@/lib/data";
import { REGION_CENTROIDS } from "@/lib/geo";
import { CATEGORY_TO_BACKEND_TYPE } from "@/lib/backendTypes";
import Link from "next/link";
import { useAppState } from "@/lib/app-state";
import { useT } from "@/lib/i18n";
import { useCreateInstitutionMutation } from "./api/registerInstitutionApi";

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

interface FormState {
  nameRu: string;
  nameTg: string;
  tk: CategoryKey;
  region: Region;
  city: string;
  district: string;
  price: number;
  phone: string;
  email: string;
  website: string;
  description: string;
}

const EMPTY: FormState = {
  nameRu: "", nameTg: "", tk: "cat_school", region: "dushanbe", city: "", district: "",
  price: 0, phone: "", email: "", website: "", description: "",
};

export default function RegisterInstitutionPage() {
  const t = useT();
  const router = useRouter();
  const { role } = useAppState();
  const [form, setForm] = useState<FormState>(EMPTY);
  const [createInstitution, { isLoading, error }] = useCreateInstitutionMutation();
  const [done, setDone] = useState(false);

  // заявку на регистрацию учреждения может подать только вошедший пользователь —
  // POST /api/v1/institutions требует Bearer-токен (owner_id берётся из него).
  if (role === "guest") {
    return (
      <div style={{ maxWidth: 480, margin: "0 auto", padding: "80px 28px", textAlign: "center" }}>
        <Building2 size={28} style={{ color: C.teal, margin: "0 auto 14px" }} />
        <h1 style={{ fontFamily: FH, fontWeight: 800, fontSize: 19, color: C.text, marginBottom: 10 }}>
          {t({ ru: "Сначала войдите в аккаунт", tg: "Аввал ба ҳисоб ворид шавед" })}
        </h1>
        <p style={{ fontSize: 13.5, color: C.sub, marginBottom: 22, lineHeight: 1.6 }}>
          {t({ ru: "Заявку на регистрацию учреждения подаёт представитель — от имени своего аккаунта.", tg: "Дархостро барои сабти муассиса намояндаи он — аз номи ҳисоби худ пешниҳод мекунад." })}
        </p>
        <Link href="/login?next=/register-institution" style={{ display: "inline-flex", padding: "12px 26px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, textDecoration: "none" }}>
          {t({ ru: "Войти", tg: "Ворид шудан" })}
        </Link>
      </div>
    );
  }

  function set<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((prev) => ({ ...prev, [key]: value }));
  }

  const canSubmit = form.nameRu.trim() && form.nameTg.trim() && form.city.trim() && form.district.trim() && form.phone.trim() && form.email.trim();

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    const centroid = REGION_CENTROIDS[form.region];
    try {
      await createInstitution({
        name: { ru: form.nameRu, tg: form.nameTg },
        types: [CATEGORY_TO_BACKEND_TYPE[form.tk]],
        region: form.region,
        city: { ru: form.city, tg: form.city },
        district: form.district,
        description: form.description ? { ru: form.description, tg: form.description } : undefined,
        phone: form.phone,
        email: form.email,
        website: form.website || undefined,
        price: form.price || undefined,
        lat: centroid.lat,
        lng: centroid.lng,
      }).unwrap();
      setDone(true);
    } catch {
      // ошибка уже в error через хук — рендерится ниже
    }
  }

  if (done) {
    return (
      <div style={{ maxWidth: 480, margin: "0 auto", padding: "80px 28px", textAlign: "center" }}>
        <Building2 size={28} style={{ color: C.teal, margin: "0 auto 14px" }} />
        <h1 style={{ fontFamily: FH, fontWeight: 800, fontSize: 19, color: C.text, marginBottom: 10 }}>
          {t({ ru: "Заявка отправлена", tg: "Дархост фиристода шуд" })}
        </h1>
        <p style={{ fontSize: 13.5, color: C.sub, marginBottom: 22, lineHeight: 1.6 }}>
          {t({ ru: "Учреждение сохранено на backend (POST /api/v1/institutions) со статусом pending и появится в каталоге после одобрения модератором.", tg: "Муассиса дар backend (POST /api/v1/institutions) бо ҳолати pending нигоҳ дошта шуд ва пас аз тасдиқи модератор дар каталог пайдо мешавад." })}
        </p>
        <Link href="/" style={{ display: "inline-flex", padding: "12px 26px", borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, textDecoration: "none" }}>
          {t({ ru: "На главную", tg: "Ба саҳифаи асосӣ" })}
        </Link>
      </div>
    );
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
        <div className="eh-mobile-1col" style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
          <FormField label={t({ ru: "Название (рус.)", tg: "Ном (рус.)" })}>
            <input required value={form.nameRu} onChange={(e) => set("nameRu", e.target.value)} style={inputStyle} />
          </FormField>
          <FormField label={t({ ru: "Название (тадж.)", tg: "Ном (тоҷ.)" })}>
            <input required value={form.nameTg} onChange={(e) => set("nameTg", e.target.value)} style={inputStyle} />
          </FormField>
        </div>

        <div className="eh-mobile-1col" style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
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

        <div className="eh-mobile-1col" style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
          <FormField label={t({ ru: "Город", tg: "Шаҳр" })}>
            <input required value={form.city} onChange={(e) => set("city", e.target.value)} style={inputStyle} />
          </FormField>
          <FormField label={t({ ru: "Район", tg: "Ноҳия" })}>
            <input required value={form.district} onChange={(e) => set("district", e.target.value)} style={inputStyle} />
          </FormField>
        </div>

        <FormField label={t({ ru: "Цена, сомони/мес", tg: "Нарх, сомонӣ/моҳ" })}>
          <input type="number" min={0} value={form.price} onChange={(e) => set("price", Number(e.target.value))} style={inputStyle} />
        </FormField>

        <div className="eh-mobile-1col" style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 14 }}>
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

        {error && (
          <p style={{ fontSize: 12.5, color: C.red, marginBottom: 10 }}>
            {t({ ru: "Не удалось отправить заявку. Проверьте данные.", tg: "Дархост фиристода нашуд. Маълумотро санҷед." })}
          </p>
        )}

        <button type="submit" disabled={!canSubmit || isLoading} style={{ width: "100%", padding: 13, borderRadius: 11, background: C.teal, color: C.overlay, fontFamily: FH, fontWeight: 800, fontSize: 14, border: "none", cursor: canSubmit ? "pointer" : "default", opacity: canSubmit && !isLoading ? 1 : 0.5, marginTop: 4 }}>
          {isLoading ? t({ ru: "Отправка…", tg: "Фиристода истодааст…" }) : t({ ru: "Отправить на модерацию", tg: "Ба модератсия фиристодан" })}
        </button>
      </form>
    </div>
  );
}
