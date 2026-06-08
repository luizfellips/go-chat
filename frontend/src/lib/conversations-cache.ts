import type { QueryClient } from "@tanstack/react-query";
import type { Conversation } from "@/types/conversation";

export function hasConversationInCache(
  queryClient: QueryClient,
  conversationId: string
): boolean {
  const list = queryClient.getQueryData<Conversation[]>(["conversations"]);
  return list?.some((c) => c.id === conversationId) ?? false;
}

export function prependConversation(queryClient: QueryClient, conversation: Conversation) {
  queryClient.setQueryData<Conversation[]>(["conversations"], (old) => {
    const list = old ?? [];
    if (list.some((c) => c.id === conversation.id)) return list;
    return [conversation, ...list];
  });
}

export function setConversationUnreadCount(
  queryClient: QueryClient,
  conversationId: string,
  unreadCount: number
) {
  queryClient.setQueryData<Conversation[]>(["conversations"], (old) => {
    if (!old) return old;
    return old.map((c) =>
      c.id === conversationId ? { ...c, unread_count: Math.max(0, unreadCount) } : c
    );
  });
}

export function clearConversationUnread(queryClient: QueryClient, conversationId: string) {
  queryClient.setQueryData<Conversation[]>(["conversations"], (old) => {
    if (!old) return old;
    const target = old.find((c) => c.id === conversationId);
    if (!target || target.unread_count === 0) return old;
    return old.map((c) =>
      c.id === conversationId ? { ...c, unread_count: 0 } : c
    );
  });
}

export function decrementConversationUnread(queryClient: QueryClient, conversationId: string) {
  queryClient.setQueryData<Conversation[]>(["conversations"], (old) => {
    if (!old) return old;
    const target = old.find((c) => c.id === conversationId);
    if (!target || target.unread_count === 0) return old;
    return old.map((c) =>
      c.id === conversationId
        ? { ...c, unread_count: c.unread_count - 1 }
        : c
    );
  });
}

export function incrementConversationUnread(queryClient: QueryClient, conversationId: string) {
  queryClient.setQueryData<Conversation[]>(["conversations"], (old) => {
    if (!old) return old;
    return old.map((c) =>
      c.id === conversationId ? { ...c, unread_count: c.unread_count + 1 } : c
    );
  });
}

export function updateParticipantOnline(
  queryClient: QueryClient,
  userId: string,
  isOnline: boolean
) {
  queryClient.setQueryData<Conversation[]>(["conversations"], (old) => {
    if (!old) return old;
    return old.map((c) =>
      c.participant.id === userId
        ? { ...c, participant: { ...c.participant, is_online: isOnline } }
        : c
    );
  });
}

export function updateConversationLastMessage(
  queryClient: QueryClient,
  conversationId: string,
  content: string,
  createdAt: string,
  senderId?: string | null
) {
  const preview = content.length > 100 ? content.slice(0, 100) : content;
  queryClient.setQueryData<Conversation[]>(["conversations"], (old) => {
    if (!old) return old;
    const updated = old.map((c) =>
      c.id === conversationId
        ? {
            ...c,
            last_message: { content: preview, created_at: createdAt, sender_id: senderId ?? null },
          }
        : c
    );
    return [...updated].sort((a, b) => {
      const aTime = a.last_message?.created_at ?? "";
      const bTime = b.last_message?.created_at ?? "";
      return bTime.localeCompare(aTime);
    });
  });
}
