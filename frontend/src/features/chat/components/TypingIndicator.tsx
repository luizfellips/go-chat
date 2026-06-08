interface TypingIndicatorProps {
  username?: string;
}

export function TypingIndicator({ username }: TypingIndicatorProps) {
  if (!username) return null;
  return (
    <p className="px-4 py-1 text-xs italic text-muted-foreground">
      {username} is typing...
    </p>
  );
}
