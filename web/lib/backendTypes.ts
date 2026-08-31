import { CATEGORY_META, type CategoryKey } from "./data";
import type { InstitCardData } from "@/components/InstitCard";
import type { BackendInstitution } from "@/app/(site)/search/api/searchApi";

// cat_kg→kindergarten, cat_school→school, cat_center→center, cat_uni→university
// (см. backend/cmd/devseed/data.go — тот же маппинг используется при сидировании,
// и backend/internal/catalog/domain/write.go — validInstitutionTypes на бэкенде).
export const CATEGORY_TO_BACKEND_TYPE: Record<CategoryKey, string> = {
  cat_kg: "kindergarten",
  cat_school: "school",
  cat_center: "center",
  cat_uni: "university",
};

// Обратный маппинг — для отображения реальных backend-институций теми же UI-компонентами
// (InstitCard и т.п.), что и mock-каталог. Неизвестный/будущий backend-тип → cat_school
// (безопасный дефолт, а не падение рендера).
const BACKEND_TYPE_TO_CATEGORY: Record<string, CategoryKey> = {
  kindergarten: "cat_kg",
  school: "cat_school",
  center: "cat_center",
  university: "cat_uni",
};

export function backendTypeToCategory(type: string | undefined): CategoryKey {
  return (type && BACKEND_TYPE_TO_CATEGORY[type]) || "cat_school";
}

// backendInstitutionToCard адаптирует ответ GET /api/v1/institutions (см. searchApi.ts) под
// InstitCardData — узкий набор полей, которые реально читает InstitCard. Поля, которых нет
// в текущем backend-контракте списка (площадь развозки/питания как явные булевы флаги на
// уровне списка, программа обучения) — заполняются безопасными дефолтами, не выдумываются.
export function backendInstitutionToCard(inst: BackendInstitution, locale: "ru" | "tg"): InstitCardData {
  const tk = backendTypeToCategory(inst.types[0]);
  return {
    id: inst.id,
    tk,
    color: CATEGORY_META[tk].color,
    coverPhoto: inst.cover_photo_s3_key || CATEGORY_META[tk].heroPhoto,
    tag: inst.tag ? { ru: inst.tag.ru, tg: inst.tag.tg } : null,
    area: inst.district || inst.city?.[locale] || "",
    name: { ru: inst.name.ru, tg: inst.name.tg },
    score: inst.rating_avg ?? 0,
    rev: inst.review_count,
    ver: inst.verified,
    transport: false,
    food: false,
    price: inst.price ?? 0,
  };
}
