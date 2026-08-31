"use client";

import { Dialog as DialogPrimitive } from "radix-ui";
import { C } from "@/lib/data";

// Тонкая обёртка над Radix Dialog (тот же примитив, что использует shadcn/ui) —
// сохраняет прежний API (open/onClose/maxWidth) и пиксель-в-пиксель прежний вид,
// но добавляет настоящий focus-trap и aria-семантику вместо самодельного оверлея.
export function Modal({
  open,
  onClose,
  children,
  maxWidth = 560,
}: {
  open: boolean;
  onClose: () => void;
  children: React.ReactNode;
  maxWidth?: number;
}) {
  return (
    <DialogPrimitive.Root open={open} onOpenChange={(next) => { if (!next) onClose(); }}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay
          style={{
            position: "fixed",
            inset: 0,
            zIndex: 200,
            background: "rgba(4,12,13,0.72)",
            backdropFilter: "blur(4px)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            padding: 20,
            animation: "eh-fade .18s ease-out",
          }}
        >
          <DialogPrimitive.Content
            aria-describedby={undefined}
            style={{
              background: C.s2,
              border: `1px solid ${C.border}`,
              borderRadius: 20,
              maxWidth,
              width: "100%",
              maxHeight: "86vh",
              overflowY: "auto",
              boxShadow: "0 24px 64px rgba(0,0,0,0.5)",
              animation: "eh-pop .22s cubic-bezier(.16,1,.3,1)",
              outline: "none",
            }}
          >
            <DialogPrimitive.Title className="sr-only">Dialog</DialogPrimitive.Title>
            {children}
          </DialogPrimitive.Content>
        </DialogPrimitive.Overlay>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  );
}
