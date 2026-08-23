"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Send, MessagesSquare } from "lucide-react";
import { C, FH, FB, PHOTOS, CONVS, MESSAGES } from "@/lib/data";
import { useAppState } from "@/lib/app-state";
import { useT, useLocale } from "@/lib/i18n";

const LS_MESSAGES_KEY = "eduhub_chat_messages_v1";

function loadMessages(): typeof MESSAGES {
  if (typeof window === "undefined") return MESSAGES;
  try {
    const raw = localStorage.getItem(LS_MESSAGES_KEY);
    return raw ? (JSON.parse(raw) as typeof MESSAGES) : MESSAGES;
  } catch {
    return MESSAGES;
  }
}

export default function MessagesPage() {
  return (
    <Suspense fallback={null}>
      <MessagesInner />
    </Suspense>
  );
}

function MessagesInner() {
  const searchParams = useSearchParams();
  const t = useT();
  const { locale } = useLocale();
  const { setUnreadMessages } = useAppState();
  const initialConv = Number(searchParams.get("conv")) || CONVS[0]?.id || 1;

  const [activeConv, setActiveConv] = useState(initialConv);
  const [messages, setMessages] = useState(loadMessages);
  const [text, setText] = useState("");

  useEffect(() => {
    setUnreadMessages(0);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    localStorage.setItem(LS_MESSAGES_KEY, JSON.stringify(messages));
  }, [messages]);

  const thread = messages.filter((m) => m.convId === activeConv);
  const conv = CONVS.find((c) => c.id === activeConv);

  function send() {
    if (!text.trim()) return;
    setMessages((prev) => [
      ...prev,
      {
        id: prev.length + 1,
        convId: activeConv,
        from: "me" as const,
        text: text.trim(),
        time: new Date().toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" }),
      },
    ]);
    setText("");
  }

  return (
    <div style={{ fontFamily: FB }}>
      {/* ── тематический фото-фон над чатом ── */}
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
      {/* ── conversation list ── */}
      <div style={{ width: 220, borderRight: `1px solid ${C.border}`, overflowY: "auto", flexShrink: 0, background: C.s1 }}>
        {CONVS.map((c) => (
          <button
            key={c.id}
            onClick={() => setActiveConv(c.id)}
            style={{
              display: "flex", alignItems: "center", gap: 10, width: "100%", padding: "12px 14px",
              background: c.id === activeConv ? C.s3 : "transparent", border: "none", borderBottom: `1px solid ${C.border}`,
              cursor: "pointer", textAlign: "left",
            }}
          >
            <div style={{ width: 34, height: 34, borderRadius: "50%", background: c.color, display: "flex", alignItems: "center", justifyContent: "center", fontFamily: FH, fontWeight: 800, color: C.overlay, fontSize: 13, flexShrink: 0, position: "relative" }}>
              {c.name[locale][0]}
              {c.unread > 0 && (
                <span style={{ position: "absolute", top: -2, right: -2, width: 9, height: 9, borderRadius: "50%", background: C.red, border: `2px solid ${C.s1}` }} />
              )}
            </div>
            <span style={{ fontSize: 13, color: C.text, lineHeight: 1.3, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
              {c.name[locale]}
            </span>
          </button>
        ))}
      </div>

      {/* ── thread ── */}
      <div style={{ flex: 1, display: "flex", flexDirection: "column", minWidth: 0 }}>
        <div style={{ padding: "16px 22px", borderBottom: `1px solid ${C.border}`, display: "flex", alignItems: "center", gap: 12 }}>
          <div style={{ width: 38, height: 38, borderRadius: "50%", background: conv?.color ?? C.teal, display: "flex", alignItems: "center", justifyContent: "center", fontFamily: FH, fontWeight: 800, color: C.overlay, fontSize: 15 }}>
            {conv?.name[locale][0]}
          </div>
          <div>
            <p style={{ fontFamily: FH, fontWeight: 700, fontSize: 15, color: C.text }}>{conv?.name[locale]}</p>
            <p style={{ fontSize: 12, color: C.ok }}>● {t({ ru: "онлайн", tg: "онлайн" })}</p>
          </div>
        </div>

        <div style={{ flex: 1, overflowY: "auto", padding: 22, display: "flex", flexDirection: "column", gap: 12 }}>
          {thread.length === 0 && (
            <div style={{ color: C.muted, fontSize: 13, textAlign: "center", marginTop: 30 }}>{t({ ru: "Начните переписку", tg: "Мукотибаро оғоз кунед" })}</div>
          )}
          {thread.map((m) => (
            <div key={m.id} style={{ alignSelf: m.from === "me" ? "flex-end" : "flex-start", maxWidth: "70%" }}>
              <div style={{ background: m.from === "me" ? C.teal : C.s2, color: m.from === "me" ? C.overlay : C.text, padding: "10px 14px", borderRadius: 16, fontSize: 14, lineHeight: 1.5 }}>
                {m.text}
              </div>
              <div style={{ fontSize: 11, color: C.dim, marginTop: 3, textAlign: m.from === "me" ? "right" : "left" }}>{m.time}</div>
            </div>
          ))}
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
      </div>
      </div>
      </div>
    </div>
  );
}
