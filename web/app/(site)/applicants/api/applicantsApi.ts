import { baseApi } from "@/lib/store/baseApi";
import type { BilingualText } from "@/app/(site)/search/api/searchApi";

// Веха 5, ядро — профиль соискателя + отклики на вакансии (см. backend/internal/applicants).
// Без employer_responses (приглашения работодателя) — не запрошено явно.
export interface ApplicantDTO {
  id: string;
  user_id: string;
  name: BilingualText;
  photo_url?: string;
  position: BilingualText;
  bio?: BilingualText;
  education?: BilingualText[];
  experience?: BilingualText[];
  skills?: BilingualText[];
  email?: string;
  phone?: string;
  cv_s3_key?: string;
  visibility: "draft" | "on_response" | "public";
  hide_contacts: boolean;
  created_at: string;
  updated_at: string;
}

export interface ApplicantRequest {
  name: BilingualText;
  photo_url?: string;
  position: BilingualText;
  bio?: BilingualText;
  education?: BilingualText[];
  experience?: BilingualText[];
  skills?: BilingualText[];
  email?: string;
  phone?: string;
  cv_s3_key?: string;
  visibility: "draft" | "on_response" | "public";
  hide_contacts: boolean;
}

export interface ApplicantAchievementDTO {
  id: string;
  applicant_id: string;
  title: string;
  year?: number;
  category?: "gold" | "silver" | "bronze" | "special";
  description?: string;
  created_at: string;
}

export interface ApplicantAchievementRequest {
  title: string;
  year?: number;
  category?: "gold" | "silver" | "bronze" | "special";
  description?: string;
}

export const applicantsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listApplicants: builder.query<{ items: ApplicantDTO[] }, void>({
      query: () => "/api/v1/applicants",
      providesTags: [{ type: "Institution", id: "APPLICANTS" }],
    }),
    getApplicant: builder.query<ApplicantDTO, string>({
      query: (id) => `/api/v1/applicants/${id}`,
      providesTags: (_r, _e, id) => [{ type: "Institution", id: `applicant-${id}` }],
    }),
    listApplicantAchievements: builder.query<{ items: ApplicantAchievementDTO[] }, string>({
      query: (id) => `/api/v1/applicants/${id}/achievements`,
      providesTags: (_r, _e, id) => [{ type: "Institution", id: `applicant-achievements-${id}` }],
    }),
    getMyApplicantProfile: builder.query<ApplicantDTO, void>({
      query: () => "/api/v1/applicants/me",
      providesTags: [{ type: "Institution", id: "MY_APPLICANT" }],
    }),
    upsertMyApplicantProfile: builder.mutation<ApplicantDTO, ApplicantRequest>({
      query: (body) => ({ url: "/api/v1/applicants/me", method: "PUT", body }),
      invalidatesTags: [{ type: "Institution", id: "MY_APPLICANT" }, { type: "Institution", id: "APPLICANTS" }],
    }),
    createMyAchievement: builder.mutation<ApplicantAchievementDTO, ApplicantAchievementRequest>({
      query: (body) => ({ url: "/api/v1/applicants/me/achievements", method: "POST", body }),
      invalidatesTags: (result) => (result ? [{ type: "Institution", id: `applicant-achievements-${result.applicant_id}` }] : []),
    }),
    deleteMyAchievement: builder.mutation<void, { achId: string; applicantId: string }>({
      query: ({ achId }) => ({ url: `/api/v1/applicants/achievements/${achId}`, method: "DELETE" }),
      invalidatesTags: (_r, _e, { applicantId }) => [{ type: "Institution", id: `applicant-achievements-${applicantId}` }],
    }),
    applyToVacancy: builder.mutation<{ vacancy_id: string }, string>({
      query: (vacancyId) => ({ url: `/api/v1/vacancies/${vacancyId}/apply`, method: "POST" }),
      invalidatesTags: [{ type: "Institution", id: "MY_APPLICATIONS" }],
    }),
    listMyApplications: builder.query<{ vacancy_ids: string[] }, void>({
      query: () => "/api/v1/applicants/me/applications",
      providesTags: [{ type: "Institution", id: "MY_APPLICATIONS" }],
    }),
  }),
});

export const {
  useListApplicantsQuery,
  useGetApplicantQuery,
  useListApplicantAchievementsQuery,
  useGetMyApplicantProfileQuery,
  useUpsertMyApplicantProfileMutation,
  useCreateMyAchievementMutation,
  useDeleteMyAchievementMutation,
  useApplyToVacancyMutation,
  useListMyApplicationsQuery,
} = applicantsApi;
