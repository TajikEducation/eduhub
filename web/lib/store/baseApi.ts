import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { getAccessToken } from "@/lib/authToken";

// Единая точка входа ко всем бэкенд-запросам. Каждый route добавляет свои
// эндпоинты через baseApi.injectEndpoints() в собственном app/<route>/api/*.ts —
// сам baseApi здесь не содержит ни одного эндпоинта.
export const baseApi = createApi({
  reducerPath: "baseApi",
  baseQuery: fetchBaseQuery({
    baseUrl: process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:3001",
    prepareHeaders: (headers) => {
      const token = getAccessToken();
      if (token) headers.set("Authorization", `Bearer ${token}`);
      return headers;
    },
  }),
  tagTypes: ["Institution"],
  endpoints: () => ({}),
});
