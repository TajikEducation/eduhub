import { baseApi } from "@/lib/store/baseApi";

// Форма ответа реального бэкенда (GET /api/v1/institutions, см.
// backend/internal/catalog/transport/http/{dto.go,query.go}). Отличается от
// mock-типа Institution в lib/data.ts — здесь нет staff/vacancies/gallery и т.п.,
// бэкенд пока отдаёт только каталожные поля.
export interface BilingualText {
  ru: string;
  tg: string;
}

export interface BackendInstitution {
  id: string;
  name: BilingualText;
  types: string[];
  region: string;
  city: BilingualText;
  district: string;
  price: number;
  rating_avg: number;
  review_count: number;
  verified: boolean;
  discount_available: boolean;
  cover_photo_s3_key: string;
  tag: BilingualText | null;
  lat: number;
  lng: number;
}

export interface InstitutionListResponse {
  items: BackendInstitution[];
  next_cursor: string | null;
  total_hint: number | null;
}

// Параметры зеркалят backend/internal/catalog/transport/http/query.go —
// parseListQuery(). Всё опционально: пустая строка/undefined == "не передан".
export interface InstitutionListParams {
  q?: string;
  type?: string; // CSV backend-типов: kindergarten|school|center|university
  region?: string;
  area?: string;
  min_price?: number;
  max_price?: number;
  min_rating?: number;
  transport?: boolean;
  food?: boolean;
  verified?: boolean;
  sort?: "" | "score" | "price_asc";
  lat?: number;
  lng?: number;
  radius_km?: number;
  limit?: number;
  cursor?: string;
}

export const searchApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getInstitutions: builder.query<InstitutionListResponse, InstitutionListParams | void>({
      query: (params) => ({
        url: "/api/v1/institutions",
        params: params ?? undefined,
      }),
      providesTags: (result) =>
        result
          ? [...result.items.map((i) => ({ type: "Institution" as const, id: i.id })), { type: "Institution" as const, id: "LIST" }]
          : [{ type: "Institution" as const, id: "LIST" }],
    }),
  }),
});

export const { useGetInstitutionsQuery } = searchApi;
