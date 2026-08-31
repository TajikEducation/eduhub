"use client";

import { Suspense, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Send, MessagesSquare } from "lucide-react";
import { C, FH, FB, PHOTOS } from "@/lib/data";
import { useT } from "@/lib/i18n";
import { useMeQuery } from "@/app/(site)/login/api/authApi";
import { getAccessToken } from "@/lib/authToken";
import { useGetInstitutionQuery } from "@/app/(site)/institutions/[id]/api/institutionApi";
import {
  useListConversationsQuery, useListMessagesQuery, useSendMessageMutation,
  POLL_INTERVAL_MS, type ConversationDTO,
} from "./api/chatApi";

function ConversationLabel({ conv }: { conv: ConversationDTO }) {
  const { data: inst } = useGetInstitutionQuery(conv.counterpart_id, { skip: conv.counterpart_type !== "institution" });
  const t = useT();
  if (conv.counterpart_type === "institution") return <>{inst?.name.ru ?? "…"}</>;
  return <>{t({ ru: "Пользователь", tg: "Корбар" })}</>;
}

export default function MessagesPage() {
  return (
    <Suspense fallback={null}>
      <MessagesInner />
    </Suspense>
  );
}

function MessagesInner() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const t = useT();
  const { data: me, isLoading: meLoading } = useMeQuery(undefined, { skip: !getAccessToken() });

  const { data: convData } = useListConversationsQuery(undefined, { pollingInterval: POLL_INTERVAL_MS, skip: !me });
  const conversations = convData?.items ?? [];

  const [selectedConv, setSelectedConv] = useState<string | null>(searchParams.get("conv"));
  const activeConv = selectedConv ?? conversations[0]?.id ?? null;

  const { data: msgData } = useListMessagesQuery({ conversationId: activeConv ?? "" }, { skip: !activeConv, pollingInterval: POLL_INTERVAL_MS });
  const thread = msgData?.items ?? [];
  const conv = conversations.find((c) => c.id === activeConv);

  const [sendMessage] = useSendMessageMutation();
  const [text, setText] = useState("");

  async function send() {
    if (!text.trim() || !activeConv) return;
    const body = text.trim();
    setText("");
    try {
      await sendMessage({ conversationId: activeConv, body }).unwrap();
    } catch {
      setText(body);
    }
  }

  if (meLoading) {
    return <div style={{ padding: 60, textAlign: "center", color: C.muted, fontFamily: FB }}>{t({ ru: "Загрузка…", tg: "Боркунӣ…" })}</div>;
  }
  if (!me) {
    return (
      <div style={{ maxWidth: 480, margin: "0 auto", padding: "60px 28px", textAlign: "center", fontFamily: FB }}>
        <MessagesSquare size={26} style={{ color: C.dim, marginBottom: 12 }} />
        <p style={{ fontFamily: FH, fontWeight: 800, fontSize: 17, color: C.text, marginBottom: 8 }}>
          {t({ ru: "Войдите, чтобы читать сообщения", tg: "Барои хондани паёмҳо ворид шавед" })}
        </p>
        <button onClick={() => router.push("/login")} style={{ marginTop: 6, color: C.teal, fontFamily: FH, fontWeight: 700, fontSize: 13.5, background: "none", border: "none", cursor: "pointer" }}>
          {t("nav.login")}
        </button>
      </div>
    );
  }

  return (
    <div style={{ fontFamily: FB }}>
      <div style={{ position: "relative", height: 130, overflow: "hidden" }}>
        <img src={PHOTOS.classroom2} alt="" style={{ position: "absolute", inset: 0, width: "100%", height: "100%", objectFit: "cover" }} />
        <div style={{ position: "absolute", inset: 0, background: `linear-gradient(180deg, ${C.overlay}66 0%, ${C.overlay}D8 80%, ${C.overlay} 100%)` }} />
        <div style={{ position: "relative", height: "100%", display: "flex", alignItems: "flex-end", maxWidth: 1100, margin: "0 auto", padding: "0 28px 18px" }}>
          <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
            <div style={{ width: 34, height: 34, borderRadius: 10, background: `${C.teal}30`, border: `1px solid ${C.teal}66`, display: "flex", alignItems: "center", justifyContent: "center" }}>
              <MessagesSquare size={16} style={{ color: C.teal }} />
            </div>
            <h1 style={{ fontFamily: FH, fontWeight: 900, fontSize: "clamp(20px,2.6vw,26px)", color: "#fff", letterSpacing: "-.02em" }}>
              {t({ ru: "Сообщения", tg: "Паёмҳо" })}
            </h1>
          </div>
        </div>
      </div>

      <div style={{ maxWidth: 1100, margin: "0 auto", padding: "20px 28px 80px" }}>
      <div style={{ height: "calc(100vh - 340px)", minHeight: 440, display: "flex", borderRadius: 18, border: `1px solid ${C.border}`, overflow: "hidden" }}>
      <div style={{ width: 220, borderRight: `1px solid ${C.border}`, overflowY: "auto", flexShrink: 0, background: C.s1 }}>
        {conversations.length === 0 && (
          <p style={{ padding: 16, fontSize: 12.5, color: C.muted }}>{t({ ru: "Нет диалогов", tg: "Мукотиба нест" })}</p>
        )}
        {conversations.map((c) => (
          <button
            key={c.id}
            onClick={() => { setSelectedConv(c.id); router.replace(`/messages?conv=${c.id}`); }}
            style={{
              display: "flex", alignItems: "center", gap: 10, width: "100%", padding: "12px 14px",
              background: c.id === activeConv ? C.s3 : "transparent", border: "none", borderBottom: `1px solid ${C.border}`,
              cursor: "pointer", textAlign: "left",
            }}
          >
            <div style={{ width: 34, height: 34, borderRadius: "50%", background: C.teal, display: "flex", alignItems: "center", justifyContent: "center", fontFamily: FH, fontWeight: 800, color: C.overlay, fontSize: 13, flexShrink: 0 }}>
              {c.counterpart_type === "institution" ? "У" : "К"}
            </div>
            <span style={{ fontSize: 13, color: C.text, lineHeight: 1.3, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
              <ConversationLabel conv={c} />
            </span>
          </button>
        ))}
      </div>

      <div style={{ flex: 1, display: "flex", flexDirection: "column", minWidth: 0 }}>
        {conv ? (
          <>
            <div style={{ padding: "16px 22px", borderBottom: `1px solid ${C.border}`, display: "flex", alignItems: "center", gap: 12 }}>
              <div style={{ width: 38, height: 38, borderRadius: "50%", background: C.teal, display: "flex", alignItems: "center", justifyContent: "center", fontFamily: FH, fontWeight: 800, color: C.overlay, fontSize: 15 }}>
                {conv.counterpart_type === "institution" ? "У" : "К"}
              </div>
              <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 15, color: C.text }}><ConversationLabel conv={conv} /></p>
            </div>

            <div style={{ flex: 1, overflowY: "auto", padding: 22, display: "flex", flexDirection: "column", gap: 12 }}>
              {thread.length === 0 && (
                <div style={{ color: C.muted, fontSize: 13, textAlign: "center", marginTop: 30 }}>{t({ ru: "Начните переписку", tg: "Мукотибаро оғоз кунед" })}</div>
              )}
              {thread.map((m) => {
                const isMe = m.sender_type === "user" && m.sender_id === me.id;
                return (
                  <div key={m.id} style={{ alignSelf: isMe ? "flex-end" : "flex-start", maxWidth: "70%" }}>
                    <div style={{ background: isMe ? C.teal : C.s2, color: isMe ? C.overlay : C.text, padding: "10px 14px", borderRadius: 16, fontSize: 14, lineHeight: 1.5 }}>
                      {m.body}
                    </div>
                    <div style={{ fontSize: 11, color: C.dim, marginTop: 3, textAlign: isMe ? "right" : "left" }}>
                      {new Date(m.created_at).toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" })}
                    </div>
                  </div>
                );
              })}
            </div>

            <div style={{ display: "flex", gap: 10, padding: 16, borderTop: `1px solid ${C.border}` }}>
              <input
                autoFocus
                value={text}
                onChange={(e) => setText(e.target.value)}
                onKeyDown={(e) => { if (e.key === "Enter") send(); }}
                placeholder={t({ ru: "Сообщение…", tg: "Паём…" })}
                style={{ flex: 1, background: C.s2, border: `1px solid ${C.border}`, borderRadius: 12, padding: "12px 16px", color: C.text, fontSize: 14, fontFamily: FB, outline: "none" }}
              />
              <button onClick={send} aria-label={t({ ru: "Отправить", tg: "Фиристодан" })} style={{ background: C.teal, border: "none", borderRadius: 12, width: 46, flexShrink: 0, display: "flex", alignItems: "center", justifyContent: "center", cursor: "pointer" }}>
                <Send size={18} color={C.overlay} />
              </button>
            </div>
          </>
        ) : (
          <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: C.muted, fontSize: 13.5 }}>
            {t({ ru: "Выберите диалог", tg: "Мукотибаро интихоб кунед" })}
          </div>
        )}
      </div>
      </div>
      </div>
    </div>
  );
}
