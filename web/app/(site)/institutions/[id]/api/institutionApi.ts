import { baseApi } from "@/lib/store/baseApi";
import type { BilingualText } from "@/app/(site)/search/api/searchApi";
import type { ReviewDTO, VacancyDTO, NewsArticleDTO } from "@/app/dashboard/api/dashboardApi";

// Форма ответа GET /api/v1/institutions/{id} — полная карточка (см.
// backend/internal/catalog/transport/http/dto.go, institutionDTO). Сознательно не включает
// News (репозиторий пока её не заполняет) и отзывы (отдельный эндпоинт).
export interface SocialsDTO {
  instagram?: string;
  telegram?: string;
  facebook?: string;
}

export interface StaffMemberDTO {
  id: string;
  name: BilingualText;
  role_type: string;
  role_label: BilingualText;
  subject?: BilingualText;
  photo_url?: string;
  exp?: string;
  bio?: BilingualText;
  email?: string;
  phone?: string;
}

export interface AchievementDTO {
  id: string;
  title: BilingualText;
  year: number;
  category: string;
  description: BilingualText;
  links: { label: string; url: string }[];
}

export interface GalleryItemDTO {
  id: string;
  s3_key: string;
  label?: BilingualText;
  sort_order: number;
}

export interface AlumnusDTO {
  id: string;
  name: BilingualText;
  photo_url?: string;
  grad_year: number;
  now_label?: BilingualText;
}

export interface TransportRouteDTO {
  id: string;
  type: string;
  label?: BilingualText;
  areas: BilingualText[];
  cost?: number;
  cost_period: string;
  sort_order: number;
}

export interface MealPlanDTO {
  id: string;
  meal_type: string;
  label?: BilingualText;
  cost?: number;
  cost_period: string;
  halal?: boolean;
  sort_order: number;
}

export interface InstitutionDetailDTO {
  id: string;
  name: BilingualText;
  types: string[];
  region: string;
  city?: BilingualText;
  district?: string;
  description?: BilingualText;
  address?: BilingualText;
  lat: number;
  lng: number;
  location_landmarks?: string;
  phone?: string;
  email?: string;
  website?: string;
  socials?: SocialsDTO;
  cover_photo_s3_key?: string;
  age_range?: string;
  tag?: BilingualText;
  license_no?: string;
  languages?: string[];
  program_level?: string[];
  curriculum?: string[];
  price?: number;
  discount_available: boolean;
  discount_type?: string[];
  discount_details?: string;
  verified: boolean;
  founded?: number;
  students_count?: number;
  rating_avg?: number;
  review_count: number;
  created_at: string;
  updated_at: string;

  staff: StaffMemberDTO[];
  achievements: AchievementDTO[];
  gallery: GalleryItemDTO[];
  alumni: AlumnusDTO[];
  transport_routes: TransportRouteDTO[];
  meal_plans: MealPlanDTO[];
}

export const institutionApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getInstitution: builder.query<InstitutionDetailDTO, string>({
      query: (id) => `/api/v1/institutions/${id}`,
      providesTags: (_r, _e, id) => [{ type: "Institution", id }],
    }),
    listInstitutionReviews: builder.query<{ items: ReviewDTO[] }, string>({
      query: (id) => `/api/v1/institutions/${id}/reviews`,
      providesTags: (_r, _e, id) => [{ type: "Institution", id }],
    }),
    createInstitutionReview: builder.mutation<ReviewDTO, { institutionId: string; rating: number; text: string }>({
      query: ({ institutionId, rating, text }) => ({
        url: `/api/v1/institutions/${institutionId}/reviews`,
        method: "POST",
        body: { rating, text },
      }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),
    listInstitutionVacancies: builder.query<{ items: VacancyDTO[] }, string>({
      query: (id) => `/api/v1/institutions/${id}/vacancies`,
      providesTags: (_r, _e, id) => [{ type: "Institution", id }],
    }),
    listInstitutionPublishedNews: builder.query<{ items: NewsArticleDTO[] }, string>({
      query: (id) => `/api/v1/institutions/${id}/news/published`,
      providesTags: (_r, _e, id) => [{ type: "Institution", id }],
    }),
  }),
});

export const {
  useGetInstitutionQuery,
  useListInstitutionReviewsQuery,
  useCreateInstitutionReviewMutation,
  useListInstitutionVacanciesQuery,
  useListInstitutionPublishedNewsQuery,
} = institutionApi;
