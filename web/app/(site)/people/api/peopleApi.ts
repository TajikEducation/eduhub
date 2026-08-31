import { baseApi } from "@/lib/store/baseApi";
import type { StaffMemberDTO } from "@/app/dashboard/api/dashboardApi";

// GET /api/v1/staff/{id} — публичный профиль сотрудника (см.
// backend/internal/catalog/transport/http/staff_public_handler.go, GetStaffHandler). Виден,
// только если институция сотрудника approved.
export interface PublicStaffMember extends StaffMemberDTO {
  institution_id: string;
}

export const peopleApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getStaffMember: builder.query<PublicStaffMember, string>({
      query: (id) => `/api/v1/staff/${id}`,
    }),
  }),
});

export const { useGetStaffMemberQuery } = peopleApi;
