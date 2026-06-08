import { useState, KeyboardEvent } from "react";
import { Send } from "lucide-react";
import { Button } from "@/components/ui/button";

interface MessageInputProps {
  onSend: (content: string) => void;
  disabled?: boolean;
}

export function MessageInput({ onSend, disabled }: MessageInputProps) {
  const [content, setContent] = useState("");

  const handleSend = () => {
    const trimmed = content.trim();
    if (!trimmed) return;
    onSend(trimmed);
    setContent("");
  };

  const onKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="safe-bottom flex shrink-0 items-end gap-2 border-t border-border bg-card p-2 sm:p-4">
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        onKeyDown={onKeyDown}
        placeholder="Type a message..."
        disabled={disabled}
        rows={1}
        className="max-h-32 min-h-[44px] flex-1 resize-none rounded-md border border-border bg-background px-3 py-2.5 text-base sm:text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
      />
      <Button
        onClick={handleSend}
        disabled={disabled || !content.trim()}
        className="h-11 shrink-0 px-3 sm:px-4"
        aria-label="Send message"
      >
        <Send className="h-4 w-4 sm:mr-2" />
        <span className="hidden sm:inline">Send</span>
      </Button>
    </div>
  );
}
