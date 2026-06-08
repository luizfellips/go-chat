import { useQuery, useQueryClient } from "@tanstack/react-query";
import { listConversations, createConversation } from "@/services/conversation.service";
import { searchUserByUsername } from "@/services/user.service";
import { prependConversation } from "@/lib/conversations-cache";
import { ConversationItem } from "./ConversationItem";
import { Input } from "@/components/ui/input";
import { useState } from "react";
import { Button } from "@/components/ui/button";

interface ConversationSidebarProps {
  selectedId?: string;
  onSelect: (id: string) => void;
}

export function ConversationSidebar({ selectedId, onSelect }: ConversationSidebarProps) {
  const [newUser, setNewUser] = useState("");
  const [error, setError] = useState("");
  const queryClient = useQueryClient();
  const { data: conversations = [], isLoading } = useQuery({
    queryKey: ["conversations"],
    queryFn: listConversations,
    refetchInterval: 30000,
  });

  const handleStartChat = async () => {
    if (!newUser.trim()) return;
    setError("");
    try {
      const user = await searchUserByUsername(newUser.trim());
      const conv = await createConversation(user.id);
      prependConversation(queryClient, {
        id: conv.id,
        participant: {
          id: user.id,
          username: user.username,
          is_online: false,
        },
        last_message: null,
        unread_count: 0,
      });
      onSelect(conv.id);
      setNewUser("");
      void queryClient.invalidateQueries({ queryKey: ["conversations"] });
    } catch {
      setError("User not found");
    }
  };

  return (
    <aside className="flex h-full min-h-0 w-full flex-col border-r border-border bg-sidebar">
      <div className="shrink-0 border-b border-border p-3 sm:p-4">
        <h2 className="mb-2 text-base font-semibold sm:mb-3 sm:text-lg">Messages</h2>
        <div className="flex gap-2">
          <Input
            placeholder="Username (e.g. bob)"
            value={newUser}
            onChange={(e) => setNewUser(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleStartChat()}
            className="min-w-0"
          />
          <Button size="sm" onClick={handleStartChat} className="shrink-0">
            New
          </Button>
        </div>
        {error && <p className="mt-2 text-xs text-red-500">{error}</p>}
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain p-1 sm:p-2">
        {isLoading && <p className="p-4 text-sm text-muted-foreground">Loading...</p>}
        {!isLoading && conversations.length === 0 && (
          <p className="p-4 text-sm text-muted-foreground">
            No conversations yet. Start one above.
          </p>
        )}
        {conversations.map((c) => (
          <ConversationItem
            key={c.id}
            conversation={c}
            active={c.id === selectedId}
            onClick={() => onSelect(c.id)}
          />
        ))}
      </div>
    </aside>
  );
}
