import { baseApi } from "@/lib/store/baseApi";
import type { BilingualText } from "../../(site)/search/api/searchApi";

// Формы ответов реального бэкенда (backend/internal/catalog/transport/http/dto.go).
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

export interface AchievementLinkDTO {
  label: string;
  url: string;
}

export interface AchievementDTO {
  id: string;
  title: BilingualText;
  year: number;
  category: string;
  description: BilingualText;
  links: AchievementLinkDTO[];
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

export interface NewsArticleDTO {
  id: string;
  institution_id: string;
  title: BilingualText;
  category?: BilingualText;
  cover_s3_key?: string;
  video_url?: string;
  content: BilingualText;
  tags?: BilingualText[];
  status: "draft" | "published";
  views_count: number;
  created_at: string;
  updated_at: string;
}

export interface FullInstitutionDTO {
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
  phone?: string;
  email?: string;
  website?: string;
  cover_photo_s3_key?: string;
  age_range?: string;
  price?: number;
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
}

export interface MineListResponse {
  items: FullInstitutionDTO[];
}

export interface UpdateInstitutionRequest {
  description?: BilingualText;
  phone?: string;
  email?: string;
  website?: string;
  cover_photo_s3_key?: string;
  price?: number;
  age_range?: string;
}

export interface StaffRequest {
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

export interface AchievementRequest {
  title: BilingualText;
  year: number;
  category: string;
  description: BilingualText;
  links?: AchievementLinkDTO[];
}

export interface GalleryItemRequest {
  s3_key: string;
  label?: BilingualText;
  sort_order: number;
}

export interface AlumnusRequest {
  name: BilingualText;
  photo_url?: string;
  grad_year: number;
  now_label?: BilingualText;
}

export interface NewsRequest {
  title: BilingualText;
  category?: BilingualText;
  cover_s3_key?: string;
  video_url?: string;
  content: BilingualText;
  tags?: BilingualText[];
  status: "draft" | "published";
}

export interface VacancyDTO {
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
}

export interface VacancyRequest {
  title: BilingualText;
  description: BilingualText;
  requirements?: BilingualText[];
  salary_from?: number;
  salary_to?: number;
  employment: BilingualText;
  status: "draft" | "published";
}

export interface ReviewDTO {
  id: string;
  institution_id: string;
  user_id: string;
  rating: number;
  text: string;
  reply?: string;
  replied_at?: string;
  status: "pending" | "approved" | "rejected";
  created_at: string;
}

export const dashboardApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listMyReviews: builder.query<{ items: ReviewDTO[] }, string>({
      query: (institutionId) => `/api/v1/institutions/${institutionId}/reviews/mine`,
      providesTags: (_r, _e, institutionId) => [{ type: "Institution", id: institutionId }],
    }),
    replyToReview: builder.mutation<ReviewDTO, { institutionId: string; reviewId: string; reply: string }>({
      query: ({ reviewId, reply }) => ({ url: `/api/v1/reviews/${reviewId}/reply`, method: "POST", body: { reply } }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),
    getMine: builder.query<MineListResponse, void>({
      query: () => "/api/v1/institutions/mine",
      providesTags: [{ type: "Institution", id: "MINE" }],
    }),
    getMineFull: builder.query<FullInstitutionDTO, string>({
      query: (id) => `/api/v1/institutions/${id}/mine`,
      providesTags: (_r, _e, id) => [{ type: "Institution", id }],
    }),
    listNews: builder.query<{ items: NewsArticleDTO[] }, string>({
      query: (institutionId) => `/api/v1/institutions/${institutionId}/news`,
      providesTags: (_r, _e, institutionId) => [{ type: "Institution", id: institutionId }],
    }),
    updateInstitution: builder.mutation<FullInstitutionDTO, { id: string; body: UpdateInstitutionRequest }>({
      query: ({ id, body }) => ({ url: `/api/v1/institutions/${id}`, method: "PATCH", body }),
      invalidatesTags: (_r, _e, { id }) => [{ type: "Institution", id }],
    }),

    createStaff: builder.mutation<StaffMemberDTO, { institutionId: string; body: StaffRequest }>({
      query: ({ institutionId, body }) => ({ url: `/api/v1/institutions/${institutionId}/staff`, method: "POST", body }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),
    updateStaff: builder.mutation<StaffMemberDTO, { institutionId: string; staffId: string; body: StaffRequest }>({
      query: ({ institutionId, staffId, body }) => ({ url: `/api/v1/institutions/${institutionId}/staff/${staffId}`, method: "PATCH", body }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),
    deleteStaff: builder.mutation<void, { institutionId: string; staffId: string }>({
      query: ({ institutionId, staffId }) => ({ url: `/api/v1/institutions/${institutionId}/staff/${staffId}`, method: "DELETE" }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),

    createAchievement: builder.mutation<AchievementDTO, { institutionId: string; body: AchievementRequest }>({
      query: ({ institutionId, body }) => ({ url: `/api/v1/institutions/${institutionId}/achievements`, method: "POST", body }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),
    deleteAchievement: builder.mutation<void, { institutionId: string; achId: string }>({
      query: ({ institutionId, achId }) => ({ url: `/api/v1/institutions/${institutionId}/achievements/${achId}`, method: "DELETE" }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),

    createGalleryItem: builder.mutation<GalleryItemDTO, { institutionId: string; body: GalleryItemRequest }>({
      query: ({ institutionId, body }) => ({ url: `/api/v1/institutions/${institutionId}/gallery`, method: "POST", body }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),
    deleteGalleryItem: builder.mutation<void, { institutionId: string; itemId: string }>({
      query: ({ institutionId, itemId }) => ({ url: `/api/v1/institutions/${institutionId}/gallery/${itemId}`, method: "DELETE" }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),

    createAlumnus: builder.mutation<AlumnusDTO, { institutionId: string; body: AlumnusRequest }>({
      query: ({ institutionId, body }) => ({ url: `/api/v1/institutions/${institutionId}/alumni`, method: "POST", body }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),
    deleteAlumnus: builder.mutation<void, { institutionId: string; alumnusId: string }>({
      query: ({ institutionId, alumnusId }) => ({ url: `/api/v1/institutions/${institutionId}/alumni/${alumnusId}`, method: "DELETE" }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),

    createNews: builder.mutation<NewsArticleDTO, { institutionId: string; body: NewsRequest }>({
      query: ({ institutionId, body }) => ({ url: `/api/v1/institutions/${institutionId}/news`, method: "POST", body }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),
    updateNews: builder.mutation<NewsArticleDTO, { institutionId: string; newsId: string; body: NewsRequest }>({
      query: ({ institutionId, newsId, body }) => ({ url: `/api/v1/institutions/${institutionId}/news/${newsId}`, method: "PATCH", body }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),
    deleteNews: builder.mutation<void, { institutionId: string; newsId: string }>({
      query: ({ institutionId, newsId }) => ({ url: `/api/v1/institutions/${institutionId}/news/${newsId}`, method: "DELETE" }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),

    listMyVacancies: builder.query<{ items: VacancyDTO[] }, string>({
      query: (institutionId) => `/api/v1/institutions/${institutionId}/vacancies/mine`,
      providesTags: (_r, _e, institutionId) => [{ type: "Institution", id: institutionId }],
    }),
    createVacancy: builder.mutation<VacancyDTO, { institutionId: string; body: VacancyRequest }>({
      query: ({ institutionId, body }) => ({ url: `/api/v1/institutions/${institutionId}/vacancies`, method: "POST", body }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),
    updateVacancy: builder.mutation<VacancyDTO, { institutionId: string; vacancyId: string; body: VacancyRequest }>({
      query: ({ vacancyId, body }) => ({ url: `/api/v1/vacancies/${vacancyId}`, method: "PATCH", body }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),
    deleteVacancy: builder.mutation<void, { institutionId: string; vacancyId: string }>({
      query: ({ vacancyId }) => ({ url: `/api/v1/vacancies/${vacancyId}`, method: "DELETE" }),
      invalidatesTags: (_r, _e, { institutionId }) => [{ type: "Institution", id: institutionId }],
    }),
  }),
});

export const {
  useListMyReviewsQuery,
  useReplyToReviewMutation,
  useGetMineQuery,
  useGetMineFullQuery,
  useListNewsQuery,
  useUpdateInstitutionMutation,
  useCreateStaffMutation,
  useUpdateStaffMutation,
  useDeleteStaffMutation,
  useCreateAchievementMutation,
  useDeleteAchievementMutation,
  useCreateGalleryItemMutation,
  useDeleteGalleryItemMutation,
  useCreateAlumnusMutation,
  useDeleteAlumnusMutation,
  useCreateNewsMutation,
  useUpdateNewsMutation,
  useDeleteNewsMutation,
  useListMyVacanciesQuery,
  useCreateVacancyMutation,
  useUpdateVacancyMutation,
  useDeleteVacancyMutation,
} = dashboardApi;
