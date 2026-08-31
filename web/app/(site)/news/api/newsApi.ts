import { baseApi } from "@/lib/store/baseApi";
import type { NewsArticleDTO } from "@/app/dashboard/api/dashboardApi";

// GET /api/v1/news/{id} — одна опубликованная новость, публичный доступ (см.
// backend/internal/catalog/transport/http/news_handler.go, GetNewsHandler). Каждый успешный
// запрос атомарно увеличивает views_count на бэкенде.
export const newsApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    getNewsArticle: builder.query<NewsArticleDTO, string>({
      query: (id) => `/api/v1/news/${id}`,
    }),
  }),
});

export const { useGetNewsArticleQuery } = newsApi;
