import { useCallback, useEffect, useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { ChatLayout } from "@/features/chat/components/ChatLayout";
import { ConversationSidebar } from "@/features/chat/components/ConversationSidebar";
import { MessageList } from "@/features/chat/components/MessageList";
import { MessageInput } from "@/features/chat/components/MessageInput";
import { UserStatus } from "@/features/chat/components/UserStatus";
import { Button } from "@/components/ui/button";
import { useWebSocket } from "@/hooks/useWebSocket";
import { listConversations } from "@/services/conversation.service";
import { useAuthStore } from "@/store/auth.store";
import { useParticipantOnline } from "@/hooks/useParticipantOnline";
import type { Message } from "@/types/message";
import { clearConversationUnread, updateConversationLastMessage } from "@/lib/conversations-cache";
import { setActiveConversationId } from "@/lib/active-conversation";
import { usePresenceStore } from "@/store/presence.store";

export function ChatPage() {
  const [selectedId, setSelectedId] = useState<string>();
  const userId = useAuthStore((s) => s.user?.id);
  const { send } = useWebSocket();
  const queryClient = useQueryClient();
  const markedReadRef = useRef<Set<string>>(new Set());

  const { data: conversations = [] } = useQuery({
    queryKey: ["conversations"],
    queryFn: listConversations,
  });

  const setOnline = usePresenceStore((s) => s.setOnline);

  useEffect(() => {
    for (const c of conversations) {
      setOnline(c.participant.id, c.participant.is_online);
    }
  }, [conversations, setOnline]);

  useEffect(() => {
    setActiveConversationId(selectedId ?? null);
    markedReadRef.current.clear();
    return () => setActiveConversationId(null);
  }, [selectedId]);

  const selected = conversations.find((c) => c.id === selectedId);
  const isOnline = useParticipantOnline(
    selected?.participant.id ?? "",
    selected?.participant.is_online ?? false
  );

  const handleSend = useCallback(
    (content: string) => {
      if (!selectedId || !userId) return;

      const optimistic: Message = {
        id: `pending-${Date.now()}`,
        sender_id: userId,
        content,
        created_at: new Date().toISOString(),
      };

      queryClient.setQueryData<{ pages: { messages: Message[] }[]; pageParams: unknown[] }>(
        ["messages", selectedId],
        (old) => {
          if (!old?.pages?.length) {
            return {
              pages: [{ messages: [optimistic] }],
              pageParams: [undefined],
            };
          }
          return {
            ...old,
            pages: old.pages.map((page, i) =>
              i === 0 ? { ...page, messages: [optimistic, ...page.messages] } : page
            ),
          };
        }
      );

      updateConversationLastMessage(
        queryClient,
        selectedId,
        content,
        optimistic.created_at,
        userId
      );

      send("message_sent", { conversation_id: selectedId, content });
    },
    [selectedId, userId, send, queryClient]
  );

  const handleMarkUnreadAsRead = useCallback(
    (messages: Message[]) => {
      if (!selectedId || !userId) return;

      const unread = messages.filter((m) => !markedReadRef.current.has(m.id));
      if (unread.length === 0) return;

      unread.forEach((m) => markedReadRef.current.add(m.id));
      clearConversationUnread(queryClient, selectedId);

      unread.forEach((m) => {
        send("message_read", {
          conversation_id: selectedId,
          message_id: m.id,
        });
      });
    },
    [selectedId, userId, send, queryClient]
  );

  return (
    <ChatLayout
      showMobileChat={!!selectedId}
      sidebar={
        <ConversationSidebar
          selectedId={selectedId}
          onSelect={setSelectedId}
        />
      }
      main={
        selected ? (
          <>
            <div className="flex shrink-0 items-center gap-2 border-b border-border px-2 py-2 sm:gap-3 sm:px-4 sm:py-3">
              <Button
                variant="ghost"
                size="icon"
                className="shrink-0 md:hidden"
                onClick={() => setSelectedId(undefined)}
                aria-label="Back to conversations"
              >
                <ArrowLeft className="h-5 w-5" />
              </Button>
              <UserStatus isOnline={isOnline} />
              <h2 className="truncate font-semibold">{selected.participant.username}</h2>
            </div>
            <MessageList
              conversationId={selected.id}
              onMarkUnreadAsRead={handleMarkUnreadAsRead}
            />
            <MessageInput onSend={handleSend} />
          </>
        ) : (
          <div className="flex flex-1 items-center justify-center p-6 text-center text-sm text-muted-foreground sm:text-base">
            Select a conversation to start chatting
          </div>
        )
      }
    />
  );
}
