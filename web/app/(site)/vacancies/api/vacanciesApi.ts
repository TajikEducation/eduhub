import { baseApi } from "@/lib/store/baseApi";
import type { BilingualText } from "@/app/(site)/search/api/searchApi";

// Форма ответа реального бэкенда (GET /api/v1/vacancies, .../{id}), см.
// backend/internal/vacancies/transport/http/dto.go. Каждая вакансия обогащена сводкой об
// учреждении (institution) — регион/тип/поиск фильтруются на фронте по этой сводке, без
// серверной пагинации по фильтрам (небольшой MVP-каталог).
export interface VacancyInstitutionSummary {
  id: string;
  name: BilingualText;
  types: string[];
  region: string;
  district?: string;
  city?: BilingualText;
  cover_photo_s3_key?: string;
  verified: boolean;
}

export interface PublicVacancy {
  id: string;
  institution_id: string;
  title: BilingualText;
  description: BilingualText;
  requirements?: BilingualText[];
  salary_from?: number;
  salary_to?: number;
  employment: BilingualText;
  status: "draft" | "published";
  created_at: string;
  updated_at: string;
  institution: VacancyInstitutionSummary;
}

export const vacanciesApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getVacancies: builder.query<{ items: PublicVacancy[] }, { limit?: number } | void>({
      query: (params) => ({ url: "/api/v1/vacancies", params: params ?? undefined }),
      providesTags: (result) =>
        result
          ? [...result.items.map((v) => ({ type: "Institution" as const, id: `vacancy-${v.id}` })), { type: "Institution" as const, id: "VACANCIES" }]
          : [{ type: "Institution" as const, id: "VACANCIES" }],
    }),
    getVacancy: builder.query<PublicVacancy, string>({
      query: (id) => `/api/v1/vacancies/${id}`,
      providesTags: (_r, _e, id) => [{ type: "Institution", id: `vacancy-${id}` }],
    }),
  }),
});

export const { useGetVacanciesQuery, useGetVacancyQuery } = vacanciesApi;
