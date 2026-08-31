import { baseApi } from "@/lib/store/baseApi";

// Форма ответа backend/internal/auth/transport/http (dto.go): userDTO/tokenPairDTO.
export interface AuthUser {
  id: string;
  email: string;
  role: "user" | "institution" | "moderator" | "admin";
  status: "unverified" | "active" | "banned" | "deleted";
  display_name?: string;
  created_at: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

export interface AuthResponse {
  user: AuthUser;
  tokens: TokenPair;
}

export interface RegisterRequest {
  email: string;
  password: string;
  display_name?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export const authApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    register: builder.mutation<AuthResponse, RegisterRequest>({
      query: (body) => ({ url: "/auth/register", method: "POST", body }),
    }),
    login: builder.mutation<AuthResponse, LoginRequest>({
      query: (body) => ({ url: "/auth/login", method: "POST", body }),
    }),
    me: builder.query<AuthUser, void>({
      query: () => "/auth/me",
    }),
  }),
});

export const { useRegisterMutation, useLoginMutation, useMeQuery } = authApi;
