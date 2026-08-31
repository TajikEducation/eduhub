import { baseApi } from "@/lib/store/baseApi";
import type { BilingualText, BackendInstitution } from "../../search/api/searchApi";

export interface CreateInstitutionRequest {
  name: BilingualText;
  types: string[];
  region: string;
  city?: BilingualText;
  district?: string;
  description?: BilingualText;
  phone?: string;
  email?: string;
  website?: string;
  price?: number;
  lat: number;
  lng: number;
}

export const registerInstitutionApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    createInstitution: builder.mutation<BackendInstitution, CreateInstitutionRequest>({
      query: (body) => ({ url: "/api/v1/institutions", method: "POST", body }),
      invalidatesTags: [{ type: "Institution", id: "LIST" }],
    }),
  }),
});

export const { useCreateInstitutionMutation } = registerInstitutionApi;
