interface MessageBubbleProps {
  content: string;
  time: string;
  sent: boolean;
  read?: boolean;
  pending?: boolean;
}

export function MessageBubble({ content, time, sent, read, pending }: MessageBubbleProps) {
  return (
    <div className={`flex ${sent ? "justify-end" : "justify-start"}`}>
      <div
        className={`max-w-[min(85vw,32rem)] rounded-2xl px-3 py-2 sm:max-w-[75%] sm:px-4 ${
          sent
            ? `rounded-br-sm bg-primary text-primary-foreground ${pending ? "opacity-80" : ""}`
            : "rounded-bl-sm bg-muted"
        }`}
      >
        <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">{content}</p>
        <div className="mt-1 flex items-center justify-end gap-1 text-[10px] opacity-70">
          <span>{time}</span>
          {sent && <span>{pending ? "…" : read ? "✓✓" : "✓"}</span>}
        </div>
      </div>
    </div>
  );
}
