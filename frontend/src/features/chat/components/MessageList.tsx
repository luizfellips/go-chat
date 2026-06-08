import { useEffect, useMemo, useRef } from "react";
import { useInfiniteQuery } from "@tanstack/react-query";
import { listMessages } from "@/services/message.service";
import { MessageBubble } from "./MessageBubble";
import { formatTime } from "@/lib/utils";
import { useAuthStore } from "@/store/auth.store";
import type { Message } from "@/types/message";

interface MessageListProps {
  conversationId: string;
  onMarkUnreadAsRead?: (messages: Message[]) => void;
}

export function MessageList({ conversationId, onMarkUnreadAsRead }: MessageListProps) {
  const userId = useAuthStore((s) => s.user?.id);
  const bottomRef = useRef<HTMLDivElement>(null);

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
    queryKey: ["messages", conversationId],
    queryFn: ({ pageParam }) => listMessages(conversationId, pageParam as string | undefined),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) => last.next_cursor,
  });

  const messages = useMemo(
    () => data?.pages.flatMap((p) => p.messages) ?? [],
    [data]
  );

  const sorted = useMemo(() => [...messages].reverse(), [messages]);

  const unreadKey = useMemo(
    () =>
      sorted
        .filter(
          (m) =>
            m.sender_id &&
            m.sender_id !== userId &&
            !m.read_at &&
            !m.id.startsWith("pending-")
        )
        .map((m) => m.id)
        .join(","),
    [sorted, userId]
  );

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [conversationId, sorted.length]);

  useEffect(() => {
    if (!onMarkUnreadAsRead || !unreadKey) return;
    const unread = sorted.filter(
      (m) =>
        m.sender_id &&
        m.sender_id !== userId &&
        !m.read_at &&
        !m.id.startsWith("pending-")
    );
    if (unread.length > 0) {
      onMarkUnreadAsRead(unread);
    }
  }, [unreadKey, onMarkUnreadAsRead, sorted, userId]);

  return (
    <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-2 py-3 sm:px-4 sm:py-4">
      {hasNextPage && (
        <button
          type="button"
          className="mb-4 w-full rounded-md py-2 text-sm text-primary hover:underline"
          onClick={() => fetchNextPage()}
          disabled={isFetchingNextPage}
        >
          {isFetchingNextPage ? "Loading..." : "Load older messages"}
        </button>
      )}
      <div className="flex flex-col gap-2 sm:gap-3">
        {sorted.map((msg) => (
          <MessageBubble
            key={msg.id}
            content={msg.content}
            time={formatTime(msg.created_at)}
            sent={msg.sender_id === userId}
            read={!!msg.read_at}
            pending={msg.id.startsWith("pending-")}
          />
        ))}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
