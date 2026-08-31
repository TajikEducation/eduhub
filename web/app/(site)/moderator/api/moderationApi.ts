import { baseApi } from "@/lib/store/baseApi";
import type { BilingualText } from "@/app/(site)/search/api/searchApi";

// Форма ответа GET /api/v1/moderation/institutions (moderator/admin, см.
// backend/internal/catalog/transport/http/staff_public_handler.go, ListForModerationHandler) —
// шире публичного каталога: включает moderation_status и профильные поля.
export interface ModerationInstitution {
  id: string;
  name: BilingualText;
  types: string[];
  region: string;
  address?: BilingualText;
  phone?: string;
  email?: string;
  description?: BilingualText;
  price?: number;
  moderation_status: "pending" | "approved" | "rejected";
  created_at: string;
}

export const moderationApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listModerationInstitutions: builder.query<{ items: ModerationInstitution[] }, { status?: string } | void>({
      query: (params) => ({ url: "/api/v1/moderation/institutions", params: params ?? undefined }),
      providesTags: [{ type: "Institution", id: "MODERATION_QUEUE" }],
    }),
    approveInstitution: builder.mutation<void, string>({
      query: (id) => ({ url: `/api/v1/moderation/institutions/${id}/approve`, method: "POST" }),
      invalidatesTags: [{ type: "Institution", id: "MODERATION_QUEUE" }],
    }),
    rejectInstitution: builder.mutation<void, { id: string; reason_code: string; reason_text: string }>({
      query: ({ id, ...body }) => ({ url: `/api/v1/moderation/institutions/${id}/reject`, method: "POST", body }),
      invalidatesTags: [{ type: "Institution", id: "MODERATION_QUEUE" }],
    }),
  }),
});

export const { useListModerationInstitutionsQuery, useApproveInstitutionMutation, useRejectInstitutionMutation } = moderationApi;
