import type { Conversation } from "@/types/conversation";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { UserStatus } from "./UserStatus";
import { cn, formatTime } from "@/lib/utils";
import { useAuthStore } from "@/store/auth.store";
import { useParticipantOnline } from "@/hooks/useParticipantOnline";

interface ConversationItemProps {
  conversation: Conversation;
  active: boolean;
  onClick: () => void;
}

function formatLastMessagePreview(
  lastMessage: Conversation["last_message"],
  currentUserId?: string
): string {
  if (!lastMessage?.content) return "No messages yet";
  if (lastMessage.sender_id && lastMessage.sender_id === currentUserId) {
    return `You: ${lastMessage.content}`;
  }
  return lastMessage.content;
}

export function ConversationItem({ conversation, active, onClick }: ConversationItemProps) {
  const userId = useAuthStore((s) => s.user?.id);
  const isOnline = useParticipantOnline(
    conversation.participant.id,
    conversation.participant.is_online
  );
  const initial = conversation.participant.username[0]?.toUpperCase() ?? "?";
  const preview = formatLastMessagePreview(conversation.last_message, userId);

  return (
    <button
      onClick={onClick}
      className={cn(
        "flex w-full items-center gap-3 rounded-lg px-2 py-3 text-left transition-colors active:scale-[0.99] hover:bg-muted sm:px-3",
        active && "bg-muted"
      )}
    >
      <div className="relative">
        <Avatar>
          <AvatarFallback>{initial}</AvatarFallback>
        </Avatar>
        <UserStatus isOnline={isOnline} className="absolute -bottom-0.5 -right-0.5" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center justify-between gap-2">
          <span className="truncate font-medium">{conversation.participant.username}</span>
          {conversation.last_message?.created_at && (
            <span className="shrink-0 text-xs text-muted-foreground">
              {formatTime(conversation.last_message.created_at)}
            </span>
          )}
        </div>
        <div className="flex items-center justify-between gap-2">
          <p className="truncate text-sm text-muted-foreground">{preview}</p>
          {conversation.unread_count > 0 && !active && (
            <span className="shrink-0 rounded-full bg-primary px-2 py-0.5 text-xs text-primary-foreground">
              {conversation.unread_count}
            </span>
          )}
        </div>
      </div>
    </button>
  );
}
