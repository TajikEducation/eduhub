import { baseApi } from "@/lib/store/baseApi";

// Веха 5, ядро — REST-поллинг вместо WebSocket/Redis Pub/Sub (см.
// backend/internal/communications, пакетный комментарий domain/chat.go). counterpart/sender
// намеренно не несут имени — модуль communications не знает о catalog/auth (владение схемами);
// фронт добирает имя институции через institutionApi, для пользователя показывает общую метку.
export type ParticipantType = "user" | "institution";

export interface ConversationDTO {
  id: string;
  counterpart_type: ParticipantType;
  counterpart_id: string;
  created_at: string;
}

export interface MessageDTO {
  id: string;
  conversation_id: string;
  sender_type: ParticipantType;
  sender_id: string;
  body: string;
  created_at: string;
}

export interface CreateConversationRequest {
  counterpart_type: ParticipantType;
  counterpart_id: string;
  as_institution_id?: string;
}

const POLL_INTERVAL_SECONDS = 4;

export const chatApi = baseApi.injectEndpoints({
  endpoints: (builder) => ({
    listConversations: builder.query<{ items: ConversationDTO[] }, { asInstitutionId?: string } | void>({
      query: (params) => ({ url: "/api/v1/conversations", params: params?.asInstitutionId ? { as_institution_id: params.asInstitutionId } : undefined }),
      providesTags: [{ type: "Institution", id: "CONVERSATIONS" }],
    }),
    createConversation: builder.mutation<ConversationDTO, CreateConversationRequest>({
      query: (body) => ({ url: "/api/v1/conversations", method: "POST", body }),
      invalidatesTags: [{ type: "Institution", id: "CONVERSATIONS" }],
    }),
    listMessages: builder.query<{ items: MessageDTO[] }, { conversationId: string; asInstitutionId?: string }>({
      query: ({ conversationId, asInstitutionId }) => ({
        url: `/api/v1/conversations/${conversationId}/messages`,
        params: asInstitutionId ? { as_institution_id: asInstitutionId } : undefined,
      }),
      providesTags: (_r, _e, { conversationId }) => [{ type: "Institution", id: `messages-${conversationId}` }],
    }),
    sendMessage: builder.mutation<MessageDTO, { conversationId: string; body: string; asInstitutionId?: string }>({
      query: ({ conversationId, body, asInstitutionId }) => ({
        url: `/api/v1/conversations/${conversationId}/messages`,
        method: "POST",
        body: { body, as_institution_id: asInstitutionId },
      }),
      invalidatesTags: (_r, _e, { conversationId }) => [{ type: "Institution", id: `messages-${conversationId}` }],
    }),
  }),
});

export const POLL_INTERVAL_MS = POLL_INTERVAL_SECONDS * 1000;
export const { useListConversationsQuery, useCreateConversationMutation, useListMessagesQuery, useSendMessageMutation } = chatApi;
