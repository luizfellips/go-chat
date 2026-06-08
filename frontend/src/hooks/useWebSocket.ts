import { useEffect, useRef, useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { api } from "@/services/api";
import { getAccessToken } from "@/lib/token";
import {
  decrementConversationUnread,
  hasConversationInCache,
  incrementConversationUnread,
  updateConversationLastMessage,
  updateParticipantOnline,
} from "@/lib/conversations-cache";
import { getActiveConversationId } from "@/lib/active-conversation";
import { useAuthStore } from "@/store/auth.store";
import { usePresenceStore } from "@/store/presence.store";
import { useWebSocketStore } from "@/store/websocket.store";
import type {
  WSEnvelope,
  WSMessageReadPayload,
  WSMessageReceivedPayload,
  WSUserPresencePayload,
} from "@/types/websocket";
import type { Message } from "@/types/message";

const WS_URL = import.meta.env.VITE_WS_URL || "ws://localhost:8080/ws/connect";

async function fetchWSTicket(): Promise<string> {
  const { data } = await api.post<{ ticket: string }>("/ws/ticket");
  return data.ticket;
}

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const accessToken = useAuthStore((s) => s.accessToken);
  const setStatus = useWebSocketStore((s) => s.setStatus);
  const setOnline = usePresenceStore((s) => s.setOnline);
  const queryClient = useQueryClient();

  const connect = useCallback(async () => {
    const token = accessToken || getAccessToken();
    if (!token) return;

    setStatus("connecting");
    try {
      const ticket = await fetchWSTicket();
      const ws = new WebSocket(`${WS_URL}?ticket=${encodeURIComponent(ticket)}`);
      wsRef.current = ws;

      ws.onopen = () => setStatus("connected");

      ws.onmessage = (event) => {
        const env: WSEnvelope = JSON.parse(event.data);
        switch (env.type) {
          case "user_online": {
            const p = env.payload as WSUserPresencePayload;
            setOnline(p.user_id, true);
            updateParticipantOnline(queryClient, p.user_id, true);
            break;
          }
          case "user_offline": {
            const p = env.payload as WSUserPresencePayload;
            setOnline(p.user_id, false);
            updateParticipantOnline(queryClient, p.user_id, false);
            break;
          }
          case "message_received": {
            const p = env.payload as WSMessageReceivedPayload;
            const msg = p.message;
            const currentUserId = useAuthStore.getState().user?.id;
            queryClient.setQueryData<{ pages: { messages: Message[] }[]; pageParams: unknown[] }>(
              ["messages", msg.conversation_id],
              (old) => {
                if (!old?.pages?.length) {
                  return {
                    pages: [{ messages: [msg] }],
                    pageParams: [undefined],
                  };
                }
                const exists = old.pages.some((page) =>
                  page.messages.some((m) => m.id === msg.id)
                );
                if (exists) return old;
                const newPages = [...old.pages];
                newPages[0] = {
                  ...newPages[0],
                  messages: [
                    msg,
                    ...newPages[0].messages.filter(
                      (m) =>
                        m.id !== msg.id &&
                        !(
                          m.id.startsWith("pending-") &&
                          m.content === msg.content &&
                          m.sender_id === msg.sender_id
                        )
                    ),
                  ],
                };
                return { ...old, pages: newPages };
              }
            );
            if (hasConversationInCache(queryClient, msg.conversation_id)) {
              updateConversationLastMessage(
                queryClient,
                msg.conversation_id,
                msg.content,
                msg.created_at,
                msg.sender_id
              );
              if (
                msg.sender_id &&
                msg.sender_id !== currentUserId &&
                msg.conversation_id !== getActiveConversationId()
              ) {
                incrementConversationUnread(queryClient, msg.conversation_id);
              }
            } else {
              void queryClient.refetchQueries({ queryKey: ["conversations"] });
            }
            break;
          }
          case "message_read": {
            const p = env.payload as WSMessageReadPayload;
            queryClient.setQueryData<{ pages: { messages: Message[] }[] }>(
              ["messages", p.conversation_id],
              (old) => {
                if (!old) return old;
                return {
                  ...old,
                  pages: old.pages.map((page) => ({
                    ...page,
                    messages: page.messages.map((m) =>
                      m.id === p.message_id ? { ...m, read_at: p.read_at } : m
                    ),
                  })),
                };
              }
            );
            decrementConversationUnread(queryClient, p.conversation_id);
            break;
          }
        }
      };

      ws.onclose = () => {
        setStatus("disconnected");
        reconnectRef.current = setTimeout(() => {
          void connect();
        }, 3000);
      };

      ws.onerror = () => ws.close();
    } catch {
      setStatus("disconnected");
      reconnectRef.current = setTimeout(() => {
        void connect();
      }, 3000);
    }
  }, [accessToken, setStatus, setOnline, queryClient]);

  useEffect(() => {
    void connect();
    return () => {
      if (reconnectRef.current) clearTimeout(reconnectRef.current);
      wsRef.current?.close();
    };
  }, [connect]);

  const send = useCallback((type: string, payload: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(
        JSON.stringify({ type, payload, timestamp: new Date().toISOString() })
      );
    }
  }, []);

  return { send };
}
